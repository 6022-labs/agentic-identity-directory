package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingAnEnsName(t *testing.T) {
	t.Parallel()

	t.Run("Given an agent of a collection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should name it under its collection and chain in 6022.eth", func(t *testing.T) {
			t.Parallel()

			sut := services.NewEnsNameBuilder()

			assert.Equal(t, "scout.agents.137.6022.eth", sut.Build(137, "agents", "scout"))
		})
	})
}
