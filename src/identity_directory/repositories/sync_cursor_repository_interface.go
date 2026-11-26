package repositories

import "context"

type SyncCursorRepositoryInterface interface {
	// Get returns nil when the chain was never indexed.
	Get(ctx context.Context, chainId uint64) (*uint64, error)
	// Advance fails with ErrSyncCursorMoved unless the stored block still is expected (nil: none stored).
	Advance(ctx context.Context, chainId uint64, expected *uint64, block uint64) error
}
