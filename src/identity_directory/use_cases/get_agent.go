package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/responses"
)

type GetAgent struct {
	agentRepository repositories.AgentRepositoryInterface
}

func NewGetAgent(agentRepository repositories.AgentRepositoryInterface) *GetAgent {
	return &GetAgent{agentRepository: agentRepository}
}

func (u *GetAgent) Execute(ctx context.Context, request *requests.GetAgentRequest) (*responses.AgentResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agent, err := u.agentRepository.Get(ctx, request.ChainId, request.AgentRef())
	if err != nil {
		return nil, errors.NewInternalError("get_agent_failed", err)
	}
	if agent == nil {
		return nil, errors.NewNotFoundError("agent_not_found", "no indexed agent has this token")
	}
	return responses.AgentResponseFromModel(agent), nil
}
