package helpers

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// IsEthAddress reports whether s is a 0x-prefixed 20-byte hex address.
// The prefix guard keeps it stricter than common.IsHexAddress alone.
func IsEthAddress(s string) bool {
	return strings.HasPrefix(s, "0x") && common.IsHexAddress(s)
}
