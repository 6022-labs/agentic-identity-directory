package configurations

import (
	"github.com/6022-labs/agentic-identity-directory/src/common/settings"
	identity_directory_configurations "github.com/6022-labs/agentic-identity-directory/src/identity_directory/configurations"
	blockchain_configurations "github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/configurations"
	daemons_configurations "github.com/6022-labs/agentic-identity-directory/src/identity_directory_daemons/configurations"
	db_configurations "github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/configurations"
	agent_card_configurations "github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/configurations"
	mcp_configurations "github.com/6022-labs/agentic-identity-directory/src/identity_directory_mcp/configurations"
	mvc_configurations "github.com/6022-labs/agentic-identity-directory/src/identity_directory_mvc/configurations"
	"github.com/knadh/koanf/v2"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

const (
	hostSettingsKey = "host"
	defaultApiPort  = 5000
)

func ConfigureDI(config *koanf.Koanf) *dig.Container {
	container := dig.New()

	if err := container.Provide(func(logger *zap.Logger, k *koanf.Koanf) *settings.HostSettings {
		return settings.NewHostSettings(logger, k, hostSettingsKey, defaultApiPort)
	}); err != nil {
		panic(err)
	}
	if err := container.Provide(func() *koanf.Koanf {
		return config
	}); err != nil {
		panic(err)
	}

	db_configurations.AddIdentityDirectoryDbConfiguration(container)
	blockchain_configurations.AddIdentityDirectoryBlockchainConfiguration(container)
	agent_card_configurations.AddIdentityDirectoryHttpAgentCardConfiguration(container)
	identity_directory_configurations.AddIdentityDirectoryConfiguration(container)
	daemons_configurations.AddIdentityDirectoryDaemonsConfiguration(container)
	mcp_configurations.AddIdentityDirectoryMcpConfiguration(container)
	mvc_configurations.AddIdentityDirectoryMvcConfiguration(container)

	return container
}
