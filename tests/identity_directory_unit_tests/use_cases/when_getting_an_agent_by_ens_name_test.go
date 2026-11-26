package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/repositories_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenGettingAnAgentByEnsNameTestingSuite struct {
	mockAgentRepository *repositories_mocks.MockAgentRepositoryInterface
	sut                 *use_cases.GetAgentByEnsName
}

func WhenGettingAnAgentByEnsNameBeforeEach(t *testing.T) *WhenGettingAnAgentByEnsNameTestingSuite {
	mockAgentRepository := repositories_mocks.NewMockAgentRepositoryInterface(gomock.NewController(t))
	return &WhenGettingAnAgentByEnsNameTestingSuite{
		mockAgentRepository: mockAgentRepository,
		sut:                 use_cases.NewGetAgentByEnsName(mockAgentRepository),
	}
}

func TestWhenGettingAnAgentByEnsName(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		request := &requests.EnsNameRequest{EnsName: " "}

		t.Run("Should fail with a validation error on ensName", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentByEnsNameBeforeEach(t)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "ensName", validationErr.Field)
		})
	})

	t.Run("Given the index cannot be read", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")
		request := &requests.EnsNameRequest{EnsName: "scout.agents.80002.6022.eth"}

		initSuite := func(suite *WhenGettingAnAgentByEnsNameTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(nil, sourceErr)
		}

		t.Run("Should fail with get_agent_failed", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentByEnsNameBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			require.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "get_agent_failed", internalErr.Code)
		})
	})

	t.Run("Given no agent has the ENS name", func(t *testing.T) {
		t.Parallel()

		request := &requests.EnsNameRequest{EnsName: "ghost.agents.80002.6022.eth"}

		initSuite := func(suite *WhenGettingAnAgentByEnsNameTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), request.EnsName).Return(nil, nil)
		}

		t.Run("Should fail with agent_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentByEnsNameBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var notFoundErr *apperrors.NotFoundError
			require.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "agent_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given an agent has the ENS name", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenGettingAnAgentByEnsNameTestingSuite) {
			suite.mockAgentRepository.EXPECT().GetByEnsName(gomock.Any(), agent.EnsName).Return(agent, nil)
		}

		t.Run("Should look it up lowercased and return it", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAnAgentByEnsNameBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(
				context.Background(),
				&requests.EnsNameRequest{EnsName: "Scout.Agents.80002.6022.ETH"},
			)

			require.NoError(t, err)
			assert.Equal(t, agent.EnsName, response.EnsName)
		})
	})
}
