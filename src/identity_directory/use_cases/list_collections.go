package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/responses"
)

type ListCollections struct {
	agentCollectionRepository repositories.AgentCollectionRepositoryInterface
}

func NewListCollections(agentCollectionRepository repositories.AgentCollectionRepositoryInterface) *ListCollections {
	return &ListCollections{agentCollectionRepository: agentCollectionRepository}
}

func (u *ListCollections) Execute(
	ctx context.Context,
	request *requests.ListCollectionsRequest,
) (*responses.ListCollectionsResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	collections, total, err := u.agentCollectionRepository.List(ctx, request.ToParams())
	if err != nil {
		return nil, errors.NewInternalError("list_collections_failed", err)
	}
	return responses.ListCollectionsResponseFromModels(collections, responses.PaginateResponse{
		Page:            *request.Page,
		ElementsPerPage: *request.ElementsPerPage,
		TotalElements:   uint(total),
	}), nil
}
