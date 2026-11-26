package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenReadingTheBlockNumberTestingSuite struct {
	fixture *chainFixture
	sut     *services.BlockchainRequester
}

func WhenReadingTheBlockNumberBeforeEach(t *testing.T) *WhenReadingTheBlockNumberTestingSuite {
	fixture := newChainFixture(t)
	return &WhenReadingTheBlockNumberTestingSuite{fixture: fixture, sut: fixture.blockchainRequester()}
}

func TestWhenReadingTheBlockNumber(t *testing.T) {
	t.Parallel()

	t.Run("Given a running chain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its head", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingTheBlockNumberBeforeEach(t)

			blockNumber, err := suite.sut.GetBlockNumber(context.Background(), anvilChainId)

			require.NoError(t, err)
			assert.GreaterOrEqual(t, blockNumber, suite.fixture.currentBlock()-1)
		})
	})
}
