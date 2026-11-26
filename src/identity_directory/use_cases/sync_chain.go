package use_cases

import (
	"context"
	"errors"
	"fmt"
	"time"

	common_errors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	metrics_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory/metrics/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
)

// SyncChain indexes a chain window by window up to its confirmed head; catch-up and live are the same pass.
type SyncChain struct {
	syncSettings         *settings.SyncSettings
	blockchainRequester  interfaces.BlockchainRequesterInterface
	syncCursorRepository repositories.SyncCursorRepositoryInterface
	chainWindowReader    interfaces.ChainWindowReaderInterface
	chainWindowWriter    interfaces.ChainWindowWriterInterface
	syncMetrics          metrics_interfaces.SyncMetricsInterface
}

func NewSyncChain(
	syncSettings *settings.SyncSettings,
	blockchainRequester interfaces.BlockchainRequesterInterface,
	syncCursorRepository repositories.SyncCursorRepositoryInterface,
	chainWindowReader interfaces.ChainWindowReaderInterface,
	chainWindowWriter interfaces.ChainWindowWriterInterface,
	syncMetrics metrics_interfaces.SyncMetricsInterface,
) *SyncChain {
	return &SyncChain{
		syncSettings:         syncSettings,
		blockchainRequester:  blockchainRequester,
		syncCursorRepository: syncCursorRepository,
		chainWindowReader:    chainWindowReader,
		chainWindowWriter:    chainWindowWriter,
		syncMetrics:          syncMetrics,
	}
}

func (u *SyncChain) Execute(ctx context.Context, request *requests.SyncChainRequest) error {
	chainSyncSettings := u.syncSettings.Get(request.ChainId)
	if chainSyncSettings == nil {
		return common_errors.NewValidationError("chainId", "is not a configured chain")
	}

	start := time.Now()
	err := u.syncUpToHead(ctx, request.ChainId, chainSyncSettings)
	u.syncMetrics.RecordPass(ctx, request.ChainId, time.Since(start), err)
	return err
}

func (u *SyncChain) syncUpToHead(
	ctx context.Context,
	chainId uint64,
	chainSyncSettings *settings.ChainSyncSettings,
) error {
	for ctx.Err() == nil {
		blockNumber, err := u.blockchainRequester.GetBlockNumber(ctx, chainId)
		if err != nil {
			return fmt.Errorf("read head block: %w", err)
		}
		if blockNumber < chainSyncSettings.Confirmations {
			return nil
		}
		headBlock := blockNumber - chainSyncSettings.Confirmations

		cursor, err := u.syncCursorRepository.Get(ctx, chainId)
		if err != nil {
			return fmt.Errorf("read sync cursor: %w", err)
		}
		fromBlock := chainSyncSettings.StartBlock
		if cursor != nil {
			fromBlock = *cursor + 1
		}
		u.syncMetrics.RecordPosition(ctx, chainId, fromBlock-1, headBlock)
		if fromBlock > headBlock {
			return nil
		}
		toBlock := min(fromBlock+chainSyncSettings.LogWindowBlocks-1, headBlock)

		chainWindow, err := u.chainWindowReader.Read(ctx, chainId, fromBlock, toBlock, headBlock)
		if err != nil {
			return fmt.Errorf("read blocks %d-%d: %w", fromBlock, toBlock, err)
		}
		err = u.chainWindowWriter.Write(ctx, chainWindow, cursor)
		if errors.Is(err, common_errors.ErrSyncCursorMoved) {
			// Another replica indexed this window first.
			return nil
		}
		if err != nil {
			return fmt.Errorf("write blocks %d-%d: %w", fromBlock, toBlock, err)
		}
		u.syncMetrics.RecordPosition(ctx, chainId, toBlock, headBlock)
	}
	return ctx.Err()
}
