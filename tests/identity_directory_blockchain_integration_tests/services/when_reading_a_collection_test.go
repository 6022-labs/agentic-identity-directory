package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenReadingACollectionTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenReadingACollectionBeforeEach(t *testing.T) *WhenReadingACollectionTestingSuite {
	fixture := newChainFixture(t)
	return &WhenReadingACollectionTestingSuite{fixture: fixture, sut: fixture.agentCollectionRequester()}
}

func TestWhenReadingACollection(t *testing.T) {
	t.Parallel()

	t.Run("Given a collection whose description changed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should read its name, symbol and current description", func(t *testing.T) {
			t.Parallel()
			suite := WhenReadingACollectionBeforeEach(t)
			collection, err := suite.fixture.stack.CreateCollection("scouts")
			suite.fixture.must(err)
			suite.fixture.must(suite.fixture.stack.UpdateDescription(collection, "better scouts"))
			collectionAddress := strings.ToLower(collection.Hex())

			agentCollection, err := suite.sut.GetCollection(
				context.Background(), anvilChainId, collectionAddress, suite.fixture.currentBlock())

			require.NoError(t, err)
			assert.Equal(t, &models.AgentCollection{
				ChainId: anvilChainId, Address: collectionAddress, Name: "scouts", Symbol: "SCOUTS",
				Description: "better scouts",
			}, agentCollection)
		})
	})
}
