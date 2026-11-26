package configurations

import (
	"net/http"

	identity_directory_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory/services/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/metrics/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/services"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/transports"
	"go.uber.org/dig"
)

func AddIdentityDirectoryHttpAgentCardConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewAgentCardSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewAgentCardMetrics,
		dig.As(new(metrics_interfaces.AgentCardMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		func(
			agentCardSettings *settings.AgentCardSettings,
			agentCardMetrics metrics_interfaces.AgentCardMetricsInterface,
		) *services.AgentCardRequester {
			return services.NewAgentCardRequester(
				&http.Client{
					Timeout:   agentCardSettings.Timeout,
					Transport: transports.NewPublicRoundTripper(agentCardSettings),
					// A redirect could point anywhere the agent's own URL is not.
					CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
				},
				agentCardMetrics,
			)
		},
		dig.As(new(identity_directory_interfaces.AgentCardRequesterInterface)),
	); err != nil {
		panic(err)
	}
}
