package identity_directory_daemons_unit_tests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	identity_directory_daemons "github.com/6022-labs/agentic-identity-directory/src/identity_directory_daemons"
	metrics_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/metrics_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type WhenRunningTheChainSyncDaemonTestingSuite struct {
	mockBlockchainRequester *interfaces_mocks.MockBlockchainRequesterInterface
	logs                    *observer.ObservedLogs
	sut                     *identity_directory_daemons.ChainSyncDaemon
}

func WhenRunningTheChainSyncDaemonBeforeEach(t *testing.T) *WhenRunningTheChainSyncDaemonTestingSuite {
	controller := gomock.NewController(t)
	mockBlockchainRequester := interfaces_mocks.NewMockBlockchainRequesterInterface(controller)
	mockSyncMetrics := metrics_mocks.NewMockSyncMetricsInterface(controller)
	mockSyncMetrics.EXPECT().RecordPass(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	syncSettings := &settings.SyncSettings{80002: {PollInterval: 10 * time.Millisecond, LogWindowBlocks: 10}}
	syncChain := use_cases.NewSyncChain(
		syncSettings,
		mockBlockchainRequester,
		repositories_mocks.NewMockSyncCursorRepositoryInterface(controller),
		interfaces_mocks.NewMockChainWindowReaderInterface(controller),
		interfaces_mocks.NewMockChainWindowWriterInterface(controller),
		mockSyncMetrics,
	)
	core, logs := observer.New(zapcore.DebugLevel)
	return &WhenRunningTheChainSyncDaemonTestingSuite{
		mockBlockchainRequester: mockBlockchainRequester,
		logs:                    logs,
		sut:                     identity_directory_daemons.NewChainSyncDaemon(zap.New(core), syncSettings, syncChain),
	}
}

func TestWhenRunningTheChainSyncDaemon(t *testing.T) {
	t.Parallel()

	t.Run("Given every pass fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRunningTheChainSyncDaemonTestingSuite) {
			suite.mockBlockchainRequester.EXPECT().
				GetBlockNumber(gomock.Any(), uint64(80002)).
				Return(uint64(0), errors.New("node down")).
				MinTimes(2)
		}

		t.Run("Should log each failure and retry at the next tick until stopped", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningTheChainSyncDaemonBeforeEach(t)
			initSuite(suite)

			suite.sut.Start(context.Background())
			assert.Eventually(t, func() bool {
				return suite.logs.FilterMessage("chain sync pass failed").Len() >= 2
			}, 2*time.Second, 5*time.Millisecond)
			suite.sut.Stop()
		})
	})
}
