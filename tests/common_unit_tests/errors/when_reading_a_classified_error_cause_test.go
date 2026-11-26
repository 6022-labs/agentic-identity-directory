package errors_test

import (
	stderrors "errors"
	"strings"
	"testing"
	"unicode/utf8"

	apperrors "github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingAClassifiedErrorCause(t *testing.T) {
	t.Parallel()

	t.Run("Given no cause", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be empty", func(t *testing.T) {
			t.Parallel()

			classified := apperrors.NewClassifiedError(apperrors.NewNotFoundError("missing", "missing"))

			assert.Empty(t, classified.CauseMessage())
		})
	})

	t.Run("Given a short cause", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it whole", func(t *testing.T) {
			t.Parallel()

			classified := apperrors.NewClassifiedError(stderrors.New("boom"))

			assert.Equal(t, "boom", classified.CauseMessage())
		})
	})

	t.Run("Given a cause a remote dependency made huge", func(t *testing.T) {
		t.Parallel()

		t.Run("Should cut it to valid UTF-8 within the log limit", func(t *testing.T) {
			t.Parallel()
			huge := "x" + strings.Repeat("é", 1<<20)

			message := apperrors.NewClassifiedError(stderrors.New(huge)).CauseMessage()

			assert.LessOrEqual(t, len(message), 2048)
			assert.True(t, utf8.ValidString(message))
			assert.True(t, strings.HasPrefix(huge, message))
		})
	})
}
