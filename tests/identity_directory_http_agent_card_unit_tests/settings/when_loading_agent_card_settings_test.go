package settings_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWhenLoadingAgentCardSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given no setting", func(t *testing.T) {
		t.Parallel()

		t.Run("Should time out after 10 seconds and refuse private networks", func(t *testing.T) {
			t.Parallel()

			agentCardSettings := settings.NewAgentCardSettings(zap.NewNop(), koanf.New("."))

			assert.Equal(t, &settings.AgentCardSettings{Timeout: 10 * time.Second}, agentCardSettings)
		})
	})

	t.Run("Given every setting from env", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		require.NoError(t, k.Load(confmap.Provider(map[string]any{
			"agent_card.request_timeout_seconds": "3",
			"agent_card.allow_private_networks":  "true",
		}, "."), nil))

		t.Run("Should use them", func(t *testing.T) {
			t.Parallel()

			agentCardSettings := settings.NewAgentCardSettings(zap.NewNop(), k)

			assert.Equal(
				t,
				&settings.AgentCardSettings{Timeout: 3 * time.Second, AllowPrivateNetworks: true},
				agentCardSettings,
			)
		})
	})
}
