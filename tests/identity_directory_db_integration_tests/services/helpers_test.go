package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db"
	db_services "github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/services"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

type unitOfWorkTestingSuite struct {
	dbName   string
	dbClient *identity_directory_db.DatabaseClient
	sut      *db_services.UnitOfWorkFactory
}

func newUnitOfWorkTestingSuite() *unitOfWorkTestingSuite {
	dbName, dbClient := common_tests.NewMigratedDatabaseClient()
	return &unitOfWorkTestingSuite{
		dbName:   dbName,
		dbClient: dbClient,
		sut:      db_services.NewUnitOfWorkFactory(dbClient),
	}
}

func (s *unitOfWorkTestingSuite) close() {
	_ = s.dbClient.Close()
	container_builders.GetPostgresContainerBuilder().DropDatabase(s.dbName)
}

func (s *unitOfWorkTestingSuite) createCollection(t *testing.T, ctx context.Context) *models.AgentCollection {
	t.Helper()

	collection := model_builders.NewAgentCollectionBuilder().Build()
	assert.NoError(t, s.dbClient.GetDB(ctx).Create(collection).Error)
	return collection
}

func (s *unitOfWorkTestingSuite) collectionExists(t *testing.T, collection *models.AgentCollection) bool {
	t.Helper()

	var count int64
	assert.NoError(t, s.dbClient.GetDB(context.Background()).Model(&models.AgentCollection{}).
		Where("chain_id = ? AND address = ?", collection.ChainId, collection.Address).
		Count(&count).Error)
	return count == 1
}
