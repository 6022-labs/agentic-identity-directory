package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenListingAgentRefsTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenListingAgentRefsBeforeEach(t *testing.T) *WhenListingAgentRefsTestingSuite {
	fixture := newChainFixture(t)
	return &WhenListingAgentRefsTestingSuite{fixture: fixture, sut: fixture.agentCollectionRequester()}
}

func TestWhenListingAgentRefs(t *testing.T) {
	t.Parallel()

	t.Run("Given a collection holding no agent", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return no token", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingAgentRefsBeforeEach(t)
			collection, err := suite.fixture.stack.CreateCollection("scouts")
			suite.fixture.must(err)

			agentRefs, err := suite.sut.ListAgentRefs(
				context.Background(), anvilChainId, strings.ToLower(collection.Hex()), suite.fixture.currentBlock())

			require.NoError(t, err)
			assert.Empty(t, agentRefs)
		})
	})

	t.Run("Given a collection holding two agents", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list both tokens", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingAgentRefsBeforeEach(t)
			collection, err := suite.fixture.stack.CreateCollection("scouts")
			suite.fixture.must(err)
			suite.fixture.must(suite.fixture.stack.MintAgent(collection, "first", common_tests.RandomEthAddress()))
			suite.fixture.must(suite.fixture.stack.MintAgent(collection, "second", common_tests.RandomEthAddress()))
			collectionAddress := strings.ToLower(collection.Hex())

			agentRefs, err := suite.sut.ListAgentRefs(
				context.Background(), anvilChainId, collectionAddress, suite.fixture.currentBlock())

			require.NoError(t, err)
			assert.Equal(t, []models.AgentRef{
				{CollectionAddress: collectionAddress, TokenId: "1"},
				{CollectionAddress: collectionAddress, TokenId: "2"},
			}, agentRefs)
		})
	})
}
