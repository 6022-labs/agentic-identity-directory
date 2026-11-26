package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenListingAnAgentEvmAddresses(t *testing.T) {
	t.Parallel()

	t.Run("Given an agent with several address types", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list only its evm addresses, in order", func(t *testing.T) {
			t.Parallel()

			firstEvm := common_tests.RandomLowerEthAddress()
			secondEvm := common_tests.RandomLowerEthAddress()
			agent := model_builders.NewAgentBuilder().WithAddresses(
				models.AgentAddress{Type: "evm", Value: firstEvm},
				models.AgentAddress{Type: "btc", Value: "bc1qexample"},
				models.AgentAddress{Type: "evm", Value: secondEvm},
			).Build()

			assert.Equal(t, []string{firstEvm, secondEvm}, agent.EvmAddresses())
		})
	})
}
