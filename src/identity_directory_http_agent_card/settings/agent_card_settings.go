package settings

import (
	"time"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const AgentCardSettingsKey = "agent_card"

const defaultAgentCardTimeoutSeconds = 10

type AgentCardSettings struct {
	Timeout time.Duration
	// AllowPrivateNetworks lets card fetches reach private, loopback and link-local addresses; local development only.
	AllowPrivateNetworks bool
}

type agentCardSettingsRaw struct {
	RequestTimeoutSeconds int  `koanf:"request_timeout_seconds"`
	AllowPrivateNetworks  bool `koanf:"allow_private_networks"`
}

func NewAgentCardSettings(logger *zap.Logger, k *koanf.Koanf) *AgentCardSettings {
	raw := agentCardSettingsRaw{RequestTimeoutSeconds: defaultAgentCardTimeoutSeconds}
	if err := k.Unmarshal(AgentCardSettingsKey, &raw); err != nil {
		logger.Fatal("failed to unmarshal agent_card settings", zap.Error(err))
	}
	if raw.RequestTimeoutSeconds <= 0 {
		logger.Warn("invalid agent_card.request_timeout_seconds value, defaulting to 10")
		raw.RequestTimeoutSeconds = defaultAgentCardTimeoutSeconds
	}
	if raw.AllowPrivateNetworks {
		logger.Warn("agent_card.allow_private_networks is on: card fetches may reach internal addresses")
	}
	return &AgentCardSettings{
		Timeout:              time.Duration(raw.RequestTimeoutSeconds) * time.Second,
		AllowPrivateNetworks: raw.AllowPrivateNetworks,
	}
}
