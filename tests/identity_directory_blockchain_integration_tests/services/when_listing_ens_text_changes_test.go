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

type WhenListingEnsTextChangesTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentEnsRegistryRequester
}

func WhenListingEnsTextChangesBeforeEach(t *testing.T) *WhenListingEnsTextChangesTestingSuite {
	fixture := newChainFixture(t)
	return &WhenListingEnsTextChangesTestingSuite{fixture: fixture, sut: fixture.agentEnsRegistryRequester()}
}

func TestWhenListingEnsTextChanges(t *testing.T) {
	t.Parallel()

	t.Run("Given an agent that published, set and cleared texts", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return every change in log order", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingEnsTextChangesBeforeEach(t)
			collection, err := suite.fixture.stack.CreateCollection("scouts")
			suite.fixture.must(err)
			suite.fixture.must(suite.fixture.stack.MintAgent(collection, "scout", common_tests.RandomEthAddress()))
			fromBlock := suite.fixture.currentBlock() + 1
			suite.fixture.must(suite.fixture.stack.PublishUrl(collection, 1, "https://scout.example"))
			suite.fixture.must(suite.fixture.stack.SetText(collection, 1, "com.github", "scout"))
			suite.fixture.must(suite.fixture.stack.SetText(collection, 1, "com.github", ""))
			collectionAddress := strings.ToLower(collection.Hex())

			records, err := suite.sut.ListTextChanges(
				context.Background(),
				anvilChainId,
				fromBlock,
				suite.fixture.currentBlock(),
			)

			require.NoError(t, err)
			record := func(key, value string) *models.AgentEnsRecord {
				return &models.AgentEnsRecord{
					ChainId: anvilChainId, CollectionAddress: collectionAddress, TokenId: "1", Key: key, Value: value,
				}
			}
			assert.Equal(t, []*models.AgentEnsRecord{
				record("url", "https://scout.example"),
				record("com.github", "scout"),
				record("com.github", ""),
			}, records)
		})
	})
}
