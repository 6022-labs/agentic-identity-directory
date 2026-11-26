package interfaces

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"

type AgentCardSignatureVerifierInterface interface {
	// Verify reports whether one of the signers signed the card; an unsigned card is false without an error.
	Verify(agentCard *models.AgentCard, signers []string) (bool, error)
}
