package configurations

import (
	common_metrics "github.com/6022-labs/agentic-identity-directory/src/common/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-identity-directory/src/common/metrics/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/common/mvc"
	mvc_interfaces "github.com/6022-labs/agentic-identity-directory/src/common/mvc/interfaces"
	mvc_middlewares "github.com/6022-labs/agentic-identity-directory/src/common/mvc/middlewares"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_mvc"
	"go.uber.org/dig"
)

func AddIdentityDirectoryMvcConfiguration(container *dig.Container) {
	if err := container.Provide(func() metrics_interfaces.ApiRequestMetricsInterface {
		return common_metrics.NewApiRequestMetrics("agentic_identity_directory_api")
	}); err != nil {
		panic(err)
	}

	if err := container.Provide(mvc_middlewares.NewApiRequestMetricsMiddleware); err != nil {
		panic(err)
	}

	if err := container.Provide(mvc.NewErrorWriter); err != nil {
		panic(err)
	}

	provideController(container, identity_directory_mvc.NewHealthController)
	provideController(container, identity_directory_mvc.NewAgentController)
	provideController(container, identity_directory_mvc.NewCollectionController)
}

func provideController(container *dig.Container, ctor any) {
	if err := container.Provide(
		ctor,
		dig.Group("controllers"),
		dig.As(new(mvc_interfaces.ControllerInterface)),
	); err != nil {
		panic(err)
	}
}
