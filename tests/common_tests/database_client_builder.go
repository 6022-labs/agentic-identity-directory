package common_tests

import (
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/settings"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/container_builders"
	"go.uber.org/zap"
)

// NewMigratedDatabaseClient opens a migrated scratch database; the caller closes it and drops dbName.
func NewMigratedDatabaseClient() (string, *identity_directory_db.DatabaseClient) {
	dbName, connectionString := container_builders.GetPostgresContainerBuilder().Create()

	migrationFolderPath, err := FindMigrationFolderPath()
	if err != nil {
		panic(err)
	}

	databaseClient, err := identity_directory_db.NewDatabaseClient(
		&settings.DatabaseSettings{Url: connectionString},
		&settings.MigrationsSettings{FolderPath: migrationFolderPath},
		zap.NewNop(),
	)
	if err != nil {
		panic(err)
	}
	if err := databaseClient.Migrate(); err != nil {
		panic(err)
	}
	return dbName, databaseClient
}
