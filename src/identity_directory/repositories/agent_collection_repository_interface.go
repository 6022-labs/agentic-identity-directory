package repositories

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
)

type AgentCollectionListParams struct {
	ChainId uint64
	Offset  int
	Limit   int
}

type AgentCollectionRepositoryInterface interface {
	Upsert(ctx context.Context, agentCollection *models.AgentCollection) error
	// Get returns nil without an error when the collection is not indexed.
	Get(ctx context.Context, chainId uint64, address string) (*models.AgentCollection, error)
	// ListAddresses returns every collection of the chain.
	ListAddresses(ctx context.Context, chainId uint64) ([]string, error)
	// List returns one page and the total across pages.
	List(ctx context.Context, params *AgentCollectionListParams) ([]*models.AgentCollection, int64, error)
}
