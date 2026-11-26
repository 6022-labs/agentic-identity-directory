package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingAnAgentUrl(t *testing.T) {
	t.Parallel()

	t.Run("Given an agent that published nothing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should have no url", func(t *testing.T) {
			t.Parallel()

			agent := model_builders.NewAgentBuilder().Build()

			assert.Empty(t, agent.Url())
		})
	})

	t.Run("Given an agent that published several records", func(t *testing.T) {
		t.Parallel()

		t.Run("Should read its url record", func(t *testing.T) {
			t.Parallel()

			agent := model_builders.NewAgentBuilder().
				WithEnsRecord("avatar", "ipfs://cid").
				WithEnsRecord("url", "https://scout.example").
				Build()

			assert.Equal(t, "https://scout.example", agent.Url())
		})
	})
}
