package repositories

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
)

type AgentEnsRecordRepositoryInterface interface {
	// Set stores the record, or deletes it when its value is empty.
	Set(ctx context.Context, agentEnsRecord *models.AgentEnsRecord) error
}
