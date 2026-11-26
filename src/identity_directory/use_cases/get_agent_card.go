package use_cases

import (
	"context"
	"net/url"
	"strings"

	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/responses"
)

const agentCardProtocol = "6022"

type GetAgentCard struct {
	agentRepository            repositories.AgentRepositoryInterface
	agentCardRequester         interfaces.AgentCardRequesterInterface
	agentCardSignatureVerifier interfaces.AgentCardSignatureVerifierInterface
}

func NewGetAgentCard(
	agentRepository repositories.AgentRepositoryInterface,
	agentCardRequester interfaces.AgentCardRequesterInterface,
	agentCardSignatureVerifier interfaces.AgentCardSignatureVerifierInterface,
) *GetAgentCard {
	return &GetAgentCard{
		agentRepository:            agentRepository,
		agentCardRequester:         agentCardRequester,
		agentCardSignatureVerifier: agentCardSignatureVerifier,
	}
}

func (u *GetAgentCard) Execute(
	ctx context.Context,
	request *requests.EnsNameRequest,
) (*responses.AgentCardResponse, error) {
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
	agentUrl := agent.Url()
	if agentUrl == "" {
		return nil, errors.NewNotFoundError("agent_url_not_published", "the agent has not published its url")
	}
	parsedUrl, err := url.Parse(agentUrl)
	if err != nil || (parsedUrl.Scheme != "https" && parsedUrl.Scheme != "http") || parsedUrl.Host == "" {
		return nil, errors.NewUnavailableError("agent_url_invalid", "the agent's url is not an http(s) URL", err)
	}

	agentCard, err := u.agentCardRequester.GetCard(ctx, agentUrl)
	if err != nil {
		return nil, errors.NewUnavailableError("agent_card_unavailable", "the agent's card could not be read", err)
	}
	if agentCard.Protocol != agentCardProtocol {
		return nil, errors.NewUnavailableError("agent_card_invalid", "the agent's card is not a 6022 card", nil)
	}
	verified, err := u.agentCardSignatureVerifier.Verify(agentCard, agent.EvmAddresses())
	if err != nil {
		return nil, errors.NewUnavailableError(
			"agent_card_signature_invalid",
			"the agent's card is not signed by the agent",
			err,
		)
	}
	// A signature only vouches for the agent the card names: another agent may list the signer's address too.
	if verified && !strings.EqualFold(agentCard.EnsDomain, agent.EnsName) {
		return nil, errors.NewUnavailableError("agent_card_invalid", "the agent's card names another agent", nil)
	}
	return responses.AgentCardResponseFromModel(agent, agentCard, verified), nil
}
