package identity_directory_db

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/settings"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

const migrationsDialect = "postgres"

// goose forbids concurrent migrations on the same DB.
var gooseMigrationMu sync.Mutex

type DatabaseClient struct {
	db                 *gorm.DB
	migrationsSettings *settings.MigrationsSettings
}

func NewDatabaseClient(
	databaseSettings *settings.DatabaseSettings,
	migrationsSettings *settings.MigrationsSettings,
	logger *zap.Logger,
) (*DatabaseClient, error) {
	gormLogger := zapgorm2.New(logger)
	gormLogger.IgnoreRecordNotFoundError = true

	db, err := gorm.Open(postgres.Open(databaseSettings.Url), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, fmt.Errorf("failed to open identity directory database: %w", err)
	}

	return &DatabaseClient{
		db:                 db,
		migrationsSettings: migrationsSettings,
	}, nil
}

// Use registers a gorm plugin at bootstrap, where there is no request context.
func (c *DatabaseClient) Use(plugin gorm.Plugin) error {
	return c.db.Use(plugin)
}

// GetDB returns the transaction the context carries, or the pooled connection.
func (c *DatabaseClient) GetDB(ctx context.Context) *gorm.DB {
	if transaction, ok := ctx.Value(transactionKey{}).(*Transaction); ok && !transaction.closed {
		return transaction.tx.WithContext(ctx)
	}
	return c.db.WithContext(ctx)
}

func (c *DatabaseClient) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (c *DatabaseClient) Migrate() error {
	gooseMigrationMu.Lock()
	defer gooseMigrationMu.Unlock()

	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	if err := goose.SetDialect(migrationsDialect); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}
	folder := filepath.Join(c.migrationsSettings.FolderPath, migrationsDialect)
	if err := goose.Up(sqlDB, folder); err != nil {
		return fmt.Errorf("goose migration failed: %w", err)
	}
	return nil
}
