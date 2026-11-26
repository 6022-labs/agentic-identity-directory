package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const (
	windowChainId   = uint64(80002)
	windowFromBlock = uint64(100)
	windowToBlock   = uint64(199)
	windowHeadBlock = uint64(250)
)

type WhenReadingAChainWindowTestingSuite struct {
	mockAgentCollectionsManagerRequester *interfaces_mocks.MockAgentCollectionsManagerRequesterInterface
	mockAgentCollectionRequester         *interfaces_mocks.MockAgentCollectionRequesterInterface
	mockAgentEnsRegistryRequester        *interfaces_mocks.MockAgentEnsRegistryRequesterInterface
	mockAgentCollectionRepository        *repositories_mocks.MockAgentCollectionRepositoryInterface
	sut                                  *services.ChainWindowReader
}

func WhenReadingAChainWindowBeforeEach(t *testing.T) *WhenReadingAChainWindowTestingSuite {
	controller := gomock.NewController(t)
	suite := &WhenReadingAChainWindowTestingSuite{
		mockAgentCollectionsManagerRequester: interfaces_mocks.NewMockAgentCollectionsManagerRequesterInterface(
			controller,
		),
		mockAgentCollectionRequester:  interfaces_mocks.NewMockAgentCollectionRequesterInterface(controller),
		mockAgentEnsRegistryRequester: interfaces_mocks.NewMockAgentEnsRegistryRequesterInterface(controller),
		mockAgentCollectionRepository: repositories_mocks.NewMockAgentCollectionRepositoryInterface(controller),
	}
	suite.sut = services.NewChainWindowReader(
		zap.NewNop(),
		suite.mockAgentCollectionsManagerRequester,
		suite.mockAgentCollectionRequester,
		suite.mockAgentEnsRegistryRequester,
		suite.mockAgentCollectionRepository,
		services.NewEnsNameBuilder(),
	)
	return suite
}

func (s *WhenReadingAChainWindowTestingSuite) expectLogs(
	known, created []string,
	changes *models.CollectionChanges,
	ensRecords []*models.AgentEnsRecord,
) {
	s.mockAgentCollectionRepository.EXPECT().ListAddresses(gomock.Any(), windowChainId).Return(known, nil)
	s.mockAgentCollectionsManagerRequester.EXPECT().
		ListCreatedCollections(gomock.Any(), windowChainId, windowFromBlock, windowToBlock).
		Return(created, nil)
	s.mockAgentCollectionRequester.EXPECT().
		ListChanges(gomock.Any(), windowChainId, append(known, created...), windowFromBlock, windowToBlock).
		Return(changes, nil)
	s.mockAgentEnsRegistryRequester.EXPECT().
		ListTextChanges(gomock.Any(), windowChainId, windowFromBlock, windowToBlock).
		Return(ensRecords, nil)
}

func (s *WhenReadingAChainWindowTestingSuite) read() (*models.ChainWindow, error) {
	return s.sut.Read(context.Background(), windowChainId, windowFromBlock, windowToBlock, windowHeadBlock)
}

