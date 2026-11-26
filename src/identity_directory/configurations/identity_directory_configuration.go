package configurations

import (
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory/metrics/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"go.uber.org/dig"
)

func AddIdentityDirectoryConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewSyncSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(settings.NewRateLimitSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(settings.NewProxySettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewSyncMetrics,
		dig.As(new(metrics_interfaces.SyncMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewEnsNameBuilder,
		dig.As(new(interfaces.EnsNameBuilderInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewChainWindowReader,
		dig.As(new(interfaces.ChainWindowReaderInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewChainWindowWriter,
		dig.As(new(interfaces.ChainWindowWriterInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewAgentCardSignatureVerifier,
		dig.As(new(interfaces.AgentCardSignatureVerifierInterface)),
	); err != nil {
		panic(err)
	}

	for _, useCase := range []any{
		use_cases.NewSyncChain,
		use_cases.NewSearchAgents,
		use_cases.NewGetAgent,
		use_cases.NewGetAgentByEnsName,
		use_cases.NewGetAgentCard,
		use_cases.NewListCollections,
		use_cases.NewGetHealth,
	} {
		if err := container.Provide(useCase); err != nil {
			panic(err)
		}
	}
}
