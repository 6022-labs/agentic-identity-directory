package repositories

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
)

type AgentSearchParams struct {
	Query             string
	Role              string
	ChainId           uint64
	CollectionAddress string
	Owner             string
	Address           string
	Published         *bool
	Offset            int
	Limit             int
}

// AgentRepositoryInterface reads return nil without an error when no agent matches.
type AgentRepositoryInterface interface {
	// Upsert replaces the agent and its addresses, its ENS records are left alone.
	Upsert(ctx context.Context, agent *models.Agent) error
	Get(ctx context.Context, chainId uint64, agentRef models.AgentRef) (*models.Agent, error)
	GetByEnsName(ctx context.Context, ensName string) (*models.Agent, error)
	// Search returns one page and the total across pages.
	Search(ctx context.Context, params *AgentSearchParams) ([]*models.Agent, int64, error)
}
