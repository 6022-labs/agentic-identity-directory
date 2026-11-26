package services_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/services"
	"github.com/6022-labs/agentic-identity-directory/tests/common_tests"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhenVerifyingAnAgentCardSignatureTestingSuite struct {
	sut *services.AgentCardSignatureVerifier
}

func WhenVerifyingAnAgentCardSignatureBeforeEach() *WhenVerifyingAnAgentCardSignatureTestingSuite {
	return &WhenVerifyingAnAgentCardSignatureTestingSuite{sut: services.NewAgentCardSignatureVerifier()}
}

func (s *WhenVerifyingAnAgentCardSignatureTestingSuite) card(t *testing.T, raw []byte) *models.AgentCard {
	var decoded struct {
		Signatures []models.AgentCardSignature `json:"signatures"`
	}
	require.NoError(t, json.Unmarshal(raw, &decoded))
	return &models.AgentCard{Protocol: "6022", Signatures: decoded.Signatures, Raw: raw}
}

func TestWhenVerifyingAnAgentCardSignature(t *testing.T) {
	t.Parallel()

	content := map[string]any{"protocol": "6022", "protocolVersion": "0.1", "name": "scout"}

	t.Run("Given an unsigned card", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not verify it, without an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnAgentCardSignatureBeforeEach()

			verified, err := suite.sut.Verify(&models.AgentCard{Raw: []byte(`{"protocol":"6022"}`)}, nil)

			assert.NoError(t, err)
			assert.False(t, verified)
		})
	})

	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex())

	t.Run("Given a card carrying more than 16 signatures", func(t *testing.T) {
		t.Parallel()

		signed := map[string]any{}
		require.NoError(t, json.Unmarshal(common_tests.SignAgentCard(content, key), &signed))
		signature := signed["signatures"].([]any)[0]
		signed["signatures"] = slices.Repeat([]any{signature}, 17)
		raw, err := json.Marshal(signed)
		require.NoError(t, err)

		t.Run("Should refuse it before verifying any", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnAgentCardSignatureBeforeEach()

			verified, err := suite.sut.Verify(suite.card(t, raw), []string{signer})

			assert.ErrorContains(t, err, "at most 16 are checked")
			assert.False(t, verified)
		})
	})

	t.Run("Given an agent without an evm address", func(t *testing.T) {
		t.Parallel()

		raw := common_tests.SignAgentCard(content, key)

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnAgentCardSignatureBeforeEach()

			verified, err := suite.sut.Verify(suite.card(t, raw), []string{"bc1qexample"})

			assert.ErrorContains(t, err, "no evm address")
			assert.False(t, verified)
		})
	})

	t.Run("Given a card signed by another key", func(t *testing.T) {
		t.Parallel()

		raw := common_tests.SignAgentCard(content, key)

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnAgentCardSignatureBeforeEach()

			verified, err := suite.sut.Verify(suite.card(t, raw), []string{common_tests.RandomLowerEthAddress()})

			assert.ErrorContains(t, err, "not by the agent")
			assert.False(t, verified)
		})
	})

	t.Run("Given a card changed after it was signed", func(t *testing.T) {
		t.Parallel()

		raw := []byte(strings.Replace(string(common_tests.SignAgentCard(content, key)), "scout", "thief", 1))

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnAgentCardSignatureBeforeEach()

			verified, err := suite.sut.Verify(suite.card(t, raw), []string{signer})

			assert.ErrorContains(t, err, "does not match its content")
			assert.False(t, verified)
		})
	})

	t.Run("Given a broken signature before a valid one", func(t *testing.T) {
		t.Parallel()

		signed := map[string]any{}
		require.NoError(t, json.Unmarshal(common_tests.SignAgentCard(content, key), &signed))
		valid := signed["signatures"].([]any)[0]
		signed["signatures"] = []any{map[string]any{"protected": "not-base64!!", "signature": "x"}, valid}
		raw, err := json.Marshal(signed)
		require.NoError(t, err)

		t.Run("Should verify it, since keys rotate", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnAgentCardSignatureBeforeEach()

			verified, err := suite.sut.Verify(suite.card(t, raw), []string{signer})

			assert.NoError(t, err)
			assert.True(t, verified)
		})
	})

	t.Run("Given a card signed by one of the agent's addresses", func(t *testing.T) {
		t.Parallel()

		raw := common_tests.SignAgentCard(content, key)

		t.Run("Should verify it", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingAnAgentCardSignatureBeforeEach()

			verified, err := suite.sut.Verify(
				suite.card(t, raw),
				[]string{common_tests.RandomLowerEthAddress(), signer},
			)

			assert.NoError(t, err)
			assert.True(t, verified)
		})
	})
}
