package identity_directory_mvc

import (
	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/common/mvc"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/gofiber/fiber/v3"
)

const (
	AgentsEndpoint         = "/agents"
	AgentByEnsNameEndpoint = "/agents/ens/:ensName"
	AgentCardEndpoint      = "/agents/ens/:ensName/card"
	AgentEndpoint          = "/agents/:chainId/:collectionAddress/:tokenId"
)

type AgentController struct {
	errorWriter       *mvc.ErrorWriter
	searchAgents      *use_cases.SearchAgents
	getAgent          *use_cases.GetAgent
	getAgentByEnsName *use_cases.GetAgentByEnsName
	getAgentCard      *use_cases.GetAgentCard
}

func NewAgentController(
	errorWriter *mvc.ErrorWriter,
	searchAgents *use_cases.SearchAgents,
	getAgent *use_cases.GetAgent,
	getAgentByEnsName *use_cases.GetAgentByEnsName,
	getAgentCard *use_cases.GetAgentCard,
) *AgentController {
	return &AgentController{
		errorWriter:       errorWriter,
		searchAgents:      searchAgents,
		getAgent:          getAgent,
		getAgentByEnsName: getAgentByEnsName,
		getAgentCard:      getAgentCard,
	}
}

func (c *AgentController) RegisterRoutes(app fiber.Router) {
	app.Get(AgentsEndpoint, c.SearchAgents)
	app.Get(AgentByEnsNameEndpoint, c.GetAgentByEnsName)
	app.Get(AgentCardEndpoint, c.GetAgentCard)
	app.Get(AgentEndpoint, c.GetAgent)
}

// @Summary Search agents
// @Description Finds indexed 6022 agents; every field is written by the agent's owner.
// @Tags Agents
// @Produce json
// @Param query query string false "words in agent names, collections and attributes"
// @Param role query string false "exact role attribute"
// @Param chainId query int false "EVM chain id"
// @Param collectionAddress query string false "0x collection address"
// @Param owner query string false "0x owner address"
// @Param address query string false "one of the agent's own addresses"
// @Param published query string false "true or false: has a published url"
// @Param page query int false "page, from 1"
// @Param elementsPerPage query int false "1 to 100, 10 by default"
// @Success 200 {object} responses.SearchAgentsResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 400 {object} errors.BadRequestError "invalid_query"
// @Failure 500 {object} errors.InternalError
// @Router /agents [get]
func (c *AgentController) SearchAgents(ctx fiber.Ctx) error {
	request := &requests.SearchAgentsRequest{}
	if err := ctx.Bind().Query(request); err != nil {
		return c.errorWriter.Write(ctx, errors.NewBadRequestError("invalid_query", "query parameters are not valid"))
	}

	response, err := c.searchAgents.Execute(ctx.Context(), request)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}

	return ctx.JSON(response)
}

// @Summary Get an agent
// @Tags Agents
// @Produce json
// @Param chainId path int true "EVM chain id"
// @Param collectionAddress path string true "0x collection address"
// @Param tokenId path string true "decimal token id"
// @Success 200 {object} responses.AgentResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 400 {object} errors.BadRequestError "invalid_path"
// @Failure 404 {object} errors.NotFoundError "agent_not_found"
// @Failure 500 {object} errors.InternalError
// @Router /agents/{chainId}/{collectionAddress}/{tokenId} [get]
func (c *AgentController) GetAgent(ctx fiber.Ctx) error {
	request := &requests.GetAgentRequest{}
	if err := ctx.Bind().URI(request); err != nil {
		return c.errorWriter.Write(ctx, errors.NewBadRequestError("invalid_path", "path parameters are not valid"))
	}

	response, err := c.getAgent.Execute(ctx.Context(), request)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}

	return ctx.JSON(response)
}

// @Summary Get an agent by ENS name
// @Tags Agents
// @Produce json
// @Param ensName path string true "e.g. my-agent.my-collection.137.6022.eth"
// @Success 200 {object} responses.AgentResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 400 {object} errors.BadRequestError "invalid_path"
// @Failure 404 {object} errors.NotFoundError "agent_not_found"
// @Failure 500 {object} errors.InternalError
// @Router /agents/ens/{ensName} [get]
func (c *AgentController) GetAgentByEnsName(ctx fiber.Ctx) error {
	request := &requests.EnsNameRequest{}
	if err := ctx.Bind().URI(request); err != nil {
		return c.errorWriter.Write(ctx, errors.NewBadRequestError("invalid_path", "path parameters are not valid"))
	}

	response, err := c.getAgentByEnsName.Execute(ctx.Context(), request)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}

	return ctx.JSON(response)
}

// @Summary Get an agent's live card
// @Description Fetches the agent's /.well-known/6022 and checks its signature against the agent's addresses.
// @Tags Agents
// @Produce json
// @Param ensName path string true "e.g. my-agent.my-collection.137.6022.eth"
// @Success 200 {object} responses.AgentCardResponse
// @Failure 400 {object} errors.ValidationError
// @Failure 400 {object} errors.BadRequestError "invalid_path"
// @Failure 404 {object} errors.NotFoundError "agent_not_found or agent_url_not_published"
// @Failure 500 {object} errors.InternalError
// @Failure 502 {object} errors.UnavailableError "the card is unreachable, invalid or wrongly signed"
// @Router /agents/ens/{ensName}/card [get]
func (c *AgentController) GetAgentCard(ctx fiber.Ctx) error {
	request := &requests.EnsNameRequest{}
	if err := ctx.Bind().URI(request); err != nil {
		return c.errorWriter.Write(ctx, errors.NewBadRequestError("invalid_path", "path parameters are not valid"))
	}

	response, err := c.getAgentCard.Execute(ctx.Context(), request)
	if err != nil {
		return c.errorWriter.Write(ctx, err)
	}

	return ctx.JSON(response)
}
