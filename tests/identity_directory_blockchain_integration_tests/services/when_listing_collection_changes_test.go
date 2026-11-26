package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenListingCollectionChangesTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenListingCollectionChangesBeforeEach(t *testing.T) *WhenListingCollectionChangesTestingSuite {
	fixture := newChainFixture(t)
	return &WhenListingCollectionChangesTestingSuite{fixture: fixture, sut: fixture.agentCollectionRequester()}
}

func TestWhenListingCollectionChanges(t *testing.T) {
	t.Parallel()

	t.Run("Given no collection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should query nothing and return no change", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingCollectionChangesBeforeEach(t)

			changes, err := suite.sut.ListChanges(
				context.Background(),
				anvilChainId,
				nil,
				0,
				suite.fixture.currentBlock(),
			)

			require.NoError(t, err)
			assert.Empty(t, changes.AgentRefs)
			assert.Empty(t, changes.CollectionAddresses)
		})
	})

	t.Run("Given agents minted, edited and transferred", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return each touched agent and collection once", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingCollectionChangesBeforeEach(t)
			collection, err := suite.fixture.stack.CreateCollection("scouts")
			suite.fixture.must(err)
			fromBlock := suite.fixture.currentBlock() + 1
			suite.fixture.must(suite.fixture.stack.MintAgent(collection, "first", common_tests.RandomEthAddress()))
			suite.fixture.must(suite.fixture.stack.MintAgent(collection, "second", common_tests.RandomEthAddress()))
			suite.fixture.must(suite.fixture.stack.SetAttribute(collection, 1, "mission", "scouting"))
			suite.fixture.must(suite.fixture.stack.Transfer(collection, 2, common_tests.RandomEthAddress()))
			suite.fixture.must(suite.fixture.stack.UpdateDescription(collection, "better scouts"))
			collectionAddress := strings.ToLower(collection.Hex())

			changes, err := suite.sut.ListChanges(
				context.Background(),
				anvilChainId,
				[]string{collectionAddress},
				fromBlock,
				suite.fixture.currentBlock(),
			)

			require.NoError(t, err)
			assert.Equal(t, []models.AgentRef{
				{CollectionAddress: collectionAddress, TokenId: "1"},
				{CollectionAddress: collectionAddress, TokenId: "2"},
			}, changes.AgentRefs)
			assert.Equal(t, []string{collectionAddress}, changes.CollectionAddresses)
		})
	})

	t.Run("Given more collections than one log query takes", func(t *testing.T) {
		t.Parallel()

		t.Run("Should query them in batches", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingCollectionChangesBeforeEach(t)
			first, err := suite.fixture.stack.CreateCollection("scouts")
			suite.fixture.must(err)
			second, err := suite.fixture.stack.CreateCollection("traders")
			suite.fixture.must(err)
			fromBlock := suite.fixture.currentBlock() + 1
			suite.fixture.must(suite.fixture.stack.MintAgent(first, "scout", common_tests.RandomEthAddress()))
			suite.fixture.must(suite.fixture.stack.MintAgent(second, "trader", common_tests.RandomEthAddress()))
			addresses := []string{strings.ToLower(first.Hex()), strings.ToLower(second.Hex())}

			changes, err := suite.fixture.agentCollectionRequesterThrough(suite.fixture.rpcUrl, 1).
				ListChanges(context.Background(), anvilChainId, addresses, fromBlock, suite.fixture.currentBlock())

			require.NoError(t, err)
			assert.Len(t, changes.AgentRefs, 2)
		})
	})

	t.Run("Given a node that refuses large block ranges", func(t *testing.T) {
		t.Parallel()

		t.Run("Should split the range until the node answers", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingCollectionChangesBeforeEach(t)
			collection, err := suite.fixture.stack.CreateCollection("scouts")
			suite.fixture.must(err)
			fromBlock := suite.fixture.currentBlock() + 1
			suite.fixture.must(suite.fixture.stack.MintAgent(collection, "first", common_tests.RandomEthAddress()))
			suite.fixture.must(suite.fixture.stack.MintAgent(collection, "second", common_tests.RandomEthAddress()))
			suite.fixture.must(suite.fixture.stack.MintAgent(collection, "third", common_tests.RandomEthAddress()))
			rpcUrl, refusals := suite.fixture.rangeRefusingRpcUrl(1)

			changes, err := suite.fixture.agentCollectionRequesterThrough(rpcUrl, 10).ListChanges(
				context.Background(),
				anvilChainId, []string{strings.ToLower(collection.Hex())}, fromBlock, suite.fixture.currentBlock())

			require.NoError(t, err)
			assert.Len(t, changes.AgentRefs, 3)
			assert.Positive(t, refusals.Load())
		})
	})
}
