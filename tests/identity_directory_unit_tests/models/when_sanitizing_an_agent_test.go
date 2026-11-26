package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenSanitizingAnAgent(t *testing.T) {
	t.Parallel()

	t.Run("Given on-chain strings holding NUL bytes and invalid UTF-8", func(t *testing.T) {
		t.Parallel()

		t.Run("Should clean its name, images, attributes and addresses", func(t *testing.T) {
			t.Parallel()

			agent := model_builders.NewAgentBuilder().
				WithName("sc\x00out").
				WithAttributes(map[string]string{"ro\x00le": "exp\xffert"}).
				WithAddresses(models.AgentAddress{Type: "ev\x00m", Value: "0x\x00ab"}).
				Build()
			agent.Images = map[string]string{"default": "cid\x00"}

			agent.Sanitize()

			assert.Equal(t, "scout", agent.Name)
			assert.Equal(t, map[string]string{"role": "exp�ert"}, agent.Attributes)
			assert.Equal(t, map[string]string{"default": "cid"}, agent.Images)
			assert.Equal(t, "evm", agent.Addresses[0].Type)
			assert.Equal(t, "0xab", agent.Addresses[0].Value)
		})
	})
}
