package errors_test

import (
	stderrors "errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenClassifyingAnApplicationError(t *testing.T) {
	t.Parallel()

	t.Run("Given a validation error", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to 400 without a code or cause", func(t *testing.T) {
			t.Parallel()
			source := apperrors.NewValidationError("name", "required")

			classified := apperrors.NewClassifiedError(source)

			assert.Equal(t, http.StatusBadRequest, classified.Status)
			assert.Empty(t, classified.Code)
			assert.Same(t, source, classified.Body)
			assert.Nil(t, classified.Cause)
			assert.False(t, classified.Failed())
		})
	})

	t.Run("Given an unavailable error", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to a failed 502 carrying its cause", func(t *testing.T) {
			t.Parallel()
			cause := stderrors.New("node down")

			classified := apperrors.NewClassifiedError(apperrors.NewUnavailableError("provider_down", "down", cause))

			assert.Equal(t, http.StatusBadGateway, classified.Status)
			assert.Equal(t, "provider_down", classified.Code)
			assert.Same(t, cause, classified.Cause)
			assert.True(t, classified.Failed())
		})
	})

	t.Run("Given an error the application did not classify", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map it to a failed 500 internal_error keeping it as the cause", func(t *testing.T) {
			t.Parallel()
			source := stderrors.New("boom")

			classified := apperrors.NewClassifiedError(source)

			assert.Equal(t, http.StatusInternalServerError, classified.Status)
			assert.Equal(t, "internal_error", classified.Code)
			assert.Same(t, source, classified.Cause)
			assert.True(t, classified.Failed())
		})
	})
}
