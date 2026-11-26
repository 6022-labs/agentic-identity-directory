package use_cases_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/repositories_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenGettingAnAgentTestingSuite struct {
	mockAgentRepository *repositories_mocks.MockAgentRepositoryInterface
	sut                 *use_cases.GetAgent
}

func WhenGettingAnAgentBeforeEach(t *testing.T) *WhenGettingAnAgentTestingSuite {
	mockAgentRepository := repositories_mocks.NewMockAgentRepositoryInterface(gomock.NewController(t))
	return &WhenGettingAnAgentTestingSuite{
		mockAgentRepository: mockAgentRepository,
		sut:                 use_cases.NewGetAgent(mockAgentRepository),
	}
}

func TestWhenGettingAnAgent(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		request := &requests.GetAgentRequest{ChainId: 80002, CollectionAddress: "nope", TokenId: "1"}

		t.Run("Should fail with a validation error on collectionAddress", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentBeforeEach(t)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "collectionAddress", validationErr.Field)
		})
	})

	checksumAddress := common_tests.RandomEthAddress().Hex()
	agentRef := models.AgentRef{CollectionAddress: strings.ToLower(checksumAddress), TokenId: "7"}

	t.Run("Given the index cannot be read", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")
		request := &requests.GetAgentRequest{
			ChainId:           80002,
			CollectionAddress: agentRef.CollectionAddress,
			TokenId:           "7",
		}

		initSuite := func(suite *WhenGettingAnAgentTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any(), uint64(80002), agentRef).Return(nil, sourceErr)
		}

		t.Run("Should fail with get_agent_failed", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			require.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "get_agent_failed", internalErr.Code)
		})
	})

	t.Run("Given no agent has the token", func(t *testing.T) {
		t.Parallel()

		request := &requests.GetAgentRequest{
			ChainId:           80002,
			CollectionAddress: agentRef.CollectionAddress,
			TokenId:           "7",
		}

		initSuite := func(suite *WhenGettingAnAgentTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any(), uint64(80002), agentRef).Return(nil, nil)
		}

		t.Run("Should fail with agent_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var notFoundErr *apperrors.NotFoundError
			require.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "agent_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given an agent has the token", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().
			WithCollectionAddress(agentRef.CollectionAddress).
			WithTokenId("7").
			Build()
		request := &requests.GetAgentRequest{ChainId: 80002, CollectionAddress: checksumAddress, TokenId: "007"}

		initSuite := func(suite *WhenGettingAnAgentTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any(), uint64(80002), agentRef).Return(agent, nil)
		}

		t.Run("Should look it up by its lowercased address and plain token id", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			require.NoError(t, err)
			assert.Equal(t, "7", response.TokenId)
			assert.Equal(t, agent.CollectionAddress, response.CollectionAddress)
		})
	})
}
