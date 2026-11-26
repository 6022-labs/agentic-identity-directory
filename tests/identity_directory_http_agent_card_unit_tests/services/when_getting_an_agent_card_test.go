package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/services"
	metrics_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_http_agent_card_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenGettingAnAgentCardTestingSuite struct {
	sut *services.AgentCardRequester
}

func WhenGettingAnAgentCardBeforeEach(t *testing.T) *WhenGettingAnAgentCardTestingSuite {
	mockAgentCardMetrics := metrics_mocks.NewMockAgentCardMetricsInterface(gomock.NewController(t))
	mockAgentCardMetrics.EXPECT().RecordFetch(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	httpClient := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return &WhenGettingAnAgentCardTestingSuite{sut: services.NewAgentCardRequester(httpClient, mockAgentCardMetrics)}
}

func (s *WhenGettingAnAgentCardTestingSuite) serve(t *testing.T, handler http.HandlerFunc) string {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server.URL
}

func TestWhenGettingAnAgentCard(t *testing.T) {
	t.Parallel()

	t.Run("Given the agent answers an error status", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail with the status", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			agentUrl := suite.serve(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			})

			agentCard, err := suite.sut.GetCard(context.Background(), agentUrl)

			assert.Nil(t, agentCard)
			assert.ErrorContains(t, err, "unexpected status code 404")
		})
	})

	t.Run("Given the agent redirects", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not follow it", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			agentUrl := suite.serve(t, func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "http://169.254.169.254/latest/meta-data", http.StatusFound)
			})

			agentCard, err := suite.sut.GetCard(context.Background(), agentUrl)

			assert.Nil(t, agentCard)
			assert.ErrorContains(t, err, "unexpected status code 302")
		})
	})

	t.Run("Given a card over 1 MiB", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			agentUrl := suite.serve(t, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"name":"` + strings.Repeat("a", 1<<20) + `"}`))
			})

			agentCard, err := suite.sut.GetCard(context.Background(), agentUrl)

			assert.Nil(t, agentCard)
			assert.ErrorContains(t, err, "card exceeds")
		})
	})

	t.Run("Given a card that is not JSON", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail to decode it", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			agentUrl := suite.serve(t, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`<html>`))
			})

			agentCard, err := suite.sut.GetCard(context.Background(), agentUrl)

			assert.Nil(t, agentCard)
			assert.ErrorContains(t, err, "decode card")
		})
	})

	t.Run("Given a 6022 card", func(t *testing.T) {
		t.Parallel()

		body := `{"protocol":"6022","protocolVersion":"0.1","name":"scout",` +
			`"endpoints":{"responses":"https://scout.example/api/responses"},"paymentMethods":["x402"],` +
			`"signatures":[{"protected":"p","signature":"s"}]}`

		t.Run("Should read it from /.well-known/6022 and keep its raw body", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentCardBeforeEach(t)
			agentUrl := suite.serve(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/.well-known/6022" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				_, _ = w.Write([]byte(body))
			})

			agentCard, err := suite.sut.GetCard(context.Background(), agentUrl+"/")

			require.NoError(t, err)
			assert.Equal(t, "6022", agentCard.Protocol)
			assert.Equal(t, "scout", agentCard.Name)
			assert.Equal(t, map[string]string{"responses": "https://scout.example/api/responses"}, agentCard.Endpoints)
			assert.Equal(t, []string{"x402"}, agentCard.PaymentMethods)
			require.Len(t, agentCard.Signatures, 1)
			assert.Equal(t, "p", agentCard.Signatures[0].Protected)
			assert.Equal(t, body, string(agentCard.Raw))
		})
	})
}
