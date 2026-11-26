package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenConvertingASearchAgentsRequestToParams(t *testing.T) {
	t.Parallel()

	t.Run("Given no published filter", func(t *testing.T) {
		t.Parallel()

		t.Run("Should leave published unset", func(t *testing.T) {
			t.Parallel()

			request := &requests.SearchAgentsRequest{}
			require.NoError(t, request.ValidateAndSanitize())

			assert.Nil(t, request.ToParams().Published)
		})
	})

	t.Run("Given the second page of 20 unpublished experts", func(t *testing.T) {
		t.Parallel()

		t.Run("Should carry every filter with the page offset and limit", func(t *testing.T) {
			t.Parallel()

			request := &requests.SearchAgentsRequest{
				PaginateRequest: requests.PaginateRequest{Page: new(uint(2)), ElementsPerPage: new(uint(20))},
				Query:           "scout",
				Role:            "expert",
				ChainId:         137,
				Published:       "FALSE",
			}
			require.NoError(t, request.ValidateAndSanitize())

			assert.Equal(t, &repositories.AgentSearchParams{
				Query: "scout", Role: "expert", ChainId: 137, Published: new(false), Offset: 20, Limit: 20,
			}, request.ToParams())
		})
	})
}
