package repositories_test

import (
	"context"
	"testing"

	domain_repositories "github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenListingCollectionsTestingSuite struct {
	*databaseTestingSuite
	sut *repositories.AgentCollectionRepository
}

func WhenListingCollectionsBeforeEach() *WhenListingCollectionsTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenListingCollectionsTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewAgentCollectionRepository(databaseTestingSuite.dbClient),
	}
}

func TestWhenListingCollections(t *testing.T) {
	t.Parallel()

	t.Run("Given collections on two chains", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list one chain's collections by name with their total", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingCollectionsBeforeEach()
			defer suite.close()
			traders := suite.seedCollection(model_builders.NewAgentCollectionBuilder().WithName("traders").Build())
			scouts := suite.seedCollection(model_builders.NewAgentCollectionBuilder().WithName("scouts").Build())
			suite.seedCollection(model_builders.NewAgentCollectionBuilder().WithChainId(137).Build())

			collections, total, err := suite.sut.List(
				context.Background(),
				&domain_repositories.AgentCollectionListParams{
					ChainId: 80002, Offset: 0, Limit: 10,
				},
			)

			require.NoError(t, err)
			assert.Equal(t, int64(2), total)
			require.Len(t, collections, 2)
			assert.Equal(t, scouts.Address, collections[0].Address)
			assert.Equal(t, traders.Address, collections[1].Address)
		})
	})
}
