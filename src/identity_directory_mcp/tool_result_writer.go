package identity_directory_mcp

import (
	"encoding/json"

	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

// ToolResultWriter turns a use case outcome into a tool result; error bodies match the HTTP API's, never a cause.
type ToolResultWriter struct {
	logger *zap.Logger
}

func NewToolResultWriter(logger *zap.Logger) *ToolResultWriter {
	return &ToolResultWriter{logger: logger}
}

func (w *ToolResultWriter) Write(toolName string, payload any, err error) *mcp.CallToolResult {
	if err != nil {
		return w.WriteError(toolName, err)
	}
	encoded, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return w.WriteError(toolName, errors.NewInternalError("tool_result_failed", marshalErr))
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(encoded)}}}
}

func (w *ToolResultWriter) WriteError(toolName string, err error) *mcp.CallToolResult {
	classified := errors.NewClassifiedError(err)
	encoded, marshalErr := json.Marshal(classified.Body)
	if marshalErr != nil {
		encoded = []byte(`{"code":"internal_error","message":"` + errors.InternalErrorMessage + `"}`)
	}
	fields := []zap.Field{
		zap.String("tool", toolName),
		zap.String("code", classified.Code),
		zap.ByteString("response", encoded),
		w.causeField(classified),
	}
	if classified.Failed() {
		w.logger.Error("tool call failed", fields...)
	} else {
		w.logger.Warn("tool call rejected", fields...)
	}
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: string(encoded)}}}
}

func (w *ToolResultWriter) causeField(classified *errors.ClassifiedError) zap.Field {
	if classified.Cause == nil {
		return zap.Skip()
	}
	return zap.String("cause", classified.CauseMessage())
}
