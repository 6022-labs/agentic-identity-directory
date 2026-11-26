package interfaces

import (
	"context"
	"time"
)

type AgentCardMetricsInterface interface {
	RecordFetch(ctx context.Context, duration time.Duration, err error)
}
