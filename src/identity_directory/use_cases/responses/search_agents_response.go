package responses

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"

type SearchAgentsResponse struct {
	PaginateResponse
	Items []*AgentResponse `json:"items"`
}

func SearchAgentsResponseFromModels(agents []*models.Agent, paginateResponse PaginateResponse) *SearchAgentsResponse {
	items := make([]*AgentResponse, 0, len(agents))
	for _, agent := range agents {
		items = append(items, AgentResponseFromModel(agent))
	}
	return &SearchAgentsResponse{PaginateResponse: paginateResponse, Items: items}
}
