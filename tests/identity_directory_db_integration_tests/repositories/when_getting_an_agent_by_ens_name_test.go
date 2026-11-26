package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenGettingAnAgentByEnsNameTestingSuite struct {
	*databaseTestingSuite
	sut *repositories.AgentRepository
}

func WhenGettingAnAgentByEnsNameBeforeEach() *WhenGettingAnAgentByEnsNameTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenGettingAnAgentByEnsNameTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewAgentRepository(databaseTestingSuite.dbClient),
	}
}

func TestWhenGettingAnAgentByEnsName(t *testing.T) {
	t.Parallel()

	t.Run("Given no agent has the name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nil without an error", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingAnAgentByEnsNameBeforeEach()
			defer suite.close()

			agent, err := suite.sut.GetByEnsName(context.Background(), "ghost.agents.80002.6022.eth")

			require.NoError(t, err)
			assert.Nil(t, agent)
		})
	})

	t.Run("Given an indexed agent with its collection and records", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it with its collection, addresses and records", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingAnAgentByEnsNameBeforeEach()
			defer suite.close()
			collection := suite.seedCollection(model_builders.NewAgentCollectionBuilder().Build())
			indexed := suite.seedAgent(
				model_builders.NewAgentBuilder().WithCollectionAddress(collection.Address).Build(),
			)
			suite.seedEnsRecord(indexed, "url", "https://scout.example")

			agent, err := suite.sut.GetByEnsName(context.Background(), indexed.EnsName)

			require.NoError(t, err)
			assert.Equal(t, indexed.TokenId, agent.TokenId)
			assert.Equal(t, collection.Name, agent.Collection.Name)
			assert.Equal(t, indexed.Addresses, agent.Addresses)
			assert.Equal(t, "https://scout.example", agent.Url())
		})
	})
}
