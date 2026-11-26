package metrics

import (
	"context"
	"time"

	common_metrics "github.com/6022-labs/agentic-identity-directory/src/common/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const (
	meterName         = "agentic_identity_directory_http_agent_card"
	fetchDurationName = "identity_directory.agent_card.fetch.duration"
)

// AgentCardMetrics carries no host attribute: agent URLs are unbounded.
type AgentCardMetrics struct {
	fetchDurationHistogram metric.Float64Histogram
}

func NewAgentCardMetrics() *AgentCardMetrics {
	meter := otel.GetMeterProvider().Meter(meterName)
	fetchDurationHistogram, err := meter.Float64Histogram(
		fetchDurationName,
		metric.WithUnit("s"),
		metric.WithDescription("Duration of an agent card fetch"),
	)
	if err != nil {
		fetchDurationHistogram, _ = noop.NewMeterProvider().Meter(meterName).Float64Histogram(fetchDurationName)
	}
	return &AgentCardMetrics{fetchDurationHistogram: fetchDurationHistogram}
}

func (m *AgentCardMetrics) RecordFetch(ctx context.Context, duration time.Duration, err error) {
	attrs := []attribute.KeyValue{}
	if err != nil {
		attrs = append(attrs, attribute.String("error.type", common_metrics.ErrorType(err)))
	}
	m.fetchDurationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}
