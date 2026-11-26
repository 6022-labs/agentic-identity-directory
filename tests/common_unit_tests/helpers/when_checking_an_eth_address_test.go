package helpers_test

import (
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/common/helpers"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

func TestWhenCheckingAnEthAddress(t *testing.T) {
	t.Parallel()

	address := common_tests.RandomEthAddress().Hex()

	t.Run("Given a 0x-prefixed hex address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			assert.True(t, helpers.IsEthAddress(address))
		})
	})

	t.Run("Given the same address without its 0x prefix", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			assert.False(t, helpers.IsEthAddress(address[2:]))
		})
	})

	t.Run("Given a value that is not hex", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			assert.False(t, helpers.IsEthAddress("0xnot-an-address"))
		})
	})
}
