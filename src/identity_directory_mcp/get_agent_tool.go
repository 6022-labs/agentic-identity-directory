package identity_directory_mcp

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const GetAgentToolName = "get_agent"

type GetAgentArgs struct {
	ChainId           uint64 `json:"chainId"           jsonschema:"EVM chain id, e.g. 137 (Polygon) or 80002 (Polygon Amoy)"`
	CollectionAddress string `json:"collectionAddress" jsonschema:"0x address of the agent collection"`
	TokenId           string `json:"tokenId"           jsonschema:"the agent's token id in its collection, in decimal"`
}

type GetAgentTool struct {
	getAgent         *use_cases.GetAgent
	toolResultWriter *ToolResultWriter
}

func NewGetAgentTool(getAgent *use_cases.GetAgent, toolResultWriter *ToolResultWriter) *GetAgentTool {
	return &GetAgentTool{getAgent: getAgent, toolResultWriter: toolResultWriter}
}

func (t *GetAgentTool) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        GetAgentToolName,
		Description: "Reads one 6022 agent by its chain, collection and token id." + untrustedDataNote,
	}, t.handle)
}

func (t *GetAgentTool) handle(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	args GetAgentArgs,
) (*mcp.CallToolResult, any, error) {
	response, err := t.getAgent.Execute(ctx, &requests.GetAgentRequest{
		ChainId:           args.ChainId,
		CollectionAddress: args.CollectionAddress,
		TokenId:           args.TokenId,
	})
	return t.toolResultWriter.Write(GetAgentToolName, response, err), nil, nil
}
