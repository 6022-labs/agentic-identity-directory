package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
)

// AgentCollectionRequesterInterface state reads fail with ErrContractUnreadable on a broken contract.
type AgentCollectionRequesterInterface interface {
	// ListChanges returns what the block range touched in these collections, without duplicates.
	ListChanges(
		ctx context.Context,
		chainId uint64,
		collectionAddresses []string,
		fromBlock, toBlock uint64,
	) (*models.CollectionChanges, error)
	GetCollection(
		ctx context.Context,
		chainId uint64,
		collectionAddress string,
		atBlock uint64,
	) (*models.AgentCollection, error)
	ListAgentRefs(
		ctx context.Context,
		chainId uint64,
		collectionAddress string,
		atBlock uint64,
	) ([]models.AgentRef, error)
	// GetAgent leaves EnsName empty: naming is the domain's.
	GetAgent(ctx context.Context, chainId uint64, agentRef models.AgentRef, atBlock uint64) (*models.Agent, error)
}
