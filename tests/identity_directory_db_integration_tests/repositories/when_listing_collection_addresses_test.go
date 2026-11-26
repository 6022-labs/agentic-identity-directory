package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_db/repositories"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenListingCollectionAddressesTestingSuite struct {
	*databaseTestingSuite
	sut *repositories.AgentCollectionRepository
}

func WhenListingCollectionAddressesBeforeEach() *WhenListingCollectionAddressesTestingSuite {
	databaseTestingSuite := newDatabaseTestingSuite()
	return &WhenListingCollectionAddressesTestingSuite{
		databaseTestingSuite: databaseTestingSuite,
		sut:                  repositories.NewAgentCollectionRepository(databaseTestingSuite.dbClient),
	}
}

func TestWhenListingCollectionAddresses(t *testing.T) {
	t.Parallel()

	t.Run("Given collections on two chains", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return every address of the chain", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingCollectionAddressesBeforeEach()
			defer suite.close()
			scouts := suite.seedCollection(model_builders.NewAgentCollectionBuilder().Build())
			traders := suite.seedCollection(model_builders.NewAgentCollectionBuilder().Build())
			suite.seedCollection(model_builders.NewAgentCollectionBuilder().WithChainId(137).Build())

			addresses, err := suite.sut.ListAddresses(context.Background(), 80002)

			require.NoError(t, err)
			assert.ElementsMatch(t, []string{scouts.Address, traders.Address}, addresses)
		})
	})
}
