package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/services"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type WhenListingEnsTextChangesTestingSuite struct {
	stub *rpcStub
	logs *observer.ObservedLogs
	sut  *services.AgentEnsRegistryRequester
}

func WhenListingEnsTextChangesBeforeEach(t *testing.T, rateLimited int32) *WhenListingEnsTextChangesTestingSuite {
	stub := newRateLimitedRpcStub(t, rateLimited, []map[string]any{})
	core, logs := observer.New(zapcore.WarnLevel)
	agentContractsSettings := &settings.AgentContractsSettings{stubChainId: {
		AgentCollectionsManager: common_tests.RandomEthAddress(),
		AgentEnsRegistry:        common_tests.RandomEthAddress(),
	}}
	sut, err := services.NewAgentEnsRegistryRequester(
		zap.New(core), stub.ethClientFactory(t), stub.chainsSettings(), agentContractsSettings)
	require.NoError(t, err)
	return &WhenListingEnsTextChangesTestingSuite{stub: stub, logs: logs, sut: sut}
}

func TestWhenListingEnsTextChanges(t *testing.T) {
	t.Parallel()

	t.Run("Given the node rate limits the first call", func(t *testing.T) {
		t.Parallel()

		t.Run("Should retry, log the retry and return the records", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingEnsTextChangesBeforeEach(t, 1)

			records, err := suite.sut.ListTextChanges(context.Background(), stubChainId, 1, 10)

			require.NoError(t, err)
			assert.Empty(t, records)
			assert.Equal(t, int32(2), suite.stub.calls.Load())
			assert.Equal(t, 1, suite.logs.FilterMessage("rpc call failed, retrying").Len())
		})
	})
}
