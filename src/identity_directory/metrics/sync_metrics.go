package metrics

import (
	"context"
	"strconv"
	"time"

	common_metrics "github.com/6022-labs/agentic-identity-directory/src/common/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const (
	meterName = "agentic_identity_directory"

	passDurationName = "identity_directory.sync.pass.duration"
	indexedBlockName = "identity_directory.sync.indexed_block"
	headBlockName    = "identity_directory.sync.head_block"
)

type SyncMetrics struct {
	passDurationHistogram metric.Float64Histogram
	indexedBlockGauge     metric.Int64Gauge
	headBlockGauge        metric.Int64Gauge
}

func NewSyncMetrics() *SyncMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	passDurationHistogram, err := meter.Float64Histogram(
		passDurationName,
		metric.WithUnit("s"),
		metric.WithDescription("Duration of a chain sync pass"),
	)
	if err != nil {
		passDurationHistogram, _ = noopMeter.Float64Histogram(passDurationName)
	}

	indexedBlockGauge, err := meter.Int64Gauge(
		indexedBlockName,
		metric.WithUnit("{block}"),
		metric.WithDescription("Last block indexed on the chain"),
	)
	if err != nil {
		indexedBlockGauge, _ = noopMeter.Int64Gauge(indexedBlockName)
	}

	headBlockGauge, err := meter.Int64Gauge(
		headBlockName,
		metric.WithUnit("{block}"),
		metric.WithDescription("Last confirmed block of the chain"),
	)
	if err != nil {
		headBlockGauge, _ = noopMeter.Int64Gauge(headBlockName)
	}

	return &SyncMetrics{
		passDurationHistogram: passDurationHistogram,
		indexedBlockGauge:     indexedBlockGauge,
		headBlockGauge:        headBlockGauge,
	}
}

func (m *SyncMetrics) RecordPass(ctx context.Context, chainId uint64, duration time.Duration, err error) {
	attrs := []attribute.KeyValue{m.chainAttribute(chainId)}
	if err != nil {
		attrs = append(attrs, attribute.String("error.type", common_metrics.ErrorType(err)))
	}
	m.passDurationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

func (m *SyncMetrics) RecordPosition(ctx context.Context, chainId, indexedBlock, headBlock uint64) {
	option := metric.WithAttributes(m.chainAttribute(chainId))
	m.indexedBlockGauge.Record(ctx, int64(indexedBlock), option)
	m.headBlockGauge.Record(ctx, int64(headBlock), option)
}

func (m *SyncMetrics) chainAttribute(chainId uint64) attribute.KeyValue {
	return attribute.String("chain.id", strconv.FormatUint(chainId, 10))
}
