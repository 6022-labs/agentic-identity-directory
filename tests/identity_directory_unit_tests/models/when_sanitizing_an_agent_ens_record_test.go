package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/stretchr/testify/assert"
)

func TestWhenSanitizingAnAgentEnsRecord(t *testing.T) {
	t.Parallel()

	t.Run("Given a key and value holding NUL bytes and invalid UTF-8", func(t *testing.T) {
		t.Parallel()

		t.Run("Should clean both", func(t *testing.T) {
			t.Parallel()

			record := models.AgentEnsRecord{Key: "u\x00rl", Value: "https://scout\xff"}

			record.Sanitize()

			assert.Equal(t, "url", record.Key)
			assert.Equal(t, "https://scout�", record.Value)
		})
	})
}
