package configurations

import (
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"
	identity_directory_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory/services/interfaces"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/metrics/interfaces"
	db_repositories "github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	db_services "github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/services"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/settings"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func AddIdentityDirectoryDbConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewDatabaseSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(settings.NewMigrationsSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewDbQueryMetrics,
		dig.As(new(metrics_interfaces.DbQueryMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(newDatabaseClient); err != nil {
		panic(err)
	}

	if err := container.Provide(
		db_services.NewUnitOfWorkFactory,
		dig.As(new(identity_directory_interfaces.UnitOfWorkFactoryInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		db_repositories.NewAgentCollectionRepository,
		dig.As(new(repositories.AgentCollectionRepositoryInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		db_repositories.NewAgentRepository,
		dig.As(new(repositories.AgentRepositoryInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		db_repositories.NewAgentEnsRecordRepository,
		dig.As(new(repositories.AgentEnsRecordRepositoryInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		db_repositories.NewSyncCursorRepository,
		dig.As(new(repositories.SyncCursorRepositoryInterface)),
	); err != nil {
		panic(err)
	}
}

// newDatabaseClient opens the DB and migrates it at boot; nothing can serve without it.
func newDatabaseClient(
	databaseSettings *settings.DatabaseSettings,
	migrationsSettings *settings.MigrationsSettings,
	logger *zap.Logger,
	dbQueryMetrics metrics_interfaces.DbQueryMetricsInterface,
) *identity_directory_db.DatabaseClient {
	client, err := identity_directory_db.NewDatabaseClient(databaseSettings, migrationsSettings, logger)
	if err != nil {
		logger.Fatal("failed to create the identity directory database client", zap.Error(err))
	}
	if err := client.Use(metrics.NewGormMetricsPlugin(dbQueryMetrics)); err != nil {
		logger.Fatal("failed to register db metrics plugin", zap.Error(err))
	}
	if err := client.Migrate(); err != nil {
		logger.Fatal("failed to migrate the identity directory database", zap.Error(err))
	}
	return client
}
