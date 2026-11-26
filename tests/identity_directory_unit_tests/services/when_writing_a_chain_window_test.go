package services_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenWritingAChainWindowTestingSuite struct {
	mockUnitOfWorkFactory         *interfaces_mocks.MockUnitOfWorkFactoryInterface
	mockUnitOfWork                *interfaces_mocks.MockUnitOfWorkInterface
	mockAgentCollectionRepository *repositories_mocks.MockAgentCollectionRepositoryInterface
	mockAgentRepository           *repositories_mocks.MockAgentRepositoryInterface
	mockAgentEnsRecordRepository  *repositories_mocks.MockAgentEnsRecordRepositoryInterface
	mockSyncCursorRepository      *repositories_mocks.MockSyncCursorRepositoryInterface
	sut                           *services.ChainWindowWriter
}

func WhenWritingAChainWindowBeforeEach(t *testing.T) *WhenWritingAChainWindowTestingSuite {
	controller := gomock.NewController(t)
	suite := &WhenWritingAChainWindowTestingSuite{
		mockUnitOfWorkFactory:         interfaces_mocks.NewMockUnitOfWorkFactoryInterface(controller),
		mockUnitOfWork:                interfaces_mocks.NewMockUnitOfWorkInterface(controller),
		mockAgentCollectionRepository: repositories_mocks.NewMockAgentCollectionRepositoryInterface(controller),
		mockAgentRepository:           repositories_mocks.NewMockAgentRepositoryInterface(controller),
		mockAgentEnsRecordRepository:  repositories_mocks.NewMockAgentEnsRecordRepositoryInterface(controller),
		mockSyncCursorRepository:      repositories_mocks.NewMockSyncCursorRepositoryInterface(controller),
	}
	suite.sut = services.NewChainWindowWriter(
		suite.mockUnitOfWorkFactory,
		suite.mockAgentCollectionRepository,
		suite.mockAgentRepository,
		suite.mockAgentEnsRecordRepository,
		suite.mockSyncCursorRepository,
	)
	return suite
}

func TestWhenWritingAChainWindow(t *testing.T) {
	t.Parallel()

	collection := model_builders.NewAgentCollectionBuilder().Build()
	agent := model_builders.NewAgentBuilder().WithCollectionAddress(collection.Address).Build()
	firstRecord := &models.AgentEnsRecord{
		ChainId: 80002, CollectionAddress: collection.Address, TokenId: "1", Key: "url", Value: "https://a",
	}
	secondRecord := &models.AgentEnsRecord{
		ChainId: 80002, CollectionAddress: collection.Address, TokenId: "1", Key: "url", Value: "https://b",
	}
	chainWindow := &models.ChainWindow{
		ChainId:     80002,
		FromBlock:   100,
		ToBlock:     199,
		Collections: []*models.AgentCollection{collection},
		Agents:      []*models.Agent{agent},
		EnsRecords:  []*models.AgentEnsRecord{firstRecord, secondRecord},
	}
	expectedCursor := uint64(99)

	t.Run("Given the transaction cannot be started", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")

		initSuite := func(suite *WhenWritingAChainWindowTestingSuite) {
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(nil, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingAChainWindowBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Write(context.Background(), chainWindow, &expectedCursor)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given a collection cannot be stored", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")

		initSuite := func(suite *WhenWritingAChainWindowTestingSuite) {
			common_tests.ExpectUnitOfWork(suite.mockUnitOfWorkFactory, suite.mockUnitOfWork)
			suite.mockAgentCollectionRepository.EXPECT().Upsert(common_tests.InScope(), collection).Return(sourceErr)
		}

		t.Run("Should return the error without committing", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingAChainWindowBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Write(context.Background(), chainWindow, &expectedCursor)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given an agent cannot be stored", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")

		initSuite := func(suite *WhenWritingAChainWindowTestingSuite) {
			common_tests.ExpectUnitOfWork(suite.mockUnitOfWorkFactory, suite.mockUnitOfWork)
			suite.mockAgentCollectionRepository.EXPECT().Upsert(common_tests.InScope(), collection).Return(nil)
			suite.mockAgentRepository.EXPECT().Upsert(common_tests.InScope(), agent).Return(sourceErr)
		}

		t.Run("Should return the error without moving the cursor", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingAChainWindowBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Write(context.Background(), chainWindow, &expectedCursor)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given an ENS record cannot be stored", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")

		initSuite := func(suite *WhenWritingAChainWindowTestingSuite) {
			common_tests.ExpectUnitOfWork(suite.mockUnitOfWorkFactory, suite.mockUnitOfWork)
			suite.mockAgentCollectionRepository.EXPECT().Upsert(common_tests.InScope(), collection).Return(nil)
			suite.mockAgentRepository.EXPECT().Upsert(common_tests.InScope(), agent).Return(nil)
			suite.mockAgentEnsRecordRepository.EXPECT().Set(common_tests.InScope(), firstRecord).Return(sourceErr)
		}

		t.Run("Should return the error without committing", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingAChainWindowBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Write(context.Background(), chainWindow, &expectedCursor)

			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given another replica moved the cursor", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenWritingAChainWindowTestingSuite) {
			common_tests.ExpectUnitOfWork(suite.mockUnitOfWorkFactory, suite.mockUnitOfWork)
			suite.mockAgentCollectionRepository.EXPECT().Upsert(common_tests.InScope(), collection).Return(nil)
			suite.mockAgentRepository.EXPECT().Upsert(common_tests.InScope(), agent).Return(nil)
			suite.mockAgentEnsRecordRepository.EXPECT().Set(common_tests.InScope(), gomock.Any()).Return(nil).Times(2)
			suite.mockSyncCursorRepository.EXPECT().
				Advance(common_tests.InScope(), uint64(80002), &expectedCursor, uint64(199)).
				Return(apperrors.ErrSyncCursorMoved)
		}

		t.Run("Should return ErrSyncCursorMoved without committing", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingAChainWindowBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Write(context.Background(), chainWindow, &expectedCursor)

			assert.ErrorIs(t, err, apperrors.ErrSyncCursorMoved)
		})
	})

	t.Run("Given every write succeeds", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenWritingAChainWindowTestingSuite) {
			common_tests.ExpectUnitOfWork(suite.mockUnitOfWorkFactory, suite.mockUnitOfWork)
			suite.mockAgentCollectionRepository.EXPECT().Upsert(common_tests.InScope(), collection).Return(nil)
			suite.mockAgentRepository.EXPECT().Upsert(common_tests.InScope(), agent).Return(nil)
			suite.mockSyncCursorRepository.EXPECT().
				Advance(common_tests.InScope(), uint64(80002), &expectedCursor, uint64(199)).
				Return(nil)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil)
		}

		t.Run("Should set the ENS records in log order and commit", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingAChainWindowBeforeEach(t)
			initSuite(suite)
			gomock.InOrder(
				suite.mockAgentEnsRecordRepository.EXPECT().Set(common_tests.InScope(), firstRecord).Return(nil),
				suite.mockAgentEnsRecordRepository.EXPECT().Set(common_tests.InScope(), secondRecord).Return(nil),
			)

			err := suite.sut.Write(context.Background(), chainWindow, &expectedCursor)

			assert.NoError(t, err)
		})
	})
}
