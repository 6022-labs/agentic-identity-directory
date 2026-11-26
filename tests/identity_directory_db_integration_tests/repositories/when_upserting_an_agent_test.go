package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenUpsertingAnAgentTestingSuite struct {
	*databaseTestingSuite
	sut *repositories.AgentRepository
}

func WhenUpsertingAnAgentBeforeEach() *WhenUpsertingAnAgentTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenUpsertingAnAgentTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewAgentRepository(databaseTestingSuite.dbClient),
	}
}

func TestWhenUpsertingAnAgent(t *testing.T) {
	t.Parallel()

	t.Run("Given an agent not indexed yet", func(t *testing.T) {
		t.Parallel()

		t.Run("Should store it with its addresses", func(t *testing.T) {
			t.Parallel()
			suite := WhenUpsertingAnAgentBeforeEach()
			defer suite.close()
			ctx := context.Background()
			agent := model_builders.NewAgentBuilder().Build()

			err := suite.sut.Upsert(ctx, agent)

			require.NoError(t, err)
			stored, err := suite.sut.Get(ctx, agent.ChainId,
				models.AgentRef{CollectionAddress: agent.CollectionAddress, TokenId: agent.TokenId})
			require.NoError(t, err)
			assert.Equal(t, agent.Name, stored.Name)
			assert.Equal(t, agent.Addresses, stored.Addresses)
		})
	})

	t.Run("Given an agent already indexed with ENS records", func(t *testing.T) {
		t.Parallel()

		t.Run("Should replace its fields and addresses, keeping its ENS records", func(t *testing.T) {
			t.Parallel()
			suite := WhenUpsertingAnAgentBeforeEach()
			defer suite.close()
			ctx := context.Background()
			agent := suite.seedAgent(model_builders.NewAgentBuilder().Build())
			suite.seedEnsRecord(agent, "url", "https://scout.example")
			newAddress := common_tests.RandomLowerEthAddress()
			address := models.AgentAddress{
				ChainId: agent.ChainId, CollectionAddress: agent.CollectionAddress, TokenId: agent.TokenId,
				Type: "evm", Value: newAddress,
			}
			updated := model_builders.NewAgentBuilder().
				WithCollectionAddress(agent.CollectionAddress).
				WithOwner(common_tests.RandomLowerEthAddress()).
				WithAttributes(map[string]string{"role": "facilitator"}).
				WithAddresses(address, address).
				Build()

			err := suite.sut.Upsert(ctx, updated)

			require.NoError(t, err)
			stored, err := suite.sut.Get(ctx, agent.ChainId,
				models.AgentRef{CollectionAddress: agent.CollectionAddress, TokenId: agent.TokenId})
			require.NoError(t, err)
			assert.Equal(t, updated.Owner, stored.Owner)
			assert.Equal(t, map[string]string{"role": "facilitator"}, stored.Attributes)
			assert.Equal(t, []models.AgentAddress{address}, stored.Addresses)
			assert.Equal(t, "https://scout.example", stored.Url())
		})
	})
}
