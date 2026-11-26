package requests

import (
	"strings"

	"github.com/6022-labs/agentic-identity-directory/src/common/errors"
)

const maxEnsNameLength = 255

type EnsNameRequest struct {
	EnsName string `uri:"ensName" json:"-"`
}

func (r *EnsNameRequest) ValidateAndSanitize() error {
	r.EnsName = strings.ToLower(strings.TrimSpace(r.EnsName))
	if r.EnsName == "" || len(r.EnsName) > maxEnsNameLength {
		return errors.NewValidationError("ensName", "must be an ENS name of 255 characters at most")
	}
	return nil
}
