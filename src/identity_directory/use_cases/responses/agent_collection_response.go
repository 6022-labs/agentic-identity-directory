package responses

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"

type AgentCollectionResponse struct {
	ChainId     uint64 `json:"chainId"`
	Address     string `json:"address"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
}

func AgentCollectionResponseFromModel(collection *models.AgentCollection) *AgentCollectionResponse {
	return &AgentCollectionResponse{
		ChainId:     collection.ChainId,
		Address:     collection.Address,
		Name:        collection.Name,
		Symbol:      collection.Symbol,
		Description: collection.Description,
	}
}
