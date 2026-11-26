package transports_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/transports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenRoundTrippingToAnAgent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/huge" {
			_, _ = w.Write([]byte(strings.Repeat("a", 1<<20+1)))
			return
		}
		_, _ = w.Write([]byte(strings.Repeat("a", 1<<20)))
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

	allowing := &http.Client{Transport: transports.NewPublicRoundTripper(&settings.AgentCardSettings{
		Timeout:              time.Second,
		AllowPrivateNetworks: true,
	})}

	t.Run("Given an answer of exactly 1 MiB", func(t *testing.T) {
		t.Parallel()

		t.Run("Should read it whole", func(t *testing.T) {
			t.Parallel()

			response, err := allowing.Get(server.URL)
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()

			body, err := io.ReadAll(response.Body)

			require.NoError(t, err)
			assert.Len(t, body, 1<<20)
		})
	})

	t.Run("Given an answer over 1 MiB", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail the read with ErrResponseTooLarge", func(t *testing.T) {
			t.Parallel()

			response, err := allowing.Get(server.URL + "/huge")
			require.NoError(t, err)
			defer func() { _ = response.Body.Close() }()

			_, err = io.ReadAll(response.Body)

			assert.ErrorIs(t, err, transports.ErrResponseTooLarge)
		})
	})
}
