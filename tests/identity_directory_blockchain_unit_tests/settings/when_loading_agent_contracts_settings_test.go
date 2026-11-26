package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWhenLoadingAgentContractsSettings(t *testing.T) {
	t.Parallel()

	logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

	t.Run("Given a chain 6022 is not deployed on", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			chainsSettings := &settings.ChainsSettings{1: {RpcHttpUrl: "http://node"}}

			assert.Panics(t, func() { settings.NewAgentContractsSettings(logger, chainsSettings) })
		})
	})

	t.Run("Given Polygon Amoy", func(t *testing.T) {
		t.Parallel()

		t.Run("Should use its baked manager and ENS registry", func(t *testing.T) {
			t.Parallel()

			chainsSettings := &settings.ChainsSettings{80002: {RpcHttpUrl: "http://node"}}

			agentContractsSettings := settings.NewAgentContractsSettings(logger, chainsSettings)

			assert.Equal(t, &settings.AgentContracts{
				AgentCollectionsManager: common.HexToAddress("0x8095A0D4bE42A61db9e609aA693f516044Bda990"),
				AgentEnsRegistry:        common.HexToAddress("0x37c5879529E5b0241663ed40feeebb2326BBEE0d"),
			}, agentContractsSettings.Get(80002))
		})
	})
}
