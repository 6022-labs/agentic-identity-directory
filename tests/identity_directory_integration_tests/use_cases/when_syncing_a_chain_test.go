package use_cases_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/metrics"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/factory"
	blockchain_metrics "github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/metrics"
	blockchain_services "github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	db_repositories "github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	db_services "github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/services"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_blockchain_integration_tests/fixtures"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const anvilChainId = uint64(31337)

type WhenSyncingAChainTestingSuite struct {
	stack           *fixtures.AgentStack
	agentRepository *db_repositories.AgentRepository
	sut             *use_cases.SyncChain
}

// WhenSyncingAChainBeforeEach wires the real adapters on a fresh stack and database, from the stack's deployment.
func WhenSyncingAChainBeforeEach(t *testing.T) *WhenSyncingAChainTestingSuite {
	ctx := context.Background()
	anvilAccountSettings := container_builders.GetAnvilContainerBuilder().Create()
	client, err := ethclient.Dial(anvilAccountSettings.HttpUrl)
	require.NoError(t, err)
	t.Cleanup(client.Close)
	startBlock, err := client.BlockNumber(ctx)
	require.NoError(t, err)
	stack, err := fixtures.DeployAgentStack(ctx, client, anvilAccountSettings)
	require.NoError(t, err)

	dbName, databaseClient := common_tests.NewMigratedDatabaseClient()
	t.Cleanup(func() {
		_ = databaseClient.Close()
		container_builders.GetPostgresContainerBuilder().DropDatabase(dbName)
	})

	chainsSettings := &blockchain_settings.ChainsSettings{anvilChainId: {
		RpcHttpUrl: anvilAccountSettings.HttpUrl, RequestTimeout: 5 * time.Second, MaxAddressesPerLogQuery: 10,
	}}
	agentContractsSettings := &blockchain_settings.AgentContractsSettings{anvilChainId: {
		AgentCollectionsManager: stack.Manager, AgentEnsRegistry: stack.AgentEnsRegistry,
	}}
	ethClientFactory := factory.NewEthClientFactory(chainsSettings, blockchain_metrics.NewBlockchainRpcMetrics())
	managerRequester, err := blockchain_services.NewAgentCollectionsManagerRequester(
		zap.NewNop(),
		ethClientFactory,
		chainsSettings,
		agentContractsSettings,
	)
	require.NoError(t, err)
	collectionRequester, err := blockchain_services.NewAgentCollectionRequester(
		zap.NewNop(),
		ethClientFactory,
		chainsSettings,
	)
	require.NoError(t, err)
	ensRegistryRequester, err := blockchain_services.NewAgentEnsRegistryRequester(
		zap.NewNop(), ethClientFactory, chainsSettings, agentContractsSettings)
	require.NoError(t, err)

	collectionRepository := db_repositories.NewAgentCollectionRepository(databaseClient)
	agentRepository := db_repositories.NewAgentRepository(databaseClient)
	syncCursorRepository := db_repositories.NewSyncCursorRepository(databaseClient)
	syncSettings := &settings.SyncSettings{anvilChainId: {
		StartBlock: startBlock, Confirmations: 0, PollInterval: time.Second, LogWindowBlocks: 4,
	}}

	return &WhenSyncingAChainTestingSuite{
		stack:           stack,
		agentRepository: agentRepository,
		sut: use_cases.NewSyncChain(
			syncSettings,
			blockchain_services.NewBlockchainRequester(zap.NewNop(), ethClientFactory, chainsSettings),
			syncCursorRepository,
			services.NewChainWindowReader(zap.NewNop(), managerRequester, collectionRequester, ensRegistryRequester,
				collectionRepository, services.NewEnsNameBuilder()),
			services.NewChainWindowWriter(db_services.NewUnitOfWorkFactory(databaseClient), collectionRepository,
				agentRepository, db_repositories.NewAgentEnsRecordRepository(databaseClient), syncCursorRepository),
			metrics.NewSyncMetrics(),
		),
	}
}

func (s *WhenSyncingAChainTestingSuite) sync(t *testing.T) {
	require.NoError(t, s.sut.Execute(context.Background(), &requests.SyncChainRequest{ChainId: anvilChainId}))
}

func (s *WhenSyncingAChainTestingSuite) agentNamed(t *testing.T, ensName string) *models.Agent {
	agent, err := s.agentRepository.GetByEnsName(context.Background(), ensName)
	require.NoError(t, err)
	require.NotNil(t, agent, ensName)
	return agent
}

func (s *WhenSyncingAChainTestingSuite) must(t *testing.T, err error) {
	require.NoError(t, err)
}

func TestWhenSyncingAChain(t *testing.T) {
	t.Parallel()

	t.Run("Given a first deployment over a chain that already holds agents", func(t *testing.T) {
		t.Parallel()

		t.Run("Should index them with their published url", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			collection, err := suite.stack.CreateCollection("scouts")
			suite.must(t, err)
			suite.must(t, suite.stack.MintAgent(collection, "scout", common_tests.RandomEthAddress()))
			suite.must(t, suite.stack.PublishUrl(collection, 1, "https://scout.example"))

			suite.sync(t)

			agent := suite.agentNamed(t, "scout.scouts.31337.6022.eth")
			assert.Equal(t, strings.ToLower(collection.Hex()), agent.CollectionAddress)
			assert.Equal(t, "https://scout.example", agent.Url())
			assert.Equal(t, "expert", agent.Role())
		})
	})

	t.Run("Given changes made while the service was down", func(t *testing.T) {
		t.Parallel()

		t.Run("Should catch up on edits and on a collection created meanwhile", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			scouts, err := suite.stack.CreateCollection("scouts")
			suite.must(t, err)
			suite.must(t, suite.stack.MintAgent(scouts, "scout", common_tests.RandomEthAddress()))
			suite.sync(t)
			newOwner := common_tests.RandomEthAddress()

			suite.must(t, suite.stack.SetAttribute(scouts, 1, "role", "facilitator"))
			suite.must(t, suite.stack.Transfer(scouts, 1, newOwner))
			traders, err := suite.stack.CreateCollection("traders")
			suite.must(t, err)
			suite.must(t, suite.stack.MintAgent(traders, "trader", common_tests.RandomEthAddress()))
			suite.must(t, suite.stack.PublishUrl(traders, 1, "https://trader.example"))
			suite.sync(t)

			scout := suite.agentNamed(t, "scout.scouts.31337.6022.eth")
			trader := suite.agentNamed(t, "trader.traders.31337.6022.eth")
			assert.Equal(t, "facilitator", scout.Role())
			assert.Equal(t, strings.ToLower(newOwner.Hex()), scout.Owner)
			assert.Equal(t, "https://trader.example", trader.Url())
		})
	})
}
