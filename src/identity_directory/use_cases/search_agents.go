package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/responses"
)

type SearchAgents struct {
	agentRepository repositories.AgentRepositoryInterface
}

func NewSearchAgents(agentRepository repositories.AgentRepositoryInterface) *SearchAgents {
	return &SearchAgents{agentRepository: agentRepository}
}

func (u *SearchAgents) Execute(
	ctx context.Context,
	request *requests.SearchAgentsRequest,
) (*responses.SearchAgentsResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agents, total, err := u.agentRepository.Search(ctx, request.ToParams())
	if err != nil {
		return nil, errors.NewInternalError("search_agents_failed", err)
	}
	return responses.SearchAgentsResponseFromModels(agents, responses.PaginateResponse{
		Page:            *request.Page,
		ElementsPerPage: *request.ElementsPerPage,
		TotalElements:   uint(total),
	}), nil
}
