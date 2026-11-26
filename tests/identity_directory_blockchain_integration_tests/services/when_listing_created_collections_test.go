package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenListingCreatedCollectionsTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionsManagerRequester
}

func WhenListingCreatedCollectionsBeforeEach(t *testing.T) *WhenListingCreatedCollectionsTestingSuite {
	fixture := newChainFixture(t)
	return &WhenListingCreatedCollectionsTestingSuite{fixture: fixture, sut: fixture.agentCollectionsManagerRequester()}
}

func TestWhenListingCreatedCollections(t *testing.T) {
	t.Parallel()

	t.Run("Given two collections created in the range", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return their lowercase addresses in creation order", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingCreatedCollectionsBeforeEach(t)
			fromBlock := suite.fixture.currentBlock()
			first, err := suite.fixture.stack.CreateCollection("scouts")
			suite.fixture.must(err)
			second, err := suite.fixture.stack.CreateCollection("traders")
			suite.fixture.must(err)

			addresses, err := suite.sut.ListCreatedCollections(
				context.Background(),
				anvilChainId,
				fromBlock,
				suite.fixture.currentBlock(),
			)

			require.NoError(t, err)
			assert.Equal(t, []string{strings.ToLower(first.Hex()), strings.ToLower(second.Hex())}, addresses)
		})
	})
}
