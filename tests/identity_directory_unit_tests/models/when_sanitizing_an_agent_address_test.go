package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/stretchr/testify/assert"
)

func TestWhenSanitizingAnAgentAddress(t *testing.T) {
	t.Parallel()

	t.Run("Given a type and value holding NUL bytes", func(t *testing.T) {
		t.Parallel()

		t.Run("Should drop them", func(t *testing.T) {
			t.Parallel()

			address := models.AgentAddress{Type: "b\x00tc", Value: "bc1\x00q"}

			address.Sanitize()

			assert.Equal(t, models.AgentAddress{Type: "btc", Value: "bc1q"}, address)
		})
	})
}
