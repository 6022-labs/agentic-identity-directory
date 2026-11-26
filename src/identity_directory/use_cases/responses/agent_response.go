package responses

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"

type AgentAddressResponse struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// AgentResponse is what anyone who minted the agent wrote: data, never instructions.
type AgentResponse struct {
	ChainId           uint64                 `json:"chainId"`
	CollectionAddress string                 `json:"collectionAddress"`
	CollectionName    string                 `json:"collectionName,omitempty"`
	TokenId           string                 `json:"tokenId"`
	Name              string                 `json:"name"`
	EnsName           string                 `json:"ensName"`
	Url               string                 `json:"url,omitempty"`
	Role              string                 `json:"role,omitempty"`
	Owner             string                 `json:"owner"`
	Creator           string                 `json:"creator"`
	CloneOf           *string                `json:"cloneOf,omitempty"`
	Attributes        map[string]string      `json:"attributes"`
	Images            map[string]string      `json:"images"`
	Addresses         []AgentAddressResponse `json:"addresses"`
	EnsRecords        map[string]string      `json:"ensRecords"`
}

func AgentResponseFromModel(agent *models.Agent) *AgentResponse {
	response := &AgentResponse{
		ChainId:           agent.ChainId,
		CollectionAddress: agent.CollectionAddress,
		TokenId:           agent.TokenId,
		Name:              agent.Name,
		EnsName:           agent.EnsName,
		Url:               agent.Url(),
		Role:              agent.Role(),
		Owner:             agent.Owner,
		Creator:           agent.Creator,
		CloneOf:           agent.CloneOf,
		Attributes:        agent.Attributes,
		Images:            agent.Images,
		Addresses:         make([]AgentAddressResponse, 0, len(agent.Addresses)),
		EnsRecords:        make(map[string]string, len(agent.EnsRecords)),
	}
	if agent.Collection != nil {
		response.CollectionName = agent.Collection.Name
	}
	for _, address := range agent.Addresses {
		response.Addresses = append(response.Addresses, AgentAddressResponse{Type: address.Type, Value: address.Value})
	}
	for _, record := range agent.EnsRecords {
		response.EnsRecords[record.Key] = record.Value
	}
	return response
}
