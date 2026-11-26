package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
)

type AgentEnsRegistryRequesterInterface interface {
	// ListTextChanges returns the text records set in the block range in log order; a cleared one is empty.
	ListTextChanges(ctx context.Context, chainId, fromBlock, toBlock uint64) ([]*models.AgentEnsRecord, error)
}
