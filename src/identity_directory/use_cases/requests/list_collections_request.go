package requests

import "github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"

type ListCollectionsRequest struct {
	PaginateRequest
	ChainId uint64 `query:"chainId" json:"-"`
}

func (r *ListCollectionsRequest) ValidateAndSanitize() error {
	r.PaginateRequest.Sanitize()
	return nil
}

func (r *ListCollectionsRequest) ToParams() *repositories.AgentCollectionListParams {
	return &repositories.AgentCollectionListParams{
		ChainId: r.ChainId,
		Offset:  r.Offset(),
		Limit:   int(*r.ElementsPerPage),
	}
}
