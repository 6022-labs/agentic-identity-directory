package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingAnAgentRole(t *testing.T) {
	t.Parallel()

	t.Run("Given an agent minted with a role attribute", func(t *testing.T) {
		t.Parallel()

		t.Run("Should read it", func(t *testing.T) {
			t.Parallel()

			agent := model_builders.NewAgentBuilder().
				WithAttributes(map[string]string{"role": "facilitator", "mission": "trade"}).
				Build()

			assert.Equal(t, "facilitator", agent.Role())
		})
	})
}
