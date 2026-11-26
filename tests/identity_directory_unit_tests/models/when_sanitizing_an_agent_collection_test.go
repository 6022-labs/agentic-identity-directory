package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenSanitizingAnAgentCollection(t *testing.T) {
	t.Parallel()

	t.Run("Given on-chain strings holding NUL bytes and invalid UTF-8", func(t *testing.T) {
		t.Parallel()

		t.Run("Should clean its name, symbol and description", func(t *testing.T) {
			t.Parallel()

			collection := model_builders.NewAgentCollectionBuilder().
				WithName("sc\x00outs").
				WithDescription("great\xff scouts").
				Build()
			collection.Symbol = "SC\x00"

			collection.Sanitize()

			assert.Equal(t, "scouts", collection.Name)
			assert.Equal(t, "SC", collection.Symbol)
			assert.Equal(t, "great� scouts", collection.Description)
		})
	})
}
