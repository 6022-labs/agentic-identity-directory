package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenUpsertingACollectionTestingSuite struct {
	*databaseTestingSuite
	sut *repositories.AgentCollectionRepository
}

func WhenUpsertingACollectionBeforeEach() *WhenUpsertingACollectionTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenUpsertingACollectionTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewAgentCollectionRepository(databaseTestingSuite.dbClient),
	}
}

func TestWhenUpsertingACollection(t *testing.T) {
	t.Parallel()

	t.Run("Given a collection already indexed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should replace its description", func(t *testing.T) {
			t.Parallel()
			suite := WhenUpsertingACollectionBeforeEach()
			defer suite.close()
			collection := suite.seedCollection(model_builders.NewAgentCollectionBuilder().Build())
			collection.Description = "better scouts"

			err := suite.sut.Upsert(context.Background(), collection)

			require.NoError(t, err)
			stored, err := suite.sut.Get(context.Background(), collection.ChainId, collection.Address)
			require.NoError(t, err)
			assert.Equal(t, "better scouts", stored.Description)
		})
	})
}
