package identity_directory_mcp

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const GetAgentByEnsNameToolName = "get_agent_by_ens_name"

type EnsNameArgs struct {
	EnsName string `json:"ensName" jsonschema:"the agent's ENS name, e.g. my-agent.my-collection.137.6022.eth"`
}

type GetAgentByEnsNameTool struct {
	getAgentByEnsName *use_cases.GetAgentByEnsName
	toolResultWriter  *ToolResultWriter
}

func NewGetAgentByEnsNameTool(
	getAgentByEnsName *use_cases.GetAgentByEnsName,
	toolResultWriter *ToolResultWriter,
) *GetAgentByEnsNameTool {
	return &GetAgentByEnsNameTool{getAgentByEnsName: getAgentByEnsName, toolResultWriter: toolResultWriter}
}

func (t *GetAgentByEnsNameTool) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        GetAgentByEnsNameToolName,
		Description: "Reads one 6022 agent by its ENS name." + untrustedDataNote,
	}, t.handle)
}

func (t *GetAgentByEnsNameTool) handle(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	args EnsNameArgs,
) (*mcp.CallToolResult, any, error) {
	response, err := t.getAgentByEnsName.Execute(ctx, &requests.EnsNameRequest{EnsName: args.EnsName})
	return t.toolResultWriter.Write(GetAgentByEnsNameToolName, response, err), nil, nil
}
