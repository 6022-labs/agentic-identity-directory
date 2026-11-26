package interfaces

import "github.com/modelcontextprotocol/go-sdk/mcp"

// ToolInterface registers one MCP tool; every tool joins the MCP tools dig group.
type ToolInterface interface {
	Register(server *mcp.Server)
}
