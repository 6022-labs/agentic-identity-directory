package services_test

import (
	"context"
	"strings"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type WhenReadingAnAgentTestingSuite struct {
	stub *rpcStub
	sut  *services.AgentCollectionRequester
}

func WhenReadingAnAgentBeforeEach(t *testing.T, stub *rpcStub) *WhenReadingAnAgentTestingSuite {
	sut, err := services.NewAgentCollectionRequester(zap.NewNop(), stub.ethClientFactory(t), stub.chainsSettings())
	require.NoError(t, err)
	return &WhenReadingAnAgentTestingSuite{stub: stub, sut: sut}
}

func TestWhenReadingAnAgent(t *testing.T) {
	t.Parallel()

	agentRef := models.AgentRef{CollectionAddress: strings.ToLower(common_tests.RandomEthAddress().Hex()), TokenId: "1"}

	t.Run("Given an agent too heavy for the node to read", func(t *testing.T) {
		t.Parallel()

		outOfGas := func(int32, string) rpcAnswer {
			return rpcAnswer{err: map[string]any{"code": -32000, "message": "out of gas"}}
		}

		t.Run("Should fail with ErrContractUnreadable without retrying", func(t *testing.T) {
			t.Parallel()
			suite := WhenReadingAnAgentBeforeEach(t, newRpcStub(t, outOfGas))

			agent, err := suite.sut.GetAgent(context.Background(), stubChainId, agentRef, 10)

			assert.Nil(t, agent)
			assert.ErrorIs(t, err, apperrors.ErrContractUnreadable)
			assert.Equal(t, int32(1), suite.stub.calls.Load())
		})
	})

	t.Run("Given a provider that refuses the call for its own reasons", func(t *testing.T) {
		t.Parallel()

		quotaExhausted := func(int32, string) rpcAnswer {
			return rpcAnswer{err: map[string]any{"code": -32603, "message": "monthly capacity limit exceeded"}}
		}

		t.Run("Should fail without blaming the contract", func(t *testing.T) {
			t.Parallel()
			suite := WhenReadingAnAgentBeforeEach(t, newRpcStub(t, quotaExhausted))

			agent, err := suite.sut.GetAgent(context.Background(), stubChainId, agentRef, 10)

			assert.Nil(t, agent)
			assert.ErrorContains(t, err, "monthly capacity limit exceeded")
			assert.NotErrorIs(t, err, apperrors.ErrContractUnreadable)
		})
	})

	t.Run("Given a node that lags behind the requested block", func(t *testing.T) {
		t.Parallel()

		lagging := func(int32, string) rpcAnswer {
			return rpcAnswer{err: map[string]any{"code": -32000, "message": "header not found"}}
		}

		t.Run("Should not mistake it for an unreadable contract", func(t *testing.T) {
			t.Parallel()
			suite := WhenReadingAnAgentBeforeEach(t, newRpcStub(t, lagging))
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := suite.sut.GetAgent(ctx, stubChainId, agentRef, 10)

			assert.Error(t, err)
			assert.NotErrorIs(t, err, apperrors.ErrContractUnreadable)
		})
	})
}
