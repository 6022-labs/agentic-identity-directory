package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/clients"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/services"
	clients_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_http_agent_card_mocks/clients_mocks"
	metrics_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_http_agent_card_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenGettingAnAgentCardTestingSuite struct {
	mockClient *clients_mocks.MockClientWithResponsesInterface
	sut        *services.AgentCardRequester
}

func WhenGettingAnAgentCardBeforeEach(t *testing.T) *WhenGettingAnAgentCardTestingSuite {
	controller := gomock.NewController(t)
	mockClient := clients_mocks.NewMockClientWithResponsesInterface(controller)
	mockAgentCardMetrics := metrics_mocks.NewMockAgentCardMetricsInterface(controller)
	mockAgentCardMetrics.EXPECT().RecordFetch(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return &WhenGettingAnAgentCardTestingSuite{
		mockClient: mockClient,
		sut:        services.NewAgentCardRequester(mockClient, mockAgentCardMetrics),
	}
}

// answer is what the generated client returns for one call.
func (s *WhenGettingAnAgentCardTestingSuite) answer(
	status int,
	card *clients.ResponsesWellKnownAgent6022,
	body string,
) *clients.GetWellKnown6022Response {
	return &clients.GetWellKnown6022Response{
		Body:         []byte(body),
		HTTPResponse: &http.Response{StatusCode: status},
		JSON200:      card,
	}
}

func TestWhenGettingAnAgentCard(t *testing.T) {
	t.Parallel()

	t.Run("Given an agent url that is not absolute", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the agent", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)

			agentCard, err := suite.sut.GetCard(context.Background(), "scout.example")

			assert.Nil(t, agentCard)
			assert.ErrorContains(t, err, "not an absolute URL")
		})
	})

	t.Run("Given the agent cannot be reached", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("connection refused")

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockClient.EXPECT().GetWellKnown6022WithResponse(gomock.Any(), gomock.Any()).Return(nil, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			agentCard, err := suite.sut.GetCard(context.Background(), "https://scout.example")

			assert.Nil(t, agentCard)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the agent answers an error status", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockClient.EXPECT().
				GetWellKnown6022WithResponse(gomock.Any(), gomock.Any()).
				Return(suite.answer(http.StatusFound, nil, ""), nil)
		}

		t.Run("Should fail with the status", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			agentCard, err := suite.sut.GetCard(context.Background(), "https://scout.example")

			assert.Nil(t, agentCard)
			assert.ErrorContains(t, err, "unexpected status code 302")
		})
	})

	t.Run("Given an answer that is not a JSON card", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockClient.EXPECT().
				GetWellKnown6022WithResponse(gomock.Any(), gomock.Any()).
				Return(suite.answer(http.StatusOK, nil, "<html>"), nil)
		}

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			agentCard, err := suite.sut.GetCard(context.Background(), "https://scout.example")

			assert.Nil(t, agentCard)
			assert.ErrorContains(t, err, "not a JSON card")
		})
	})

	t.Run("Given a 6022 card", func(t *testing.T) {
		t.Parallel()

		body := `{"protocol":"6022","name":"scout"}`
		card := &clients.ResponsesWellKnownAgent6022{
			Protocol:        new("6022"),
			ProtocolVersion: new("0.1"),
			Name:            new("scout"),
			EnsDomain:       new("scout.agents.137.6022.eth"),
			Endpoints: &clients.ResponsesAgent6022Endpoints{
				Responses: new("https://scout.example/api/responses"),
			},
			PaymentMethods: &[]string{"x402"},
			Signatures: &[]clients.ResponsesAgentCardSignature{
				{Protected: new("p"), Signature: new("s")},
			},
		}

		initSuite := func(suite *WhenGettingAnAgentCardTestingSuite) {
			suite.mockClient.EXPECT().
				GetWellKnown6022WithResponse(gomock.Any(), gomock.Any()).
				Return(suite.answer(http.StatusOK, card, body), nil)
		}

		t.Run("Should map it and keep the raw body its signatures cover", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			initSuite(suite)

			agentCard, err := suite.sut.GetCard(context.Background(), "https://scout.example")

			require.NoError(t, err)
			assert.Equal(t, &models.AgentCard{
				Protocol:        "6022",
				ProtocolVersion: "0.1",
				Name:            "scout",
				EnsDomain:       "scout.agents.137.6022.eth",
				Endpoints:       map[string]string{"responses": "https://scout.example/api/responses"},
				PaymentMethods:  []string{"x402"},
				Signatures:      []models.AgentCardSignature{{Protected: "p", Signature: "s"}},
				Raw:             []byte(body),
			}, agentCard)
		})

		t.Run("Should send the request to the agent's own url", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			suite.mockClient.EXPECT().
				GetWellKnown6022WithResponse(gomock.Any(), gomock.Any()).
				DoAndReturn(func(
					ctx context.Context,
					editors ...clients.RequestEditorFn,
				) (*clients.GetWellKnown6022Response, error) {
					request, err := clients.NewGetWellKnown6022Request("http://agent.invalid")
					require.NoError(t, err)
					for _, editor := range editors {
						require.NoError(t, editor(ctx, request))
					}
					assert.Equal(t, "https://scout.example/base/.well-known/6022", request.URL.String())
					assert.Equal(t, "scout.example", request.Host)
					return suite.answer(http.StatusOK, card, body), nil
				})

			_, err := suite.sut.GetCard(context.Background(), "https://scout.example/base/")

			require.NoError(t, err)
		})
	})
}
