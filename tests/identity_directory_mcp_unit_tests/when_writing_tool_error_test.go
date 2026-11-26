package identity_directory_mcp_unit_tests_test

import (
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type WhenWritingToolErrorTestingSuite struct {
	sut  *identity_directory_mcp.ToolResultWriter
	logs *observer.ObservedLogs
}

func WhenWritingToolErrorBeforeEach(t *testing.T) *WhenWritingToolErrorTestingSuite {
	core, logs := observer.New(zapcore.DebugLevel)
	return &WhenWritingToolErrorTestingSuite{sut: identity_directory_mcp.NewToolResultWriter(zap.New(core)), logs: logs}
}

func (s *WhenWritingToolErrorTestingSuite) text(t *testing.T, result *mcp.CallToolResult) string {
	assert.Len(t, result.Content, 1)
	content, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok)
	return content.Text
}

func TestWhenWritingToolError(t *testing.T) {
	t.Parallel()

	t.Run("Given a ValidationError", func(t *testing.T) {
		t.Parallel()

		sourceErr := apperrors.NewValidationError("name", "required")

		t.Run("Should return the field and message", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolErrorBeforeEach(t)

			result := suite.sut.WriteError("search_agents", sourceErr)

			assert.True(t, result.IsError)
			assert.JSONEq(t, `{"field":"name","message":"required"}`, suite.text(t, result))
		})

		t.Run("Should log a warning", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolErrorBeforeEach(t)

			suite.sut.WriteError("search_agents", sourceErr)

			logs := suite.logs.All()
			assert.Len(t, logs, 1)
			assert.Equal(t, zapcore.WarnLevel, logs[0].Level)
			assert.Equal(t, "search_agents", logs[0].ContextMap()["tool"])
		})
	})

	t.Run("Given a NotFoundError", func(t *testing.T) {
		t.Parallel()

		sourceErr := apperrors.NewNotFoundError("agent_not_found", "no indexed agent has this ENS name")

		t.Run("Should return the code and message", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolErrorBeforeEach(t)

			result := suite.sut.WriteError("get_agent_by_ens_name", sourceErr)

			assert.True(t, result.IsError)
			assert.JSONEq(
				t,
				`{"code":"agent_not_found","message":"no indexed agent has this ENS name"}`,
				suite.text(t, result),
			)
		})
	})

	t.Run("Given an InternalError", func(t *testing.T) {
		t.Parallel()

		sourceErr := apperrors.NewInternalError(
			"get_agent_failed",
			errors.New("dial https://rpc.example/v2/secret-api-key"),
		)

		t.Run("Should return the code and the generic message", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolErrorBeforeEach(t)

			result := suite.sut.WriteError("get_agent_card", sourceErr)

			assert.True(t, result.IsError)
			assert.JSONEq(t, `{"code":"get_agent_failed","message":"internal server error"}`, suite.text(t, result))
		})

		t.Run("Should log an error with the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolErrorBeforeEach(t)

			suite.sut.WriteError("get_agent_card", sourceErr)

			logs := suite.logs.All()
			assert.Len(t, logs, 1)
			assert.Equal(t, zapcore.ErrorLevel, logs[0].Level)
			assert.Equal(t, "dial https://rpc.example/v2/secret-api-key", logs[0].ContextMap()["cause"])
		})
	})

	t.Run("Given an UnavailableError", func(t *testing.T) {
		t.Parallel()

		sourceErr := apperrors.NewUnavailableError(
			"agent_card_unavailable",
			"the agent's card could not be read",
			errors.New("dial https://rpc.example/v2/secret-api-key"),
		)

		t.Run("Should return the fixed message without the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolErrorBeforeEach(t)

			result := suite.sut.WriteError("get_agent_card", sourceErr)

			assert.True(t, result.IsError)
			assert.JSONEq(
				t,
				`{"code":"agent_card_unavailable","message":"the agent's card could not be read"}`,
				suite.text(t, result),
			)
		})
	})

	t.Run("Given a plain error", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("dial https://rpc.example/v2/secret-api-key")

		t.Run("Should return internal_error with the generic message", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingToolErrorBeforeEach(t)

			result := suite.sut.WriteError("get_agent_card", sourceErr)

			assert.True(t, result.IsError)
			assert.JSONEq(t, `{"code":"internal_error","message":"internal server error"}`, suite.text(t, result))
		})
	})
}
