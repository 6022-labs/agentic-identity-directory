package settings_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type WhenLoadingSyncSettingsTestingSuite struct {
	logger *zap.Logger
}

func WhenLoadingSyncSettingsBeforeEach() *WhenLoadingSyncSettingsTestingSuite {
	return &WhenLoadingSyncSettingsTestingSuite{
		logger: zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic)),
	}
}

func (s *WhenLoadingSyncSettingsTestingSuite) koanf(t *testing.T, values map[string]any) *koanf.Koanf {
	k := koanf.New(".")
	require.NoError(t, k.Load(confmap.Provider(values, "."), nil))
	return k
}

func TestWhenLoadingSyncSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given no chain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingSyncSettingsBeforeEach()

			assert.Panics(t, func() { settings.NewSyncSettings(suite.logger, suite.koanf(t, map[string]any{})) })
		})
	})

	t.Run("Given a chain 6022 is not deployed on", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingSyncSettingsBeforeEach()
			k := suite.koanf(t, map[string]any{"chains.1.testnet": false})

			assert.Panics(t, func() { settings.NewSyncSettings(suite.logger, k) })
		})
	})

	t.Run("Given a chain without its testnet flag", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingSyncSettingsBeforeEach()
			k := suite.koanf(t, map[string]any{"chains.80002.rpc_http_url": "http://node"})

			assert.Panics(t, func() { settings.NewSyncSettings(suite.logger, k) })
		})
	})

	t.Run("Given a testnet and a mainnet chain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fatal", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingSyncSettingsBeforeEach()
			k := suite.koanf(t, map[string]any{"chains.80002.testnet": true, "chains.137.testnet": false})

			assert.Panics(t, func() { settings.NewSyncSettings(suite.logger, k) })
		})
	})

	t.Run("Given a chain with only its testnet flag", func(t *testing.T) {
		t.Parallel()

		t.Run("Should start at the baked deployment block with the defaults", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingSyncSettingsBeforeEach()
			k := suite.koanf(t, map[string]any{"chains.80002.testnet": "true"})

			syncSettings := settings.NewSyncSettings(suite.logger, k)

			assert.Equal(t, &settings.ChainSyncSettings{
				StartBlock:      44057511,
				Confirmations:   20,
				PollInterval:    15 * time.Second,
				LogWindowBlocks: 10_000,
			}, syncSettings.Get(80002))
		})
	})

	t.Run("Given a chain with every setting", func(t *testing.T) {
		t.Parallel()

		t.Run("Should use them, zero confirmations included", func(t *testing.T) {
			t.Parallel()

			suite := WhenLoadingSyncSettingsBeforeEach()
			k := suite.koanf(t, map[string]any{
				"chains.31337.testnet":               true,
				"chains.31337.confirmations":         "0",
				"chains.31337.poll_interval_seconds": "2",
				"chains.31337.log_window_blocks":     "500",
			})

			syncSettings := settings.NewSyncSettings(suite.logger, k)

			assert.Equal(t, &settings.ChainSyncSettings{
				StartBlock:      1,
				Confirmations:   0,
				PollInterval:    2 * time.Second,
				LogWindowBlocks: 500,
			}, syncSettings.Get(31337))
			assert.Equal(t, []uint64{31337}, syncSettings.ChainIds())
		})
	})
}
