package services

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/factory"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	"go.uber.org/zap"
)

type BlockchainRequester struct {
	AbstractRpcRequester
}

func NewBlockchainRequester(
	logger *zap.Logger,
	ethClientFactory *factory.EthClientFactory,
	chainsSettings *settings.ChainsSettings,
) *BlockchainRequester {
	return &BlockchainRequester{
		AbstractRpcRequester: AbstractRpcRequester{
			logger:           logger,
			ethClientFactory: ethClientFactory,
			chainsSettings:   chainsSettings,
		},
	}
}

func (r *BlockchainRequester) GetBlockNumber(ctx context.Context, chainId uint64) (uint64, error) {
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return 0, err
	}
	var blockNumber uint64
	err = r.retryRpc(ctx, func() error {
		blockNumber, err = client.BlockNumber(ctx)
		return err
	}, r.isRetryableRpcError)
	return blockNumber, err
}
