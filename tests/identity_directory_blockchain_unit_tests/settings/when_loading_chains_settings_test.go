package settings_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_blockchain/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type WhenLoadingChainsSettingsTestingSuite struct {
	logger *zap.Logger
}

func WhenLoadingChainsSettingsBeforeEach() *WhenLoadingChainsSettingsTestingSuite {
	return &WhenLoadingChainsSettingsTestingSuite{
		logger: zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic)),
	}
}

func (s *WhenLoadingChainsSettingsTestingSuite) koanf(t *testing.T, values map[string]any) *koanf.Koanf {
	k := koanf.New(".")
	require.NoError(t, k.Load(confmap.Provider(values, "."), nil))
	return k
}

func TestWhenLoadingChainsSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a chain without an rpc url", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()
			k := suite.koanf(t, map[string]any{"chains.80002.testnet": true})

			assert.Panics(t, func() { settings.NewChainsSettings(suite.logger, k) })
		})
	})

	t.Run("Given a chain with only its rpc url", func(t *testing.T) {
		t.Parallel()

		t.Run("Should apply the defaults", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()
			k := suite.koanf(t, map[string]any{"chains.80002.rpc_http_url": " http://node "})

			chainsSettings := settings.NewChainsSettings(suite.logger, k)

			assert.Equal(t, &settings.ChainConfig{
				RpcHttpUrl:              "http://node",
				RequestTimeout:          10 * time.Second,
				MaxAddressesPerLogQuery: 10,
			}, chainsSettings.Get(80002))
		})
	})

	t.Run("Given a chain with every setting", func(t *testing.T) {
		t.Parallel()

		t.Run("Should use them", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingChainsSettingsBeforeEach()
			k := suite.koanf(t, map[string]any{
				"chains.80002.rpc_http_url":                "http://node",
				"chains.80002.request_timeout_seconds":     "3",
				"chains.80002.max_addresses_per_log_query": "100",
			})

			chainsSettings := settings.NewChainsSettings(suite.logger, k)

			assert.Equal(t, &settings.ChainConfig{
				RpcHttpUrl:              "http://node",
				RequestTimeout:          3 * time.Second,
				MaxAddressesPerLogQuery: 100,
			}, chainsSettings.Get(80002))
		})
	})
}
