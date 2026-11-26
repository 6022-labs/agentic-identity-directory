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

type WhenListingCollectionsTestingSuite struct {
	mockAgentCollectionRepository *repositories_mocks.MockAgentCollectionRepositoryInterface
	sut                           *use_cases.ListCollections
}

func WhenListingCollectionsBeforeEach(t *testing.T) *WhenListingCollectionsTestingSuite {
	mockAgentCollectionRepository := repositories_mocks.NewMockAgentCollectionRepositoryInterface(
		gomock.NewController(t),
	)
	return &WhenListingCollectionsTestingSuite{
		mockAgentCollectionRepository: mockAgentCollectionRepository,
		sut:                           use_cases.NewListCollections(mockAgentCollectionRepository),
	}
}

func TestWhenListingCollections(t *testing.T) {
	t.Parallel()

	t.Run("Given the index cannot be read", func(t *testing.T) {
		t.Parallel()

		sourceErr := errors.New("db down")

		initSuite := func(suite *WhenListingCollectionsTestingSuite) {
			suite.mockAgentCollectionRepository.EXPECT().
				List(gomock.Any(), gomock.Any()).
				Return(nil, int64(0), sourceErr)
		}

		t.Run("Should fail with list_collections_failed", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingCollectionsBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), &requests.ListCollectionsRequest{})

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			require.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_collections_failed", internalErr.Code)
		})
	})

	t.Run("Given indexed collections", func(t *testing.T) {
		t.Parallel()

		collection := model_builders.NewAgentCollectionBuilder().Build()

		initSuite := func(suite *WhenListingCollectionsTestingSuite) {
			suite.mockAgentCollectionRepository.EXPECT().
				List(gomock.Any(), &repositories.AgentCollectionListParams{ChainId: 80002, Offset: 10, Limit: 10}).
				Return([]*models.AgentCollection{collection}, int64(11), nil)
		}

		t.Run("Should return the page with its total", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingCollectionsBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background(), &requests.ListCollectionsRequest{
				PaginateRequest: requests.PaginateRequest{Page: new(uint(2))},
				ChainId:         80002,
			})

			require.NoError(t, err)
			assert.Equal(t, uint(11), response.TotalElements)
			assert.Equal(t, uint(2), response.Page)
			assert.Equal(t, uint(10), response.ElementsPerPage)
			require.Len(t, response.Items, 1)
			assert.Equal(t, collection.Address, response.Items[0].Address)
		})
	})
}
