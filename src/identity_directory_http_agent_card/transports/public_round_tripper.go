package transports

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"syscall"

	"github.com/6022-labs/agentic-identity-directory/src/common/helpers"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/settings"
)

var ErrAddressRefused = errors.New("agent card fetch refuses a non-public address")

// PublicRoundTripper dials agent URLs, which anyone can publish, only on public addresses unless allowed.
type PublicRoundTripper struct {
	base *http.Transport
}

func NewPublicRoundTripper(agentCardSettings *settings.AgentCardSettings) *PublicRoundTripper {
	t := &PublicRoundTripper{}
	dialer := &net.Dialer{Timeout: agentCardSettings.Timeout}
	if !agentCardSettings.AllowPrivateNetworks {
		dialer.Control = t.refuseNonPublic
	}
	t.base = http.DefaultTransport.(*http.Transport).Clone()
	t.base.DialContext = dialer.DialContext
	t.base.Proxy = nil
	return t
}

func (t *PublicRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.base.RoundTrip(req)
}

// refuseNonPublic checks the address actually dialed, after DNS, so a rebinding hostname cannot slip through.
func (t *PublicRoundTripper) refuseNonPublic(_ string, address string, _ syscall.RawConn) error {
	addressPort, err := netip.ParseAddrPort(address)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrAddressRefused, address)
	}
	if !helpers.IsPublicAddress(addressPort.Addr()) {
		return fmt.Errorf("%w: %s", ErrAddressRefused, addressPort.Addr().Unmap())
	}
	return nil
}
