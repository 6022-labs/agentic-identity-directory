package identity_directory_mcp

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const SearchAgentsToolName = "search_agents"

type SearchAgentsArgs struct {
	Query             string `json:"query,omitempty"             jsonschema:"words to find in agent names, collections and attributes"`
	Role              string `json:"role,omitempty"              jsonschema:"exact role attribute, e.g. expert, facilitator, clone or human"`
	ChainId           uint64 `json:"chainId,omitempty"           jsonschema:"EVM chain id, e.g. 137 (Polygon) or 80002 (Polygon Amoy)"`
	CollectionAddress string `json:"collectionAddress,omitempty" jsonschema:"0x address of the agent collection"`
	Owner             string `json:"owner,omitempty"             jsonschema:"0x address owning the agent NFT"`
	Address           string `json:"address,omitempty"           jsonschema:"one of the agent's own addresses (evm, btc...)"`
	Published         string `json:"published,omitempty"         jsonschema:"true: only agents that published a url, false: only those that did not"`
	Page              uint   `json:"page,omitempty"              jsonschema:"page number from 1 (default 1)"`
	ElementsPerPage   uint   `json:"elementsPerPage,omitempty"   jsonschema:"page size, 1 to 100 (default 10)"`
}

type SearchAgentsTool struct {
	searchAgents     *use_cases.SearchAgents
	toolResultWriter *ToolResultWriter
}

func NewSearchAgentsTool(searchAgents *use_cases.SearchAgents, toolResultWriter *ToolResultWriter) *SearchAgentsTool {
	return &SearchAgentsTool{searchAgents: searchAgents, toolResultWriter: toolResultWriter}
}

func (t *SearchAgentsTool) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name: SearchAgentsToolName,
		Description: "Finds 6022 agents registered on-chain, by words, role, chain, collection, owner or address. " +
			"Each agent comes with its ENS name and, once published, the url to call it." + untrustedDataNote,
	}, t.handle)
}

func (t *SearchAgentsTool) handle(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	args SearchAgentsArgs,
) (*mcp.CallToolResult, any, error) {
	response, err := t.searchAgents.Execute(ctx, &requests.SearchAgentsRequest{
		PaginateRequest: requests.PaginateRequest{
			Page:            t.optionalUint(args.Page),
			ElementsPerPage: t.optionalUint(args.ElementsPerPage),
		},
		Query:             args.Query,
		Role:              args.Role,
		ChainId:           args.ChainId,
		CollectionAddress: args.CollectionAddress,
		Owner:             args.Owner,
		Address:           args.Address,
		Published:         args.Published,
	})
	return t.toolResultWriter.Write(SearchAgentsToolName, response, err), nil, nil
}

func (t *SearchAgentsTool) optionalUint(value uint) *uint {
	if value == 0 {
		return nil
	}
	return &value
}
