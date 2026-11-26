package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/responses"
)

type GetAgentByEnsName struct {
	agentRepository repositories.AgentRepositoryInterface
}

func NewGetAgentByEnsName(agentRepository repositories.AgentRepositoryInterface) *GetAgentByEnsName {
	return &GetAgentByEnsName{agentRepository: agentRepository}
}

func (u *GetAgentByEnsName) Execute(
	ctx context.Context,
	request *requests.EnsNameRequest,
) (*responses.AgentResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agent, err := u.agentRepository.GetByEnsName(ctx, request.EnsName)
	if err != nil {
		return nil, errors.NewInternalError("get_agent_failed", err)
	}
	if agent == nil {
		return nil, errors.NewNotFoundError("agent_not_found", "no indexed agent has this ENS name")
	}
	return responses.AgentResponseFromModel(agent), nil
}
