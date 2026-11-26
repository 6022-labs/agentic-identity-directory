package requests

import (
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/common/helpers"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/models"
)

var maxTokenId = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

type GetAgentRequest struct {
	ChainId           uint64 `uri:"chainId"           json:"-"`
	CollectionAddress string `uri:"collectionAddress" json:"-"`
	TokenId           string `uri:"tokenId"           json:"-"`
}

func (r *GetAgentRequest) ValidateAndSanitize() error {
	if r.ChainId == 0 {
		return errors.NewValidationError("chainId", "is required")
	}
	r.CollectionAddress = strings.TrimSpace(r.CollectionAddress)
	if !helpers.IsEthAddress(r.CollectionAddress) {
		return errors.NewValidationError("collectionAddress", "must be a 0x-prefixed hex address")
	}
	r.CollectionAddress = strings.ToLower(r.CollectionAddress)

	tokenId, ok := new(big.Int).SetString(strings.TrimSpace(r.TokenId), 10)
	if !ok || tokenId.Sign() <= 0 || tokenId.Cmp(maxTokenId) > 0 {
		return errors.NewValidationError("tokenId", "must be a positive decimal uint256")
	}
	r.TokenId = tokenId.String()
	return nil
}

func (r *GetAgentRequest) AgentRef() models.AgentRef {
	return models.AgentRef{CollectionAddress: r.CollectionAddress, TokenId: r.TokenId}
}
