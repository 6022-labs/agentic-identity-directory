package requests_test

import (
	"strings"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/use_cases/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhenValidatingAnEnsNameRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a name longer than 255 characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail on ensName", func(t *testing.T) {
			t.Parallel()

			request := &requests.EnsNameRequest{EnsName: strings.Repeat("a", 256)}

			var validationErr *apperrors.ValidationError
			require.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "ensName", validationErr.Field)
		})
	})

	t.Run("Given a padded name in capitals", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim and lowercase it", func(t *testing.T) {
			t.Parallel()

			request := &requests.EnsNameRequest{EnsName: " Scout.Agents.137.6022.ETH "}

			require.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, "scout.agents.137.6022.eth", request.EnsName)
		})
	})
}
