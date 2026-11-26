package transports

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"syscall"

	"github.com/6022-labs/agentic-identity-directory/src/common/helpers"
	"github.com/6022-labs/agentic-identity-directory/src/identity_directory_http_agent_card/settings"
)

// maxResponseBytes bounds what a hostile agent can make us buffer and canonicalize.
const maxResponseBytes = 1 << 20

var (
	ErrAddressRefused   = errors.New("agent card fetch refuses a non-public address")
	ErrResponseTooLarge = fmt.Errorf("agent card answer exceeds %d bytes", maxResponseBytes)
)

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
	response, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	response.Body = &cappedBody{body: response.Body, remaining: maxResponseBytes}
	return response, nil
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

// cappedBody fails the read once the answer goes past the cap, rather than truncating it into a valid-looking card.
type cappedBody struct {
	body      io.ReadCloser
	remaining int64
}

func (b *cappedBody) Read(buffer []byte) (int, error) {
	if b.remaining < 0 {
		return 0, ErrResponseTooLarge
	}
	if int64(len(buffer)) > b.remaining+1 {
		buffer = buffer[:b.remaining+1]
	}
	read, err := b.body.Read(buffer)
	b.remaining -= int64(read)
	if b.remaining < 0 {
		return 0, ErrResponseTooLarge
	}
	return read, err
}

func (b *cappedBody) Close() error {
	return b.body.Close()
}
