package interfaces

import "context"

// DaemonInterface is a background entry point into the domain, started and stopped by the host.
type DaemonInterface interface {
	Start(ctx context.Context)
	// Stop returns once the running work ended.
	Stop()
}
