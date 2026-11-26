package services

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/6022-labs/agentic-identity-directory/src/common/helpers"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gowebpki/jcs"
)

const (
	cardSignatureAlgorithm = "ES256K"
	cardSignaturesMember   = "signatures"
	secp256k1CoordinateLen = 32
	ecdsaSignatureLen      = 64
	// maxCardSignatures bounds the verifications one hostile card can make us run; rotation needs a couple.
	maxCardSignatures = 4
)

type jwsProtectedHeader struct {
	Algorithm string       `json:"alg"`
	Jwk       secp256k1Jwk `json:"jwk"`
	KeyId     string       `json:"kid"`
	Type      string       `json:"typ"`
}

type secp256k1Jwk struct {
	Curve   string `json:"crv"`
	KeyType string `json:"kty"`
	X       string `json:"x"`
	Y       string `json:"y"`
}

// AgentCardSignatureVerifier checks detached JWS card signatures against the agent's on-chain EVM addresses.
type AgentCardSignatureVerifier struct{}

func NewAgentCardSignatureVerifier() *AgentCardSignatureVerifier {
	return &AgentCardSignatureVerifier{}
}

func (v *AgentCardSignatureVerifier) Verify(agentCard *models.AgentCard, signers []string) (bool, error) {
	if len(agentCard.Signatures) == 0 {
		return false, nil
	}
	if len(agentCard.Signatures) > maxCardSignatures {
		return false, fmt.Errorf(
			"agent card carries %d signatures, at most %d are checked",
			len(agentCard.Signatures),
			maxCardSignatures,
		)
	}

	canonical, err := v.canonicalizeWithoutSignatures(agentCard.Raw)
	if err != nil {
		return false, err
	}

	expected := map[common.Address]bool{}
	for _, signer := range signers {
		if helpers.IsEthAddress(signer) {
			expected[common.HexToAddress(signer)] = true
		}
	}
	if len(expected) == 0 {
		return false, fmt.Errorf("the agent has no evm address to verify its card against")
	}

	var lastErr error
	for _, signature := range agentCard.Signatures {
		if err := v.verifyOne(canonical, signature, expected); err != nil {
			lastErr = err
			continue
		}
		// One valid signature is enough; multiple exist for key rotation.
		return true, nil
	}
	return false, fmt.Errorf("no valid agent card signature: %w", lastErr)
}

// canonicalizeWithoutSignatures digests the received bytes, so unknown members and number formats survive.
func (v *AgentCardSignatureVerifier) canonicalizeWithoutSignatures(raw []byte) ([]byte, error) {
	members := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &members); err != nil {
		return nil, fmt.Errorf("agent card is not a JSON object: %w", err)
	}
	delete(members, cardSignaturesMember)

	unsigned, err := json.Marshal(members)
	if err != nil {
		return nil, err
	}

	canonical, err := jcs.Transform(unsigned)
	if err != nil {
		return nil, fmt.Errorf("failed to canonicalize agent card: %w", err)
	}
	return canonical, nil
}

func (v *AgentCardSignatureVerifier) verifyOne(
	canonical []byte,
	signature models.AgentCardSignature,
	expected map[common.Address]bool,
) error {
	header, err := v.decodeProtectedHeader(signature.Protected)
	if err != nil {
		return err
	}
	if header.Algorithm != cardSignatureAlgorithm {
		return fmt.Errorf("unsupported agent card signature algorithm %q", header.Algorithm)
	}

	publicKey, err := v.decodeJwk(header.Jwk)
	if err != nil {
		return err
	}

	// The jwk is self-asserted; the on-chain address is what makes it trustworthy.
	signer, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		return fmt.Errorf("invalid agent card signing key: %w", err)
	}
	if !expected[crypto.PubkeyToAddress(*signer)] {
		return fmt.Errorf(
			"agent card signed by %s, not by the agent",
			strings.ToLower(crypto.PubkeyToAddress(*signer).Hex()),
		)
	}

	rawSignature, err := base64.RawURLEncoding.DecodeString(signature.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode agent card signature: %w", err)
	}
	if len(rawSignature) != ecdsaSignatureLen {
		return fmt.Errorf("agent card signature must be %d bytes, got %d", ecdsaSignatureLen, len(rawSignature))
	}

	digest := sha256.Sum256([]byte(signature.Protected + "." + base64.RawURLEncoding.EncodeToString(canonical)))
	if !crypto.VerifySignature(publicKey, digest[:], rawSignature) {
		return fmt.Errorf("agent card signature does not match its content")
	}
	return nil
}

func (v *AgentCardSignatureVerifier) decodeProtectedHeader(protected string) (*jwsProtectedHeader, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(protected)
	if err != nil {
		return nil, fmt.Errorf("failed to decode protected header: %w", err)
	}

	var header jwsProtectedHeader
	if err := json.Unmarshal(decoded, &header); err != nil {
		return nil, fmt.Errorf("failed to parse protected header: %w", err)
	}
	return &header, nil
}

func (v *AgentCardSignatureVerifier) decodeJwk(jwk secp256k1Jwk) ([]byte, error) {
	x, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode jwk x: %w", err)
	}
	y, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode jwk y: %w", err)
	}
	if len(x) != secp256k1CoordinateLen || len(y) != secp256k1CoordinateLen {
		return nil, fmt.Errorf("jwk coordinates must be %d bytes each", secp256k1CoordinateLen)
	}

	return append(append([]byte{0x04}, x...), y...), nil
}
