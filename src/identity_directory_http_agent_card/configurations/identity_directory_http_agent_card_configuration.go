package configurations

import (
	"context"
	"net/http"

	identity_directory_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory/services/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/clients"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/metrics/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/services"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/transports"
	"go.uber.org/dig"
)

// placeholderServer is never dialed: every call sends the request to the agent's own URL.
const placeholderServer = "http://agent.invalid"

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

	if err := container.Provide(newAgentNodeClient); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewAgentCardRequester,
		dig.As(new(identity_directory_interfaces.AgentCardRequesterInterface)),
	); err != nil {
		panic(err)
	}
}

// newAgentNodeClient builds the generated gateway client over the guarded transport, redirects refused.
func newAgentNodeClient(agentCardSettings *settings.AgentCardSettings) (clients.ClientWithResponsesInterface, error) {
	return clients.NewClientWithResponses(
		placeholderServer,
		clients.WithHTTPClient(&http.Client{
			Timeout:   agentCardSettings.Timeout,
			Transport: transports.NewPublicRoundTripper(agentCardSettings),
			// A redirect could point anywhere the agent's own URL is not.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}),
		clients.WithRequestEditorFn(func(_ context.Context, request *http.Request) error {
			request.Header.Set("Accept", "application/json")
			return nil
		}),
	)
}
