package repositories_test

import (
	"context"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
)

type databaseTestingSuite struct {
	dbName   string
	dbClient *identity_directory_db.DatabaseClient
}

func newDatabaseTestingSuite() *databaseTestingSuite {
	dbName, dbClient := common_tests.NewMigratedDatabaseClient()
	return &databaseTestingSuite{dbName: dbName, dbClient: dbClient}
}

func (s *databaseTestingSuite) close() {
	_ = s.dbClient.Close()
	container_builders.GetPostgresContainerBuilder().DropDatabase(s.dbName)
}

func (s *databaseTestingSuite) seed(model any) {
	if err := s.dbClient.GetDB(context.Background()).Create(model).Error; err != nil {
		panic(err)
	}
}

func (s *databaseTestingSuite) seedCollection(collection *models.AgentCollection) *models.AgentCollection {
	s.seed(collection)
	return collection
}

// seedAgent stores the agent and its addresses, the way an indexed agent is stored.
func (s *databaseTestingSuite) seedAgent(agent *models.Agent) *models.Agent {
	s.seed(agent)
	return agent
}

func (s *databaseTestingSuite) seedEnsRecord(agent *models.Agent, key, value string) {
	s.seed(&models.AgentEnsRecord{
		ChainId: agent.ChainId, CollectionAddress: agent.CollectionAddress, TokenId: agent.TokenId,
		Key: key, Value: value,
	})
}

func (s *databaseTestingSuite) seedScoutAndTrader(collection *models.AgentCollection) (*models.Agent, *models.Agent) {
	scout := s.seedAgent(model_builders.NewAgentBuilder().
		WithCollectionAddress(collection.Address).
		WithName("scout").
		WithEnsName("scout.markets.80002.6022.eth").
		Build())
	trader := s.seedAgent(model_builders.NewAgentBuilder().
		WithCollectionAddress(collection.Address).
		WithTokenId("2").
		WithName("trader").
		WithEnsName("trader.markets.80002.6022.eth").
		WithAttributes(map[string]string{"role": "facilitator", "mission": "trade_bonds"}).
		Build())
	return scout, trader
}
