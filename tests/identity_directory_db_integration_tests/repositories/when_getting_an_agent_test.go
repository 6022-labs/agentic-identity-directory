package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenGettingAnAgentTestingSuite struct {
	*databaseTestingSuite
	sut *repositories.AgentRepository
}

func WhenGettingAnAgentBeforeEach() *WhenGettingAnAgentTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenGettingAnAgentTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewAgentRepository(databaseTestingSuite.dbClient),
	}
}

func TestWhenGettingAnAgent(t *testing.T) {
	t.Parallel()

	t.Run("Given the token on another chain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nil without an error", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingAnAgentBeforeEach()
			defer suite.close()
			indexed := suite.seedAgent(model_builders.NewAgentBuilder().Build())

			agent, err := suite.sut.Get(context.Background(), 137,
				models.AgentRef{CollectionAddress: indexed.CollectionAddress, TokenId: indexed.TokenId})

			require.NoError(t, err)
			assert.Nil(t, agent)
		})
	})

	t.Run("Given an indexed agent", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it with its addresses", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingAnAgentBeforeEach()
			defer suite.close()
			indexed := suite.seedAgent(model_builders.NewAgentBuilder().WithTokenId("42").Build())

			agent, err := suite.sut.Get(context.Background(), indexed.ChainId,
				models.AgentRef{CollectionAddress: indexed.CollectionAddress, TokenId: "42"})

			require.NoError(t, err)
			assert.Equal(t, indexed.Name, agent.Name)
			assert.Equal(t, indexed.Addresses, agent.Addresses)
		})
	})
}
