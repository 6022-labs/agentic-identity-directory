package requests_test

import (
	"strings"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenValidatingASearchAgentsRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a query longer than 100 characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail on query", func(t *testing.T) {
			t.Parallel()

			request := &requests.SearchAgentsRequest{Query: strings.Repeat("a", 101)}

			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "query", validationErr.Field)
		})
	})

	t.Run("Given a collection address that is not hex", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail on collectionAddress", func(t *testing.T) {
			t.Parallel()

			request := &requests.SearchAgentsRequest{CollectionAddress: "0xnope"}

			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "collectionAddress", validationErr.Field)
		})
	})

	t.Run("Given a published filter that is not a boolean", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail on published", func(t *testing.T) {
			t.Parallel()

			request := &requests.SearchAgentsRequest{Published: "maybe"}

			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "published", validationErr.Field)
		})
	})

	t.Run("Given an empty request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should default to the first page of 10", func(t *testing.T) {
			t.Parallel()

			request := &requests.SearchAgentsRequest{}

			require.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, uint(1), *request.Page)
			assert.Equal(t, uint(10), *request.ElementsPerPage)
		})
	})

	t.Run("Given checksummed addresses", func(t *testing.T) {
		t.Parallel()

		owner := common_tests.RandomEthAddress().Hex()

		t.Run("Should lowercase them", func(t *testing.T) {
			t.Parallel()

			request := &requests.SearchAgentsRequest{Owner: owner, Address: " " + owner + " "}

			require.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, strings.ToLower(owner), request.Owner)
			assert.Equal(t, strings.ToLower(owner), request.Address)
		})
	})

	t.Run("Given a bech32 bitcoin address in capitals", func(t *testing.T) {
		t.Parallel()

		t.Run("Should lowercase it", func(t *testing.T) {
			t.Parallel()

			request := &requests.SearchAgentsRequest{Address: "BC1QAR0SRRR7XFKVY5L643LYDNW9RE59GTZZWF5MDQ"}

			require.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq", request.Address)
		})
	})

	t.Run("Given a base58 address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep its case", func(t *testing.T) {
			t.Parallel()

			request := &requests.SearchAgentsRequest{Address: "1BoatSLRHtKNngkdXEeobR76b53LETtpyT"}

			require.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, "1BoatSLRHtKNngkdXEeobR76b53LETtpyT", request.Address)
		})
	})

}
