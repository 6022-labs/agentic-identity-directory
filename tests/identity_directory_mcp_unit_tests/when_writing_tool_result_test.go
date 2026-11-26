package identity_directory_mcp_unit_tests_test

import (
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenWritingToolResultTestingSuite struct {
	sut *identity_directory_mcp.ToolResultWriter
}

func WhenWritingToolResultBeforeEach(t *testing.T) *WhenWritingToolResultTestingSuite {
	return &WhenWritingToolResultTestingSuite{sut: identity_directory_mcp.NewToolResultWriter(zap.NewNop())}
}

func (s *WhenWritingToolResultTestingSuite) text(t *testing.T, result *mcp.CallToolResult) string {
	assert.Len(t, result.Content, 1)
	content, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok)
	return content.Text
}

func TestWhenWritingToolResult(t *testing.T) {
	t.Parallel()

	t.Run("Given a payload and no error", func(t *testing.T) {
		t.Parallel()

		payload := map[string]string{"status": "healthy"}

		t.Run("Should return the payload as JSON text", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolResultBeforeEach(t)

			result := suite.sut.Write("get_agent", payload, nil)

			assert.False(t, result.IsError)
			assert.JSONEq(t, `{"status":"healthy"}`, suite.text(t, result))
		})
	})

	t.Run("Given an error", func(t *testing.T) {
		t.Parallel()

		sourceErr := apperrors.NewNotFoundError("agent_not_found", "no indexed agent has this ENS name")

		t.Run("Should return an error result instead of the payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolResultBeforeEach(t)

			result := suite.sut.Write("get_agent_by_ens_name", map[string]string{"ignored": "yes"}, sourceErr)

			assert.True(t, result.IsError)
			assert.JSONEq(
				t,
				`{"code":"agent_not_found","message":"no indexed agent has this ENS name"}`,
				suite.text(t, result),
			)
		})
	})

	t.Run("Given a payload that cannot be marshalled", func(t *testing.T) {
		t.Parallel()

		payload := map[string]any{"channel": make(chan int)}

		t.Run("Should return the generic internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolResultBeforeEach(t)

			result := suite.sut.Write("get_agent", payload, nil)

			assert.True(t, result.IsError)
			assert.JSONEq(
				t,
				`{"code":"tool_result_failed","message":"internal server error"}`,
				suite.text(t, result),
			)
		})
	})

	t.Run("Given a plain error", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("dial https://rpc.example/v2/secret-api-key")

		t.Run("Should hide the cause behind internal_error", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolResultBeforeEach(t)

			result := suite.sut.Write("get_agent_card", nil, sourceErr)

			assert.True(t, result.IsError)
			assert.JSONEq(t, `{"code":"internal_error","message":"internal server error"}`, suite.text(t, result))
		})
	})
}
