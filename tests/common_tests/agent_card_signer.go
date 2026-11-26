package common_tests

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gowebpki/jcs"
)

// SignAgentCard adds a detached ES256K signature to the card, the way agent-node signs its /.well-known/6022.
func SignAgentCard(card map[string]any, key *ecdsa.PrivateKey) []byte {
	unsigned, err := json.Marshal(card)
	if err != nil {
		panic(err)
	}
	canonical, err := jcs.Transform(unsigned)
	if err != nil {
		panic(err)
	}

	encode := base64.RawURLEncoding.EncodeToString
	publicKey := crypto.FromECDSAPub(&key.PublicKey)
	header, err := json.Marshal(map[string]any{
		"alg": "ES256K",
		"jwk": map[string]string{
			"crv": "secp256k1",
			"kty": "EC",
			"x":   encode(publicKey[1:33]),
			"y":   encode(publicKey[33:]),
		},
		"kid": crypto.PubkeyToAddress(key.PublicKey).Hex(),
		"typ": "JOSE",
	})
	if err != nil {
		panic(err)
	}
	protected := encode(header)
	digest := sha256.Sum256([]byte(protected + "." + encode(canonical)))
	signature, err := crypto.Sign(digest[:], key)
	if err != nil {
		panic(err)
	}

	signed := map[string]any{}
	for key, value := range card {
		signed[key] = value
	}
	signed["signatures"] = []map[string]string{{"protected": protected, "signature": encode(signature[:64])}}
	raw, err := json.Marshal(signed)
	if err != nil {
		panic(err)
	}
	return raw
}
