package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
)

type AgentCardRequesterInterface interface {
	GetCard(ctx context.Context, agentUrl string) (*models.AgentCard, error)
}