func TestWhenReadingAChainWindow(t *testing.T) {
	t.Parallel()

	knownCollection := model_builders.NewAgentCollectionBuilder().WithName("scouts").Build()

	t.Run("Given the known collections cannot be listed", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.mockAgentCollectionRepository.EXPECT().
				ListAddresses(gomock.Any(), windowChainId).
				Return(nil, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			assert.Nil(t, chainWindow)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the created collections cannot be listed", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("rpc down")

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.mockAgentCollectionRepository.EXPECT().
				ListAddresses(gomock.Any(), windowChainId).
				Return([]string{knownCollection.Address}, nil)
			suite.mockAgentCollectionsManagerRequester.EXPECT().
				ListCreatedCollections(gomock.Any(), windowChainId, windowFromBlock, windowToBlock).
				Return(nil, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			assert.Nil(t, chainWindow)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the collection changes cannot be listed", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("rpc down")

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.mockAgentCollectionRepository.EXPECT().
				ListAddresses(gomock.Any(), windowChainId).
				Return([]string{knownCollection.Address}, nil)
			suite.mockAgentCollectionsManagerRequester.EXPECT().
				ListCreatedCollections(gomock.Any(), windowChainId, windowFromBlock, windowToBlock).
				Return(nil, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				ListChanges(gomock.Any(), windowChainId, []string{knownCollection.Address}, windowFromBlock, windowToBlock).
				Return(nil, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			assert.Nil(t, chainWindow)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given the ENS text changes cannot be listed", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("rpc down")

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.mockAgentCollectionRepository.EXPECT().
				ListAddresses(gomock.Any(), windowChainId).
				Return([]string{knownCollection.Address}, nil)
			suite.mockAgentCollectionsManagerRequester.EXPECT().
				ListCreatedCollections(gomock.Any(), windowChainId, windowFromBlock, windowToBlock).
				Return(nil, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				ListChanges(gomock.Any(), windowChainId, []string{knownCollection.Address}, windowFromBlock, windowToBlock).
				Return(&models.CollectionChanges{}, nil)
			suite.mockAgentEnsRegistryRequester.EXPECT().
				ListTextChanges(gomock.Any(), windowChainId, windowFromBlock, windowToBlock).
				Return(nil, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			assert.Nil(t, chainWindow)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given a created collection whose contract cannot be read", func(t *testing.T) {
		t.Parallel()

		brokenAddress := common_tests.RandomLowerEthAddress()
		unreadable := fmt.Errorf("%w: execution reverted", apperrors.ErrContractUnreadable)

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.expectLogs(nil, []string{brokenAddress}, &models.CollectionChanges{}, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				GetCollection(gomock.Any(), windowChainId, brokenAddress, windowHeadBlock).
				Return(nil, unreadable)
			suite.mockAgentCollectionRequester.EXPECT().
				ListAgentRefs(gomock.Any(), windowChainId, brokenAddress, windowHeadBlock).
				Return(nil, unreadable)
		}

		t.Run("Should skip it", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			require.NoError(t, err)
			assert.Empty(t, chainWindow.Collections)
			assert.Empty(t, chainWindow.Agents)
		})
	})

	t.Run("Given a created collection the node fails to read", func(t *testing.T) {
		t.Parallel()

		createdAddress := common_tests.RandomLowerEthAddress()
		sourceErr := errors.New("connection reset")

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.expectLogs(nil, []string{createdAddress}, &models.CollectionChanges{}, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				GetCollection(gomock.Any(), windowChainId, createdAddress, windowHeadBlock).
				Return(nil, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			assert.Nil(t, chainWindow)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given an agent of a collection that is not indexed", func(t *testing.T) {
		t.Parallel()

		agentRef := models.AgentRef{CollectionAddress: common_tests.RandomLowerEthAddress(), TokenId: "1"}

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.expectLogs([]string{knownCollection.Address}, nil,
				&models.CollectionChanges{AgentRefs: []models.AgentRef{agentRef}}, nil)
			suite.mockAgentCollectionRepository.EXPECT().
				Get(gomock.Any(), windowChainId, agentRef.CollectionAddress).
				Return(nil, nil)
		}

		t.Run("Should skip it without reading it", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			require.NoError(t, err)
			assert.Empty(t, chainWindow.Agents)
		})
	})

	t.Run("Given an agent whose contract cannot be read", func(t *testing.T) {
		t.Parallel()

		agentRef := models.AgentRef{CollectionAddress: knownCollection.Address, TokenId: "9"}

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.expectLogs([]string{knownCollection.Address}, nil,
				&models.CollectionChanges{AgentRefs: []models.AgentRef{agentRef}}, nil)
			suite.mockAgentCollectionRepository.EXPECT().
				Get(gomock.Any(), windowChainId, knownCollection.Address).
				Return(knownCollection, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				GetAgent(gomock.Any(), windowChainId, agentRef, windowHeadBlock).
				Return(nil, apperrors.ErrContractUnreadable)
		}

		t.Run("Should skip it", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			require.NoError(t, err)
			assert.Empty(t, chainWindow.Agents)
		})
	})

	t.Run("Given an agent cannot be read because the node fails", func(t *testing.T) {
		t.Parallel()

		agentRef := models.AgentRef{CollectionAddress: knownCollection.Address, TokenId: "9"}
		sourceErr := errors.New("connection reset")

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.expectLogs([]string{knownCollection.Address}, nil,
				&models.CollectionChanges{AgentRefs: []models.AgentRef{agentRef}}, nil)
			suite.mockAgentCollectionRepository.EXPECT().
				Get(gomock.Any(), windowChainId, knownCollection.Address).
				Return(knownCollection, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				GetAgent(gomock.Any(), windowChainId, agentRef, windowHeadBlock).
				Return(nil, sourceErr)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			assert.Nil(t, chainWindow)
			assert.ErrorIs(t, err, sourceErr)
		})
	})

	t.Run("Given an agent of a known collection changed", func(t *testing.T) {
		t.Parallel()

		agentRef := models.AgentRef{CollectionAddress: knownCollection.Address, TokenId: "3"}
		agent := model_builders.NewAgentBuilder().
			WithCollectionAddress(knownCollection.Address).
			WithTokenId("3").
			WithName("scout\x00").
			WithEnsName("").
			Build()

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.expectLogs([]string{knownCollection.Address}, nil,
				&models.CollectionChanges{AgentRefs: []models.AgentRef{agentRef}}, nil)
			suite.mockAgentCollectionRepository.EXPECT().
				Get(gomock.Any(), windowChainId, knownCollection.Address).
				Return(knownCollection, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				GetAgent(gomock.Any(), windowChainId, agentRef, windowHeadBlock).
				Return(agent, nil)
		}

		t.Run("Should read it at the head block, sanitized and named", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			require.NoError(t, err)
			require.Len(t, chainWindow.Agents, 1)
			assert.Equal(t, "scout", chainWindow.Agents[0].Name)
			assert.Equal(t, "scout.scouts.80002.6022.eth", chainWindow.Agents[0].EnsName)
			assert.Equal(t, windowToBlock, chainWindow.ToBlock)
		})
	})

	t.Run("Given a collection created in the window that already holds agents", func(t *testing.T) {
		t.Parallel()

		createdCollection := model_builders.NewAgentCollectionBuilder().WithName("newcomers").Build()
		agentRefs := []models.AgentRef{
			{CollectionAddress: createdCollection.Address, TokenId: "1"},
			{CollectionAddress: createdCollection.Address, TokenId: "2"},
		}

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.expectLogs(nil, []string{createdCollection.Address},
				&models.CollectionChanges{AgentRefs: agentRefs[1:]}, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				GetCollection(gomock.Any(), windowChainId, createdCollection.Address, windowHeadBlock).
				Return(createdCollection, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				ListAgentRefs(gomock.Any(), windowChainId, createdCollection.Address, windowHeadBlock).
				Return(agentRefs, nil)
			for _, agentRef := range agentRefs {
				suite.mockAgentCollectionRequester.EXPECT().
					GetAgent(gomock.Any(), windowChainId, agentRef, windowHeadBlock).
					Return(model_builders.NewAgentBuilder().
						WithCollectionAddress(agentRef.CollectionAddress).
						WithTokenId(agentRef.TokenId).
						WithName(fmt.Sprintf("agent-%s", agentRef.TokenId)).
						Build(), nil)
			}
		}

		t.Run("Should read the collection and each of its agents once", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			require.NoError(t, err)
			assert.Equal(t, []*models.AgentCollection{createdCollection}, chainWindow.Collections)
			require.Len(t, chainWindow.Agents, 2)
			assert.Equal(t, "agent-2.newcomers.80002.6022.eth", chainWindow.Agents[0].EnsName)
			assert.Equal(t, "agent-1.newcomers.80002.6022.eth", chainWindow.Agents[1].EnsName)
		})
	})

	t.Run("Given ENS records set in the window", func(t *testing.T) {
		t.Parallel()

		knownRecord := &models.AgentEnsRecord{
			ChainId:           windowChainId,
			CollectionAddress: knownCollection.Address,
			TokenId:           "1",
			Key:               "url",
			Value:             "https://a\x00",
		}
		strangerRecord := &models.AgentEnsRecord{
			ChainId: windowChainId, CollectionAddress: common_tests.RandomLowerEthAddress(), TokenId: "1", Key: "url",
			Value: "https://b",
		}

		initSuite := func(suite *WhenReadingAChainWindowTestingSuite) {
			suite.expectLogs([]string{knownCollection.Address}, nil, &models.CollectionChanges{},
				[]*models.AgentEnsRecord{knownRecord, strangerRecord})
		}

		t.Run("Should keep the sanitized ones of indexed collections", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingAChainWindowBeforeEach(t)
			initSuite(suite)

			chainWindow, err := suite.read()

			require.NoError(t, err)
			require.Len(t, chainWindow.EnsRecords, 1)
			assert.Equal(t, knownCollection.Address, chainWindow.EnsRecords[0].CollectionAddress)
			assert.Equal(t, "https://a", chainWindow.EnsRecords[0].Value)
		})
	})
}
