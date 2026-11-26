package identity_directory_mcp

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const ListCollectionsToolName = "list_collections"

type ListCollectionsArgs struct {
	ChainId         uint64 `json:"chainId,omitempty"         jsonschema:"EVM chain id, e.g. 137 (Polygon) or 80002 (Polygon Amoy)"`
	Page            uint   `json:"page,omitempty"            jsonschema:"page number from 1 (default 1)"`
	ElementsPerPage uint   `json:"elementsPerPage,omitempty" jsonschema:"page size, 1 to 100 (default 10)"`
}

type ListCollectionsTool struct {
	listCollections  *use_cases.ListCollections
	toolResultWriter *ToolResultWriter
}

func NewListCollectionsTool(
	listCollections *use_cases.ListCollections,
	toolResultWriter *ToolResultWriter,
) *ListCollectionsTool {
	return &ListCollectionsTool{listCollections: listCollections, toolResultWriter: toolResultWriter}
}

func (t *ListCollectionsTool) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: ListCollectionsToolName,
		Description: "Lists the 6022 agent collections: groups of agents, each with a name and a description." +
			untrustedDataNote,
	}, t.handle)
}

func (t *ListCollectionsTool) handle(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	args ListCollectionsArgs,
) (*mcp.CallToolResult, any, error) {
	response, err := t.listCollections.Execute(ctx, &requests.ListCollectionsRequest{
		PaginateRequest: requests.PaginateRequest{
			Page:            t.optionalUint(args.Page),
			ElementsPerPage: t.optionalUint(args.ElementsPerPage),
		},
		ChainId: args.ChainId,
	})
	return t.toolResultWriter.Write(ListCollectionsToolName, response, err), nil, nil
}

func (t *ListCollectionsTool) optionalUint(value uint) *uint {
	if value == 0 {
		return nil
	}
	return &value
}
