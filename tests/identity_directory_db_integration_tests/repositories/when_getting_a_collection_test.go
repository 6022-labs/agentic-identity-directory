package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenGettingACollectionTestingSuite struct {
	*databaseTestingSuite
	sut *repositories.AgentCollectionRepository
}

func WhenGettingACollectionBeforeEach() *WhenGettingACollectionTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenGettingACollectionTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewAgentCollectionRepository(databaseTestingSuite.dbClient),
	}
}

func TestWhenGettingACollection(t *testing.T) {
	t.Parallel()

	t.Run("Given a collection that is not indexed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nil without an error", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingACollectionBeforeEach()
			defer suite.close()

			collection, err := suite.sut.Get(context.Background(), 80002, common_tests.RandomLowerEthAddress())

			require.NoError(t, err)
			assert.Nil(t, collection)
		})
	})

	t.Run("Given an indexed collection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingACollectionBeforeEach()
			defer suite.close()
			indexed := suite.seedCollection(model_builders.NewAgentCollectionBuilder().WithName("scouts").Build())

			collection, err := suite.sut.Get(context.Background(), indexed.ChainId, indexed.Address)

			require.NoError(t, err)
			assert.Equal(t, "scouts", collection.Name)
		})
	})
}
