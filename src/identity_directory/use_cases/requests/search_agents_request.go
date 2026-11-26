package requests

import (
	"strings"

	"github.com/6022-labs/agentic-identity-directory/src/identity_directory/repositories"

	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
	"github.com/6022-labs/agentic-identity-directory/src/common/helpers"
)

const maxQueryLength = 100

type SearchAgentsRequest struct {
	PaginateRequest
	Query             string `query:"query"             json:"-"`
	Role              string `query:"role"              json:"-"`
	ChainId           uint64 `query:"chainId"           json:"-"`
	CollectionAddress string `query:"collectionAddress" json:"-"`
	Owner             string `query:"owner"             json:"-"`
	Address           string `query:"address"           json:"-"`
	// Published is "true", "false" or empty for both.
	Published string `query:"published" json:"-"`
}

func (r *SearchAgentsRequest) ValidateAndSanitize() error {
	r.PaginateRequest.Sanitize()
	r.Query = strings.TrimSpace(r.Query)
	if len(r.Query) > maxQueryLength {
		return errors.NewValidationError("query", "must be 100 characters at most")
	}
	r.Role = strings.TrimSpace(r.Role)

	var err error
	if r.CollectionAddress, err = r.evmAddress("collectionAddress", r.CollectionAddress); err != nil {
		return err
	}
	if r.Owner, err = r.evmAddress("owner", r.Owner); err != nil {
		return err
	}
	r.Address = r.canonicalAddress(strings.TrimSpace(r.Address))

	r.Published = strings.ToLower(strings.TrimSpace(r.Published))
	if r.Published != "" && r.Published != "true" && r.Published != "false" {
		return errors.NewValidationError("published", "must be true or false")
	}
	return nil
}

// canonicalAddress lowercases the address forms the contracts store lowercase: evm and bech32 btc.
func (r *SearchAgentsRequest) canonicalAddress(address string) string {
	lowered := strings.ToLower(address)
	if strings.HasPrefix(lowered, "0x") || strings.HasPrefix(lowered, "bc1") ||
		strings.HasPrefix(lowered, "tb1") || strings.HasPrefix(lowered, "bcrt1") {
		return lowered
	}
	return address
}

func (r *SearchAgentsRequest) ToParams() *repositories.AgentSearchParams {
	var published *bool
	if r.Published != "" {
		value := r.Published == "true"
		published = &value
	}
	return &repositories.AgentSearchParams{
		Query:             r.Query,
		Role:              r.Role,
		ChainId:           r.ChainId,
		CollectionAddress: r.CollectionAddress,
		Owner:             r.Owner,
		Address:           r.Address,
		Published:         published,
		Offset:            r.Offset(),
		Limit:             int(*r.ElementsPerPage),
	}
}

// evmAddress lowercases an optional 0x hex address, the form the index stores.
func (r *SearchAgentsRequest) evmAddress(field, address string) (string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", nil
	}
	if !helpers.IsEthAddress(address) {
		return "", errors.NewValidationError(field, "must be a 0x-prefixed hex address")
	}
	return strings.ToLower(address), nil
}
