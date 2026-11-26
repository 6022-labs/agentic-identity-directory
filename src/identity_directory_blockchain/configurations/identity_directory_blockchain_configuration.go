package configurations

import (
	identity_directory_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory/services/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/factory"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/metrics/interfaces"
	blockchain_services "github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	"go.uber.org/dig"
)

func AddIdentityDirectoryBlockchainConfiguration(container *dig.Container) {
	if err := container.Provide(blockchain_settings.NewChainsSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(blockchain_settings.NewAgentContractsSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewBlockchainRpcMetrics,
		dig.As(new(metrics_interfaces.BlockchainRpcMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(factory.NewEthClientFactory); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewBlockchainRequester,
		dig.As(new(identity_directory_interfaces.BlockchainRequesterInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewAgentCollectionsManagerRequester,
		dig.As(new(identity_directory_interfaces.AgentCollectionsManagerRequesterInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewAgentCollectionRequester,
		dig.As(new(identity_directory_interfaces.AgentCollectionRequesterInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewAgentEnsRegistryRequester,
		dig.As(new(identity_directory_interfaces.AgentEnsRegistryRequesterInterface)),
	); err != nil {
		panic(err)
	}
}
