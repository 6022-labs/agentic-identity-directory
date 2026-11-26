package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/settings"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	metrics_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/metrics_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const syncedChainId = uint64(80002)

type WhenSyncingAChainTestingSuite struct {
	mockBlockchainRequester  *interfaces_mocks.MockBlockchainRequesterInterface
	mockSyncCursorRepository *repositories_mocks.MockSyncCursorRepositoryInterface
	mockChainWindowReader    *interfaces_mocks.MockChainWindowReaderInterface
	mockChainWindowWriter    *interfaces_mocks.MockChainWindowWriterInterface
	mockSyncMetrics          *metrics_mocks.MockSyncMetricsInterface
	sut                      *use_cases.SyncChain
}

func WhenSyncingAChainBeforeEach(t *testing.T) *WhenSyncingAChainTestingSuite {
	controller := gomock.NewController(t)
	suite := &WhenSyncingAChainTestingSuite{
		mockBlockchainRequester:  interfaces_mocks.NewMockBlockchainRequesterInterface(controller),
		mockSyncCursorRepository: repositories_mocks.NewMockSyncCursorRepositoryInterface(controller),
		mockChainWindowReader:    interfaces_mocks.NewMockChainWindowReaderInterface(controller),
		mockChainWindowWriter:    interfaces_mocks.NewMockChainWindowWriterInterface(controller),
		mockSyncMetrics:          metrics_mocks.NewMockSyncMetricsInterface(controller),
	}
	suite.mockSyncMetrics.EXPECT().RecordPass(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	suite.mockSyncMetrics.EXPECT().RecordPosition(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	syncSettings := &settings.SyncSettings{syncedChainId: {
		StartBlock:      100,
		Confirmations:   10,
		PollInterval:    time.Second,
		LogWindowBlocks: 50,
	}}
	suite.sut = use_cases.NewSyncChain(
		syncSettings,
		suite.mockBlockchainRequester,
		suite.mockSyncCursorRepository,
		suite.mockChainWindowReader,
		suite.mockChainWindowWriter,
		suite.mockSyncMetrics,
	)
	return suite
}

func (s *WhenSyncingAChainTestingSuite) window(fromBlock, toBlock uint64) *models.ChainWindow {
	return &models.ChainWindow{ChainId: syncedChainId, FromBlock: fromBlock, ToBlock: toBlock}
}

func TestWhenSyncingAChain(t *testing.T) {
	t.Parallel()

	t.Run("Given a chain that is not configured", func(t *testing.T) {
		t.Parallel()

		request := &requests.SyncChainRequest{ChainId: 137}

		t.Run("Should fail with a validation error on chainId", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)

			err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "chainId", validationErr.Field)
		})
	})

	request := &requests.SyncChainRequest{ChainId: syncedChainId}
	cursor := uint64(129)

	t.Run("Given the head block cannot be read", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("node down")

		initSuite := func(suite *WhenSyncingAChainTestingSuite) {
			suite.mockBlockchainRequester.EXPECT().
				GetBlockNumber(gomock.Any(), syncedChainId).
				Return(uint64(0), sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given a chain shorter than its confirmations", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSyncingAChainTestingSuite) {
			suite.mockBlockchainRequester.EXPECT().GetBlockNumber(gomock.Any(), syncedChainId).Return(uint64(5), nil)
		}

		t.Run("Should read no window", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the cursor cannot be read", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")

		initSuite := func(suite *WhenSyncingAChainTestingSuite) {
			suite.mockBlockchainRequester.EXPECT().GetBlockNumber(gomock.Any(), syncedChainId).Return(uint64(200), nil)
			suite.mockSyncCursorRepository.EXPECT().Get(gomock.Any(), syncedChainId).Return(nil, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given a chain indexed up to its confirmed head", func(t *testing.T) {
		t.Parallel()

		head := uint64(140)

		initSuite := func(suite *WhenSyncingAChainTestingSuite) {
			suite.mockBlockchainRequester.EXPECT().GetBlockNumber(gomock.Any(), syncedChainId).Return(uint64(150), nil)
			suite.mockSyncCursorRepository.EXPECT().Get(gomock.Any(), syncedChainId).Return(&head, nil)
		}

		t.Run("Should read no window", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given a window cannot be read", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("rpc refused")

		initSuite := func(suite *WhenSyncingAChainTestingSuite) {
			suite.mockBlockchainRequester.EXPECT().GetBlockNumber(gomock.Any(), syncedChainId).Return(uint64(150), nil)
			suite.mockSyncCursorRepository.EXPECT().Get(gomock.Any(), syncedChainId).Return(&cursor, nil)
			suite.mockChainWindowReader.EXPECT().
				Read(gomock.Any(), syncedChainId, uint64(130), uint64(140), uint64(140)).
				Return(nil, sourceErr)
		}

		t.Run("Should return the error without moving the cursor", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given another replica indexed the window first", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSyncingAChainTestingSuite) {
			window := suite.window(130, 140)
			suite.mockBlockchainRequester.EXPECT().GetBlockNumber(gomock.Any(), syncedChainId).Return(uint64(150), nil)
			suite.mockSyncCursorRepository.EXPECT().Get(gomock.Any(), syncedChainId).Return(&cursor, nil)
			suite.mockChainWindowReader.EXPECT().
				Read(gomock.Any(), syncedChainId, uint64(130), uint64(140), uint64(140)).
				Return(window, nil)
			suite.mockChainWindowWriter.EXPECT().
				Write(gomock.Any(), window, &cursor).
				Return(apperrors.ErrSyncCursorMoved)
		}

		t.Run("Should end the pass without an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given a window cannot be written", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")

		initSuite := func(suite *WhenSyncingAChainTestingSuite) {
			window := suite.window(130, 140)
			suite.mockBlockchainRequester.EXPECT().GetBlockNumber(gomock.Any(), syncedChainId).Return(uint64(150), nil)
			suite.mockSyncCursorRepository.EXPECT().Get(gomock.Any(), syncedChainId).Return(&cursor, nil)
			suite.mockChainWindowReader.EXPECT().
				Read(gomock.Any(), syncedChainId, uint64(130), uint64(140), uint64(140)).
				Return(window, nil)
			suite.mockChainWindowWriter.EXPECT().Write(gomock.Any(), window, &cursor).Return(sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), request)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given a chain never indexed", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSyncingAChainTestingSuite) {
			suite.mockBlockchainRequester.EXPECT().
				GetBlockNumber(gomock.Any(), syncedChainId).
				Return(uint64(140), nil).
				Times(2)
			gomock.InOrder(
				suite.mockSyncCursorRepository.EXPECT().Get(gomock.Any(), syncedChainId).Return(nil, nil),
				suite.mockSyncCursorRepository.EXPECT().Get(gomock.Any(), syncedChainId).Return(new(uint64(130)), nil),
			)
		}

		t.Run("Should read from the start block up to the confirmed head", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			initSuite(suite)
			window := suite.window(100, 130)
			suite.mockChainWindowReader.EXPECT().
				Read(gomock.Any(), syncedChainId, uint64(100), uint64(130), uint64(130)).
				Return(window, nil)
			suite.mockChainWindowWriter.EXPECT().Write(gomock.Any(), window, nil).Return(nil)

			err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given a chain more than a window behind", func(t *testing.T) {
		t.Parallel()

		firstCursor := uint64(99)
		secondCursor := uint64(149)

		initSuite := func(suite *WhenSyncingAChainTestingSuite) {
			suite.mockBlockchainRequester.EXPECT().
				GetBlockNumber(gomock.Any(), syncedChainId).
				Return(uint64(190), nil).
				Times(3)
			gomock.InOrder(
				suite.mockSyncCursorRepository.EXPECT().Get(gomock.Any(), syncedChainId).Return(&firstCursor, nil),
				suite.mockSyncCursorRepository.EXPECT().Get(gomock.Any(), syncedChainId).Return(&secondCursor, nil),
				suite.mockSyncCursorRepository.EXPECT().Get(gomock.Any(), syncedChainId).Return(new(uint64(180)), nil),
			)
		}

		t.Run("Should write one window after the other up to the confirmed head", func(t *testing.T) {
			t.Parallel()

			suite := WhenSyncingAChainBeforeEach(t)
			initSuite(suite)
			firstWindow := suite.window(100, 149)
			secondWindow := suite.window(150, 180)
			gomock.InOrder(
				suite.mockChainWindowReader.EXPECT().
					Read(gomock.Any(), syncedChainId, uint64(100), uint64(149), uint64(180)).
					Return(firstWindow, nil),
				suite.mockChainWindowWriter.EXPECT().Write(gomock.Any(), firstWindow, &firstCursor).Return(nil),
				suite.mockChainWindowReader.EXPECT().
					Read(gomock.Any(), syncedChainId, uint64(150), uint64(180), uint64(180)).
					Return(secondWindow, nil),
				suite.mockChainWindowWriter.EXPECT().Write(gomock.Any(), secondWindow, &secondCursor).Return(nil),
			)

			err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})
}
