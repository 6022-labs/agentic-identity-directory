package transports_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/transports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenDialingAnAgent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	t.Run("Given an agent on a loopback address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to dial it", func(t *testing.T) {
			t.Parallel()

			client := &http.Client{
				Transport: transports.NewPublicRoundTripper(&settings.AgentCardSettings{Timeout: time.Second}),
			}

			response, err := client.Get(server.URL)
			if response != nil {
				_ = response.Body.Close()
			}

			assert.ErrorIs(t, err, transports.ErrAddressRefused)
		})
	})

	t.Run("Given private networks are allowed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should dial a loopback agent", func(t *testing.T) {
			t.Parallel()

			client := &http.Client{Transport: transports.NewPublicRoundTripper(&settings.AgentCardSettings{
				Timeout:              time.Second,
				AllowPrivateNetworks: true,
			})}

			response, err := client.Get(server.URL)

			require.NoError(t, err)
			_ = response.Body.Close()
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})
	})
}
