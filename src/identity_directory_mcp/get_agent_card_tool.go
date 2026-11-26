package identity_directory_mcp

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const GetAgentCardToolName = "get_agent_card"

type GetAgentCardTool struct {
	getAgentCard     *use_cases.GetAgentCard
	toolResultWriter *ToolResultWriter
}

func NewGetAgentCardTool(getAgentCard *use_cases.GetAgentCard, toolResultWriter *ToolResultWriter) *GetAgentCardTool {
	return &GetAgentCardTool{getAgentCard: getAgentCard, toolResultWriter: toolResultWriter}
}

func (t *GetAgentCardTool) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: GetAgentCardToolName,
		Description: "Fetches a 6022 agent's live card (/.well-known/6022): the endpoints to call it, per API " +
			"dialect, and the payment methods it accepts. verified is true when one of the agent's on-chain " +
			"addresses signed it." + untrustedDataNote,
	}, t.handle)
}

func (t *GetAgentCardTool) handle(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	args EnsNameArgs,
) (*mcp.CallToolResult, any, error) {
	response, err := t.getAgentCard.Execute(ctx, &requests.EnsNameRequest{EnsName: args.EnsName})
	return t.toolResultWriter.Write(GetAgentCardToolName, response, err), nil, nil
}
