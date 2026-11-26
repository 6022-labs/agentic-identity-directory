package interfaces

import "context"

type BlockchainRequesterInterface interface {
	GetBlockNumber(ctx context.Context, chainId uint64) (uint64, error)
}
