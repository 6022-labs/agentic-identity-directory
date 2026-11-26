package configurations

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_daemons/interfaces"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/dig"
)

type configureDaemonsParams struct {
	dig.In

	App     *fiber.App
	Daemons []interfaces.DaemonInterface `group:"daemons"`
}

// ConfigureDaemons starts the daemons and stops them before Fiber shuts down.
func ConfigureDaemons(container *dig.Container) {
	err := container.Invoke(func(p configureDaemonsParams) {
		for _, daemon := range p.Daemons {
			daemon.Start(context.Background())
		}
		p.App.Hooks().OnPreShutdown(func() error {
			for _, daemon := range p.Daemons {
				daemon.Stop()
			}
			return nil
		})
	})
	if err != nil {
		panic(err)
	}
}
