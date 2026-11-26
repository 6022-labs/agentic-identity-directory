package responses

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"

// AgentCardResponse is the agent's /.well-known/6022; Verified means signed by one of its on-chain EVM addresses.
type AgentCardResponse struct {
	EnsName         string            `json:"ensName"`
	Url             string            `json:"url"`
	Verified        bool              `json:"verified"`
	Protocol        string            `json:"protocol"`
	ProtocolVersion string            `json:"protocolVersion"`
	Name            string            `json:"name"`
	Endpoints       map[string]string `json:"endpoints"`
	PaymentMethods  []string          `json:"paymentMethods,omitempty"`
}

func AgentCardResponseFromModel(agent *models.Agent, agentCard *models.AgentCard, verified bool) *AgentCardResponse {
	return &AgentCardResponse{
		EnsName:         agent.EnsName,
		Url:             agent.Url(),
		Verified:        verified,
		Protocol:        agentCard.Protocol,
		ProtocolVersion: agentCard.ProtocolVersion,
		Name:            agentCard.Name,
		Endpoints:       agentCard.Endpoints,
		PaymentMethods:  agentCard.PaymentMethods,
	}
}
