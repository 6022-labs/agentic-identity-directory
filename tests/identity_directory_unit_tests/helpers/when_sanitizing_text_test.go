package helpers_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenSanitizingText(t *testing.T) {
	t.Parallel()

	t.Run("Given text holding NUL bytes", func(t *testing.T) {
		t.Parallel()

		t.Run("Should drop them", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "scout", helpers.SanitizeText("sc\x00out\x00"))
		})
	})

	t.Run("Given invalid UTF-8", func(t *testing.T) {
		t.Parallel()

		t.Run("Should replace it", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "a�b", helpers.SanitizeText("a\xffb"))
		})
	})

	t.Run("Given keys and values to sanitize", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sanitize both", func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				map[string]string{"role": "expert"},
				helpers.SanitizeTexts(map[string]string{"ro\x00le": "exp\x00ert"}),
			)
		})
	})
}
