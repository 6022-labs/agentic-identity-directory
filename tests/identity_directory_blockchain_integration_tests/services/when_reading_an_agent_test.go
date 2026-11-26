package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenReadingAnAgentTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenReadingAnAgentBeforeEach(t *testing.T) *WhenReadingAnAgentTestingSuite {
	fixture := newChainFixture(t)
	return &WhenReadingAnAgentTestingSuite{fixture: fixture, sut: fixture.agentCollectionRequester()}
}

func TestWhenReadingAnAgent(t *testing.T) {
	t.Parallel()

	t.Run("Given a token that was never minted", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail with ErrContractUnreadable", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnAgentBeforeEach(t)
			collection, err := suite.fixture.stack.CreateCollection("scouts")
			suite.fixture.must(err)
			agentRef := models.AgentRef{CollectionAddress: strings.ToLower(collection.Hex()), TokenId: "1"}

			agent, err := suite.sut.GetAgent(context.Background(), anvilChainId, agentRef, suite.fixture.currentBlock())

			assert.Nil(t, agent)
			assert.ErrorIs(t, err, apperrors.ErrContractUnreadable)
		})
	})

	t.Run("Given an address without a contract", func(t *testing.T) {
		t.Parallel()

		t.Run(
			"Should fail without blaming the contract, as a node behind the chain answers the same",
			func(t *testing.T) {
				t.Parallel()

				suite := WhenReadingAnAgentBeforeEach(t)
				agentRef := models.AgentRef{CollectionAddress: common_tests.RandomLowerEthAddress(), TokenId: "1"}

				agent, err := suite.sut.GetAgent(
					context.Background(),
					anvilChainId,
					agentRef,
					suite.fixture.currentBlock(),
				)

				assert.Nil(t, agent)
				assert.ErrorContains(t, err, "no contract code")
				assert.NotErrorIs(t, err, apperrors.ErrContractUnreadable)
			},
		)
	})

	t.Run("Given a minted and transferred agent", func(t *testing.T) {
		t.Parallel()

		t.Run("Should read its information and current owner", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAnAgentBeforeEach(t)
			collection, err := suite.fixture.stack.CreateCollection("scouts")
			suite.fixture.must(err)
			evmAddress := common_tests.RandomEthAddress()
			newOwner := common_tests.RandomEthAddress()
			suite.fixture.must(suite.fixture.stack.MintAgent(collection, "scout", evmAddress))
			suite.fixture.must(suite.fixture.stack.Transfer(collection, 1, newOwner))
			collectionAddress := strings.ToLower(collection.Hex())

			agent, err := suite.sut.GetAgent(context.Background(), anvilChainId,
				models.AgentRef{CollectionAddress: collectionAddress, TokenId: "1"}, suite.fixture.currentBlock())

			require.NoError(t, err)
			assert.Equal(t, "scout", agent.Name)
			assert.Equal(t, strings.ToLower(newOwner.Hex()), agent.Owner)
			assert.Equal(t, map[string]string{"role": "expert", "mission": "integration testing"}, agent.Attributes)
			assert.Contains(t, agent.Images, "default")
			assert.Nil(t, agent.CloneOf)
			assert.Equal(t, []models.AgentAddress{{
				ChainId: anvilChainId, CollectionAddress: collectionAddress, TokenId: "1",
				Type: "evm", Value: strings.ToLower(evmAddress.Hex()),
			}}, agent.Addresses)
		})
	})
}
