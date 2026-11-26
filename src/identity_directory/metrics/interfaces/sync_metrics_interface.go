package interfaces

import (
	"context"
	"time"
)

type SyncMetricsInterface interface {
	RecordPass(ctx context.Context, chainId uint64, duration time.Duration, err error)
	// RecordPosition records how far the chain is indexed and the block it can be indexed up to.
	RecordPosition(ctx context.Context, chainId, indexedBlock, headBlock uint64)
}
