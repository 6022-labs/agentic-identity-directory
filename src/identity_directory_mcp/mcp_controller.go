package identity_directory_mcp

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const MCPEndpoint = "/mcp"

// MCPController serves the MCP streamable HTTP transport, stateless so any replica answers any call.
type MCPController struct {
	httpHandler fiber.Handler
}

func NewMCPController(mcpServer *mcp.Server) *MCPController {
	streamableHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return mcpServer
	}, &mcp.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
	})
	return &MCPController{httpHandler: adaptor.HTTPHandler(streamableHandler)}
}

func (c *MCPController) RegisterRoutes(app fiber.Router) {
	app.Post(MCPEndpoint, c.httpHandler)
	app.Get(MCPEndpoint, c.httpHandler)
	app.Delete(MCPEndpoint, c.httpHandler)
}
