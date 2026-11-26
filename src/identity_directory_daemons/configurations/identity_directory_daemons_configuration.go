package configurations

import (
	identity_directory_daemons "github.com/6022-labs/agentic-identity-directory/src/identity_directory_daemons"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_daemons/interfaces"
	"go.uber.org/dig"
)

func AddIdentityDirectoryDaemonsConfiguration(container *dig.Container) {
	if err := container.Provide(
		identity_directory_daemons.NewChainSyncDaemon,
		dig.Group("daemons"),
		dig.As(new(interfaces.DaemonInterface)),
	); err != nil {
		panic(err)
	}
}
