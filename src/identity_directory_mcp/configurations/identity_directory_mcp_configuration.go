package configurations

import (
	mvc_interfaces "github.com/6022-labs/agentic-identity-directory/src/common/mvc/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_mcp"
	mcp_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory_mcp/interfaces"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/dig"
)

const (
	controllersGroup = "controllers"
	mcpToolsGroup    = "mcp_tools"

	mcpServerName    = "agentic-identity-directory"
	mcpServerVersion = "1.0.0"
)

func AddIdentityDirectoryMcpConfiguration(container *dig.Container) {
	if err := container.Provide(identity_directory_mcp.NewToolResultWriter); err != nil {
		panic(err)
	}

	provideTool(container, identity_directory_mcp.NewSearchAgentsTool)
	provideTool(container, identity_directory_mcp.NewGetAgentTool)
	provideTool(container, identity_directory_mcp.NewGetAgentByEnsNameTool)
	provideTool(container, identity_directory_mcp.NewGetAgentCardTool)
	provideTool(container, identity_directory_mcp.NewListCollectionsTool)

	if err := container.Provide(newMCPServer); err != nil {
		panic(err)
	}
	if err := container.Provide(
		identity_directory_mcp.NewMCPController,
		dig.Group(controllersGroup),
		dig.As(new(mvc_interfaces.ControllerInterface)),
	); err != nil {
		panic(err)
	}
}

type newMCPServerParams struct {
	dig.In

	Tools []mcp_interfaces.ToolInterface `group:"mcp_tools"`
}

func newMCPServer(p newMCPServerParams) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: mcpServerName, Version: mcpServerVersion}, nil)
	for _, tool := range p.Tools {
		tool.Register(server)
	}
	return server
}

func provideTool(container *dig.Container, ctor any) {
	if err := container.Provide(
		ctor,
		dig.Group(mcpToolsGroup),
		dig.As(new(mcp_interfaces.ToolInterface)),
	); err != nil {
		panic(err)
	}
}
