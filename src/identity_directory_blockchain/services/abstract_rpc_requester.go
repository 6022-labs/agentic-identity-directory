package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"

	common_errors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/factory"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/helpers"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	"github.com/avast/retry-go/v5"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// EIP-1474 "limit exceeded", used for rate limits.
const jsonRpcLimitExceededCode = -32005

const (
	rpcRetryAttempts = 8
	rpcRetryDelay    = time.Second
	rpcRetryMaxDelay = 30 * time.Second
)

// AbstractRpcRequester retries transient RPC failures and splits log queries a node refuses as too large.
type AbstractRpcRequester struct {
	logger           *zap.Logger
	ethClientFactory *factory.EthClientFactory
	chainsSettings   *settings.ChainsSettings
}

// filterLogs keeps the node's order: by block, then log index.
func (r *AbstractRpcRequester) filterLogs(
	ctx context.Context,
	chainId uint64,
	addresses []common.Address,
	topics []common.Hash,
	fromBlock, toBlock uint64,
) ([]types.Log, error) {
	// An empty address list would match every contract on the chain.
	if len(addresses) == 0 {
		return nil, nil
	}
	chain := r.chainsSettings.Get(chainId)
	if chain == nil {
		return nil, fmt.Errorf("no RPC configured for chain %d", chainId)
	}

	var logs []types.Log
	for _, batch := range lo.Chunk(addresses, chain.MaxAddressesPerLogQuery) {
		batchLogs, err := r.filterLogsInRange(ctx, chainId, batch, topics, fromBlock, toBlock)
		if err != nil {
			return nil, err
		}
		logs = append(logs, batchLogs...)
	}
	return logs, nil
}

func (r *AbstractRpcRequester) filterLogsInRange(
	ctx context.Context,
	chainId uint64,
	addresses []common.Address,
	topics []common.Hash,
	fromBlock, toBlock uint64,
) ([]types.Log, error) {
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return nil, err
	}

	var logs []types.Log
	err = r.retryRpc(ctx, func() error {
		logs, err = client.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: new(big.Int).SetUint64(fromBlock),
			ToBlock:   new(big.Int).SetUint64(toBlock),
			Addresses: addresses,
			Topics:    [][]common.Hash{topics},
		})
		return err
	}, r.isRetryableLogQueryError)
	if err == nil {
		return logs, nil
	}
	if toBlock == fromBlock || !r.isRangeError(err) {
		return nil, fmt.Errorf("filter logs of blocks %d-%d: %w", fromBlock, toBlock, err)
	}

	middle := fromBlock + (toBlock-fromBlock)/2
	firstHalf, err := r.filterLogsInRange(ctx, chainId, addresses, topics, fromBlock, middle)
	if err != nil {
		return nil, err
	}
	secondHalf, err := r.filterLogsInRange(ctx, chainId, addresses, topics, middle+1, toBlock)
	if err != nil {
		return nil, err
	}
	return append(firstHalf, secondHalf...), nil
}

// callContract marks a revert, a failed execution or an undecodable answer as ErrContractUnreadable.
func (r *AbstractRpcRequester) callContract(ctx context.Context, call func() error) error {
	err := r.retryRpc(ctx, call, r.isRetryableRpcError)
	if err == nil {
		return nil
	}
	// No code at all is a node behind the chain: the manager only registers contracts.
	if r.isExecutionFailure(err) || helpers.IsRevert(err) || strings.HasPrefix(err.Error(), "abi:") {
		return fmt.Errorf("%w: %w", common_errors.ErrContractUnreadable, err)
	}
	return err
}

func (r *AbstractRpcRequester) retryRpc(ctx context.Context, call func() error, retryable func(error) bool) error {
	return retry.New(
		retry.Context(ctx),
		retry.Attempts(rpcRetryAttempts),
		retry.Delay(rpcRetryDelay),
		retry.MaxDelay(rpcRetryMaxDelay),
		retry.DelayType(retry.CombineDelay(retry.BackOffDelay, retry.RandomDelay)),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(attempt uint, err error) {
			r.logger.Warn("rpc call failed, retrying", zap.Uint("attempt", attempt+1), zap.Error(err))
		}),
	).Do(func() error {
		err := call()
		if err != nil && !retryable(err) {
			return retry.Unrecoverable(err)
		}
		return err
	})
}

func (r *AbstractRpcRequester) isRetryableRpcError(err error) bool {
	var httpErr rpc.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == http.StatusTooManyRequests || httpErr.StatusCode >= http.StatusInternalServerError
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	var rpcErr rpc.Error
	if errors.As(err, &rpcErr) && rpcErr.ErrorCode() == jsonRpcLimitExceededCode {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	message := strings.ToLower(err.Error())
	// A load-balanced node behind the others does not know the block yet.
	return strings.Contains(message, "rate limit") || strings.Contains(message, "too many requests") ||
		strings.Contains(message, "header not found") || strings.Contains(message, "unknown block")
}

// isRetryableLogQueryError leaves timeouts, gateway errors and result caps to the range split, which fixes them.
func (r *AbstractRpcRequester) isRetryableLogQueryError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return false
	}
	if r.isServerError(err) {
		return false
	}
	return !r.isResultLimitError(err) && r.isRetryableRpcError(err)
}

// isRangeError is a refusal a smaller range may avoid: a node error, a refused request or a timeout.
func (r *AbstractRpcRequester) isRangeError(err error) bool {
	if r.isResultLimitError(err) {
		return true
	}
	var rpcErr rpc.Error
	if errors.As(err, &rpcErr) && rpcErr.ErrorCode() != jsonRpcLimitExceededCode {
		return true
	}
	var httpErr rpc.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == http.StatusBadRequest || httpErr.StatusCode == http.StatusRequestEntityTooLarge ||
			r.isServerError(err)
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// isServerError is a 5xx: on a log query, usually a range too heavy for the node's gateway.
func (r *AbstractRpcRequester) isServerError(err error) bool {
	var httpErr rpc.HTTPError
	return errors.As(err, &httpErr) && httpErr.StatusCode >= http.StatusInternalServerError
}

// isResultLimitError is a cap on the logs one query returns, which some nodes report with the rate-limit code.
func (r *AbstractRpcRequester) isResultLimitError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "returned more than") || strings.Contains(message, "block range")
}

// isExecutionFailure is the contract's own code failing in the node, unlike a quota or node outage.
func (r *AbstractRpcRequester) isExecutionFailure(err error) bool {
	var rpcErr rpc.Error
	if !errors.As(err, &rpcErr) {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "out of gas") || strings.Contains(message, "gas required exceeds") ||
		strings.Contains(message, "invalid opcode") || strings.Contains(message, "stack overflow") ||
		strings.Contains(message, "stack underflow")
}
