package kgox

import (
	"context"
	"testing"

	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
)

// TestMultipleSubscribersPrometheusCollection verifies that calling
// Subscriber() multiple times with different consumer groups does not produce
// duplicate otel_scope_info entries in Prometheus.
//
// Regression test: before the fix, Subscriber() called
//
//	p.mp.Meter("pubsubx_consumer", WithInstrumentationAttributes(consumerGroup))
//
// per consumer group, creating separate instrumentation scopes with the same
// name but different attributes. The Prometheus exporter emitted identical
// otel_scope_info gauges, erroring with "collected before with the same name
// and label values".
func TestMultipleSubscribersPrometheusCollection(t *testing.T) {
	ctx := context.Background()
	registry := promclient.NewRegistry()

	exporter, err := otelprom.New(otelprom.WithRegisterer(registry))
	require.NoError(t, err)

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	defer provider.Shutdown(ctx)

	l := getLogger()
	conf := &pubsubx.Config{Scope: "test-scope"}
	topics := []messagex.Topic{messagex.TopicFromName("test-topic")}

	// Construct a PubSub directly with the Prometheus-backed MeterProvider.
	// We bypass NewPubSub (which requires Kafka) since we only need the
	// Subscriber() → meter creation → metric recording path.
	ps := &PubSub{
		l:         l,
		mp:        provider,
		tp:        tracenoop.NewTracerProvider(),
		conf:      conf,
		consumers: make(map[messagex.ConsumerGroup]*consumer),
	}

	// Call Subscriber() with multiple different consumer groups, just as the
	// application does at startup.
	groups := []string{"group-a", "group-b", "group-c"}
	for _, group := range groups {
		_, err := ps.Subscriber(group, topics)
		require.NoError(t, err)
	}

	// Record metrics on each consumer — otel_scope_info is only emitted by
	// the Prometheus exporter when the scope's instruments have data.
	for _, c := range ps.consumers {
		if c.consumerMetric != nil {
			c.recordProcessingCount.Record(ctx, 1, metric.WithAttributes(
				attribute.String("topic", "test-topic"),
			))
		}
	}

	// Gather must succeed — duplicate otel_scope_info entries would cause an
	// error here.
	_, err = registry.Gather()
	require.NoError(t, err, "Prometheus gather failed; likely duplicate otel_scope_info from multiple meter scopes")
}
