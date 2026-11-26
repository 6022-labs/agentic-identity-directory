package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/6022-labs/agentic-identity-directory/src/common/host_configurations"
	"github.com/6022-labs/agentic-identity-directory/src/common/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_host/configurations"
	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// @title Agentic Identity Directory API
// @version 1.0
// @description Directory of the 6022 agents registered on-chain, also served over MCP at /mcp.
// @BasePath /
func main() {
	_ = godotenv.Load(".env")

	config, err := host_configurations.LoadKoanfConfig()
	if err != nil {
		panic(err)
	}

	container := configurations.ConfigureDI(config)
	host_configurations.ConfigureLogging(container)

	shutdownTelemetry := host_configurations.ConfigureTelemetry(container, configurations.AppName)
	defer func() { _ = shutdownTelemetry(context.Background()) }()

	configurations.ConfigureServer(container)
	configurations.ConfigureDaemons(container)

	err = container.Invoke(func(app *fiber.App, logger *zap.Logger, hostSettings *settings.HostSettings) error {
		bind := fmt.Sprintf("%s:%d", hostSettings.ListenAddress, hostSettings.ApiPort)
		logger.Info("Starting identity directory server", zap.String("bind", bind))

		serverErrCh := make(chan error, 1)
		go func() {
			serverErrCh <- app.Listen(bind, fiber.ListenConfig{ListenerNetwork: "tcp"})
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

			if err := app.Shutdown(); err != nil {
				logger.Warn("Fiber shutdown returned an error", zap.Error(err))
			}

			if err := <-serverErrCh; err != nil {
				logger.Error("Server stopped with error", zap.Error(err))
			} else {
				logger.Info("Server stopped")
			}

			logger.Info("Shutdown complete")
			_ = logger.Sync()
			return nil

		case err := <-serverErrCh:
			if err != nil {
				return fmt.Errorf("server exited unexpectedly: %w", err)
			}
			return nil
		}
	})
	if err != nil {
		panic(err)
	}
}
