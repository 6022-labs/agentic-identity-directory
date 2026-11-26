package migrations_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/migrations/schema"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/settings"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/container_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenMigratingMigrationsTestingSuite struct {
	dbName string
	sut    *identity_directory_db.DatabaseClient
}

func WhenMigratingMigrationsBeforeEach() *WhenMigratingMigrationsTestingSuite {
	dbName, connectionString := container_builders.GetPostgresContainerBuilder().Create()

	migrationFolderPath, err := common_tests.FindMigrationFolderPath()
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

	return &WhenMigratingMigrationsTestingSuite{dbName: dbName, sut: databaseClient}
}

func WhenMigratingMigrationsAfterEach(suite *WhenMigratingMigrationsTestingSuite) {
	_ = suite.sut.Close()
	container_builders.GetPostgresContainerBuilder().DropDatabase(suite.dbName)
}

func TestWhenMigratingMigrations(t *testing.T) {
	t.Parallel()

	t.Run("Given no migrations are applied", func(t *testing.T) {
		t.Parallel()

		t.Run("Should apply all migrations without error", func(t *testing.T) {
			t.Parallel()
			suite := WhenMigratingMigrationsBeforeEach()
			defer WhenMigratingMigrationsAfterEach(suite)

			assert.NoError(t, suite.sut.Migrate())
		})
	})

	t.Run("Given every migration is applied", func(t *testing.T) {
		t.Parallel()

		// A model with no migration boots fine and fails on first query.
		t.Run("Should create a table for every registered model", func(t *testing.T) {
			t.Parallel()
			suite := WhenMigratingMigrationsBeforeEach()
			defer WhenMigratingMigrationsAfterEach(suite)

			assert.NoError(t, suite.sut.Migrate())

			migrator := suite.sut.GetDB(context.Background()).Migrator()
			for _, model := range schema.Models() {
				assert.Truef(t, migrator.HasTable(model), "no table for %T", model)
			}
		})
	})

	t.Run("Given migrations are already applied", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be idempotent on a second run", func(t *testing.T) {
			t.Parallel()
			suite := WhenMigratingMigrationsBeforeEach()
			defer WhenMigratingMigrationsAfterEach(suite)

			assert.NoError(t, suite.sut.Migrate())
			assert.NoError(t, suite.sut.Migrate())
		})
	})
}
