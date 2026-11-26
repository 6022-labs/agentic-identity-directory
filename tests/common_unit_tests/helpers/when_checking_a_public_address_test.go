package helpers_test

import (
	"net/netip"
	"testing"

	"github.com/6022-labs/agentic-identity-directory/src/common/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenCheckingAPublicAddress(t *testing.T) {
	t.Parallel()

	t.Run("Given public addresses", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept IPv4, IPv6 and IPv4-mapped public ones", func(t *testing.T) {
			t.Parallel()

			for _, address := range []string{
				"1.1.1.1",
				"8.8.8.8",
				"100.128.0.1",
				"198.20.0.1",
				"2606:4700:4700::1111",
				"2001:4860:4860::8888",
				"::ffff:1.1.1.1",
			} {
				assert.True(t, helpers.IsPublicAddress(netip.MustParseAddr(address)), address)
			}
		})
	})

	t.Run("Given non-public addresses", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse them", func(t *testing.T) {
			t.Parallel()

			for name, address := range map[string]string{
				"IPv4 loopback":            "127.0.0.1",
				"IPv6 loopback":            "::1",
				"RFC1918 10/8":             "10.1.2.3",
				"RFC1918 172.16/12":        "172.31.255.254",
				"RFC1918 192.168/16":       "192.168.1.1",
				"IPv6 unique local":        "fd00::1",
				"IPv4 link-local":          "169.254.169.254",
				"IPv6 link-local":          "fe80::1",
				"CGNAT 100.64/10":          "100.64.0.1",
				"CGNAT upper bound":        "100.127.255.254",
				"benchmarking 198.18/15":   "198.19.255.254",
				"this network 0/8":         "0.1.2.3",
				"IETF protocol 192.0.0/24": "192.0.0.8",
				"NAT64 64:ff9b::/96":       "64:ff9b::a00:1",
				"local NAT64 64:ff9b:1::":  "64:ff9b:1::1",
				"IPv4-mapped private":      "::ffff:10.0.0.1",
				"IPv4-mapped loopback":     "::ffff:127.0.0.1",
				"IPv4 unspecified":         "0.0.0.0",
				"IPv6 unspecified":         "::",
				"IPv4 multicast":           "224.0.0.1",
				"IPv6 multicast":           "ff02::1",
			} {
				assert.False(t, helpers.IsPublicAddress(netip.MustParseAddr(address)), name)
			}
		})
	})
}
