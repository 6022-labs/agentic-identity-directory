package identity_directory_mcp_unit_tests_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_mcp"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/services_mocks/interfaces_mocks"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// agentNodeToolsResult is the shape agent-node decodes tools/list into: a property type is a single string.
type agentNodeToolsResult struct {
	Result struct {
		Tools []struct {
			Name        string `json:"name"`
			InputSchema struct {
				Type                 string `json:"type"`
				AdditionalProperties bool   `json:"additionalProperties"`
				Properties           map[string]struct {
					Type string `json:"type"`
				} `json:"properties"`
			} `json:"inputSchema"`
		} `json:"tools"`
	} `json:"result"`
}

type agentNodeCallResult struct {
	Result struct {
		IsError bool `json:"isError"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	} `json:"result"`
}

type WhenServingAnAgentNodeClientTestingSuite struct {
	mockAgentRepository           *repositories_mocks.MockAgentRepositoryInterface
	mockAgentCollectionRepository *repositories_mocks.MockAgentCollectionRepositoryInterface
	server                        *httptest.Server
}

func WhenServingAnAgentNodeClientBeforeEach(t *testing.T) *WhenServingAnAgentNodeClientTestingSuite {
	controller := gomock.NewController(t)
	suite := &WhenServingAnAgentNodeClientTestingSuite{
		mockAgentRepository:           repositories_mocks.NewMockAgentRepositoryInterface(controller),
		mockAgentCollectionRepository: repositories_mocks.NewMockAgentCollectionRepositoryInterface(controller),
	}
	toolResultWriter := identity_directory_mcp.NewToolResultWriter(zap.NewNop())
	getAgentCard := use_cases.NewGetAgentCard(
		suite.mockAgentRepository,
		interfaces_mocks.NewMockAgentCardRequesterInterface(controller),
		interfaces_mocks.NewMockAgentCardSignatureVerifierInterface(controller),
	)

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	identity_directory_mcp.NewSearchAgentsTool(use_cases.NewSearchAgents(suite.mockAgentRepository), toolResultWriter).
		Register(server)
	identity_directory_mcp.NewGetAgentTool(use_cases.NewGetAgent(suite.mockAgentRepository), toolResultWriter).
		Register(server)
	getAgentByEnsName := use_cases.NewGetAgentByEnsName(suite.mockAgentRepository)
	identity_directory_mcp.NewGetAgentByEnsNameTool(getAgentByEnsName, toolResultWriter).Register(server)
	identity_directory_mcp.NewGetAgentCardTool(getAgentCard, toolResultWriter).Register(server)
	listCollections := use_cases.NewListCollections(suite.mockAgentCollectionRepository)
	identity_directory_mcp.NewListCollectionsTool(listCollections, toolResultWriter).Register(server)

	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	suite.server = httptest.NewServer(handler)
	t.Cleanup(suite.server.Close)
	return suite
}

// post sends one JSON-RPC call the way agent-node does: no initialize, no session, no protocol version header.
func (s *WhenServingAnAgentNodeClientTestingSuite) post(t *testing.T, body string, result any) {
	request, err := http.NewRequest(http.MethodPost, s.server.URL, bytes.NewBufferString(body))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")

	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.NoError(t, json.NewDecoder(response.Body).Decode(result))
}

func TestWhenServingAnAgentNodeClient(t *testing.T) {
	t.Parallel()

	t.Run("Given a tools/list call", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list every tool in a schema agent-node can decode", func(t *testing.T) {
			t.Parallel()

			suite := WhenServingAnAgentNodeClientBeforeEach(t)
			var result agentNodeToolsResult

			suite.post(t, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`, &result)

			names := []string{}
			for _, tool := range result.Result.Tools {
				names = append(names, tool.Name)
				assert.Equal(t, "object", tool.InputSchema.Type)
				for name, property := range tool.InputSchema.Properties {
					assert.NotEmpty(t, property.Type, "%s.%s", tool.Name, name)
				}
			}
			assert.ElementsMatch(t, []string{
				"search_agents", "get_agent", "get_agent_by_ens_name", "get_agent_card", "list_collections",
			}, names)
		})
	})

	t.Run("Given a tools/call with null arguments", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenServingAnAgentNodeClientTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Search(gomock.Any(), gomock.Any()).
				Return([]*models.Agent{}, int64(0), nil)
		}

		t.Run("Should run the search with its defaults", func(t *testing.T) {
			t.Parallel()

			suite := WhenServingAnAgentNodeClientBeforeEach(t)
			initSuite(suite)
			var result agentNodeCallResult

			suite.post(t, `{"jsonrpc":"2.0","id":2,"method":"tools/call",`+
				`"params":{"name":"search_agents","arguments":null}}`, &result)

			assert.False(t, result.Result.IsError)
			require.Len(t, result.Result.Content, 1)
			assert.JSONEq(
				t,
				`{"items":[],"page":1,"elementsPerPage":10,"totalElements":0}`,
				result.Result.Content[0].Text,
			)
		})
	})

	t.Run("Given a tools/call the use case rejects", func(t *testing.T) {
		t.Parallel()

		t.Run("Should answer a tool error with the validation body", func(t *testing.T) {
			t.Parallel()

			suite := WhenServingAnAgentNodeClientBeforeEach(t)
			var result agentNodeCallResult

			suite.post(t, `{"jsonrpc":"2.0","id":3,"method":"tools/call",`+
				`"params":{"name":"get_agent_by_ens_name","arguments":{"ensName":" "}}}`, &result)

			assert.True(t, result.Result.IsError)
			require.Len(t, result.Result.Content, 1)
			assert.JSONEq(t, `{"field":"ensName","message":"must be an ENS name of 255 characters at most"}`,
				result.Result.Content[0].Text)
		})
	})
}
