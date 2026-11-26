package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-identity-directory/tests/identity_directory_mocks/repositories_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type WhenSearchingAgentsTestingSuite struct {
	mockAgentRepository *repositories_mocks.MockAgentRepositoryInterface
	sut                 *use_cases.SearchAgents
}

func WhenSearchingAgentsBeforeEach(t *testing.T) *WhenSearchingAgentsTestingSuite {
	mockAgentRepository := repositories_mocks.NewMockAgentRepositoryInterface(gomock.NewController(t))
	return &WhenSearchingAgentsTestingSuite{
		mockAgentRepository: mockAgentRepository,
		sut:                 use_cases.NewSearchAgents(mockAgentRepository),
	}
}

func TestWhenSearchingAgents(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		request := &requests.SearchAgentsRequest{Owner: "not-an-address"}

		t.Run("Should fail with a validation error on owner", func(t *testing.T) {
			t.Parallel()

			suite := WhenSearchingAgentsBeforeEach(t)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "owner", validationErr.Field)
		})
	})

	t.Run("Given the index cannot be searched", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")
		request := &requests.SearchAgentsRequest{Role: "expert"}

		initSuite := func(suite *WhenSearchingAgentsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, int64(0), sourceErr)
		}

		t.Run("Should fail with search_agents_failed", func(t *testing.T) {
			t.Parallel()

			suite := WhenSearchingAgentsBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			require.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "search_agents_failed", internalErr.Code)
		})
	})

	t.Run("Given agents match", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().WithEnsRecord("url", "https://scout.example").Build()
		request := &requests.SearchAgentsRequest{Query: " scout ", Role: "expert", Published: "true"}

		initSuite := func(suite *WhenSearchingAgentsTestingSuite) {
			published := true
			suite.mockAgentRepository.EXPECT().Search(gomock.Any(), &repositories.AgentSearchParams{
				Query: "scout", Role: "expert", Published: &published, Offset: 0, Limit: 10,
			}).Return([]*models.Agent{agent}, int64(11), nil)
		}

		t.Run("Should return the page with its total", func(t *testing.T) {
			t.Parallel()

			suite := WhenSearchingAgentsBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), request)

			require.NoError(t, err)
			assert.Equal(t, uint(11), response.TotalElements)
			assert.Equal(t, uint(1), response.Page)
			assert.Equal(t, uint(10), response.ElementsPerPage)
			require.Len(t, response.Items, 1)
			assert.Equal(t, agent.EnsName, response.Items[0].EnsName)
			assert.Equal(t, "https://scout.example", response.Items[0].Url)
			assert.Equal(t, "expert", response.Items[0].Role)
		})
	})
}
