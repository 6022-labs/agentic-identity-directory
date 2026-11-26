package interfaces

import "context"

type AgentCollectionsManagerRequesterInterface interface {
	// ListCreatedCollections returns the addresses registered in the block range, both ends included.
	ListCreatedCollections(ctx context.Context, chainId, fromBlock, toBlock uint64) ([]string, error)
}
