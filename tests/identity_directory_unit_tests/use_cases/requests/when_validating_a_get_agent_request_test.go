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

func TestWhenValidatingAGetAgentRequest(t *testing.T) {
	t.Parallel()

	collectionAddress := common_tests.RandomEthAddress().Hex()

	t.Run("Given no chain id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail on chainId", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAgentRequest{CollectionAddress: collectionAddress, TokenId: "1"}

			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "chainId", validationErr.Field)
		})
	})

	t.Run("Given a collection address without 0x", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail on collectionAddress", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAgentRequest{ChainId: 137, CollectionAddress: collectionAddress[2:], TokenId: "1"}

			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "collectionAddress", validationErr.Field)
		})
	})

	t.Run("Given a token id of zero", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail on tokenId", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAgentRequest{ChainId: 137, CollectionAddress: collectionAddress, TokenId: "0"}

			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "tokenId", validationErr.Field)
		})
	})

	t.Run("Given a token id beyond uint256", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail on tokenId", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAgentRequest{
				ChainId:           137,
				CollectionAddress: collectionAddress,
				TokenId:           "115792089237316195423570985008687907853269984665640564039457584007913129639936",
			}

			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "tokenId", validationErr.Field)
		})
	})

	t.Run("Given a valid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should lowercase the address and normalize the token id", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAgentRequest{ChainId: 137, CollectionAddress: collectionAddress, TokenId: " 0042"}

			require.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, strings.ToLower(collectionAddress), request.CollectionAddress)
			assert.Equal(t, "42", request.TokenId)
		})
	})
}
