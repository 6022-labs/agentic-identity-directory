package responses

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"

type ListCollectionsResponse struct {
	PaginateResponse
	Items []*AgentCollectionResponse `json:"items"`
}

func ListCollectionsResponseFromModels(
	collections []*models.AgentCollection,
	paginateResponse PaginateResponse,
) *ListCollectionsResponse {
	items := make([]*AgentCollectionResponse, 0, len(collections))
	for _, collection := range collections {
		items = append(items, AgentCollectionResponseFromModel(collection))
	}
	return &ListCollectionsResponse{PaginateResponse: paginateResponse, Items: items}
}
