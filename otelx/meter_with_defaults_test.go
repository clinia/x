package otelx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/semconv/v1.38.0/dbconv"
)

func TestInstrumentWithDefaults_Int64Histogram(t *testing.T) {
	ctx := context.Background()

	t.Run("defaults applied", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		hist, err := meter.Int64Histogram("h")
		require.NoError(t, err)

		wrapped := NewInstrumentWithDefaults(hist,
			metric.WithAttributes(attribute.String("env", "test")),
		)
		wrapped.Record(ctx, 10)

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findHistogramDataPoints[int64](t, rm, "h")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("env", "test"))
	})

	t.Run("defaults merged with extra", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		hist, err := meter.Int64Histogram("h")
		require.NoError(t, err)

		wrapped := NewInstrumentWithDefaults(hist,
			metric.WithAttributes(attribute.String("env", "test")),
		)
		wrapped.Record(ctx, 20, metric.WithAttributes(attribute.String("extra", "val")))

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findHistogramDataPoints[int64](t, rm, "h")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("env", "test"))
		assertHasAttr(t, dp[0].Attributes, attribute.String("extra", "val"))
	})
}

func TestInstrumentWithDefaults_Float64Histogram(t *testing.T) {
	ctx := context.Background()

	t.Run("defaults applied", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		hist, err := meter.Float64Histogram("h")
		require.NoError(t, err)

		wrapped := NewInstrumentWithDefaults(hist,
			metric.WithAttributes(attribute.String("env", "test")),
		)
		wrapped.Record(ctx, 1.5)

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findHistogramDataPoints[float64](t, rm, "h")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("env", "test"))
	})

	t.Run("defaults merged with extra", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		hist, err := meter.Float64Histogram("h")
		require.NoError(t, err)

		wrapped := NewInstrumentWithDefaults(hist,
			metric.WithAttributes(attribute.String("env", "test")),
		)
		wrapped.Record(ctx, 2.5, metric.WithAttributes(attribute.String("extra", "val")))

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findHistogramDataPoints[float64](t, rm, "h")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("env", "test"))
		assertHasAttr(t, dp[0].Attributes, attribute.String("extra", "val"))
	})
}

func TestInstrumentWithDefaults_Int64Gauge(t *testing.T) {
	ctx := context.Background()

	t.Run("defaults applied", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		gauge, err := meter.Int64Gauge("g")
		require.NoError(t, err)

		wrapped := NewInstrumentWithDefaults(gauge,
			metric.WithAttributes(attribute.String("env", "test")),
		)
		wrapped.Record(ctx, 42)

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findGaugeDataPoints[int64](t, rm, "g")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("env", "test"))
	})

	t.Run("defaults merged with extra", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		gauge, err := meter.Int64Gauge("g")
		require.NoError(t, err)

		wrapped := NewInstrumentWithDefaults(gauge,
			metric.WithAttributes(attribute.String("env", "test")),
		)
		wrapped.Record(ctx, 99, metric.WithAttributes(attribute.String("extra", "val")))

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findGaugeDataPoints[int64](t, rm, "g")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("env", "test"))
		assertHasAttr(t, dp[0].Attributes, attribute.String("extra", "val"))
	})
}

func TestInstrumentWithDefaults_Float64Gauge(t *testing.T) {
	ctx := context.Background()

	t.Run("defaults applied", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		gauge, err := meter.Float64Gauge("g")
		require.NoError(t, err)

		wrapped := NewInstrumentWithDefaults(gauge,
			metric.WithAttributes(attribute.String("env", "test")),
		)
		wrapped.Record(ctx, 3.14)

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findGaugeDataPoints[float64](t, rm, "g")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("env", "test"))
	})

	t.Run("defaults merged with extra", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		gauge, err := meter.Float64Gauge("g")
		require.NoError(t, err)

		wrapped := NewInstrumentWithDefaults(gauge,
			metric.WithAttributes(attribute.String("env", "test")),
		)
		wrapped.Record(ctx, 2.72, metric.WithAttributes(attribute.String("extra", "val")))

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findGaugeDataPoints[float64](t, rm, "g")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("env", "test"))
		assertHasAttr(t, dp[0].Attributes, attribute.String("extra", "val"))
	})
}

func TestInstrumentWithDefaults_MultipleDefaults(t *testing.T) {
	ctx := context.Background()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	meter := provider.Meter("test")

	gauge, err := meter.Int64Gauge("g")
	require.NoError(t, err)

	wrapped := NewInstrumentWithDefaults(gauge,
		metric.WithAttributes(attribute.String("service", "api")),
		metric.WithAttributes(attribute.String("env", "prod")),
	)
	wrapped.Record(ctx, 1)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))

	dp := findGaugeDataPoints[int64](t, rm, "g")
	require.Len(t, dp, 1)
	assertHasAttr(t, dp[0].Attributes, attribute.String("service", "api"))
	assertHasAttr(t, dp[0].Attributes, attribute.String("env", "prod"))
}

func TestInstrumentWithDefaults_NoDefaults(t *testing.T) {
	ctx := context.Background()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	meter := provider.Meter("test")

	gauge, err := meter.Int64Gauge("g")
	require.NoError(t, err)

	wrapped := NewInstrumentWithDefaults(gauge)
	wrapped.Record(ctx, 7, metric.WithAttributes(attribute.String("only", "this")))

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))

	dp := findGaugeDataPoints[int64](t, rm, "g")
	require.Len(t, dp, 1)
	assertHasAttr(t, dp[0].Attributes, attribute.String("only", "this"))
}

// --- helpers ---

func findHistogramDataPoints[N int64 | float64](t *testing.T, rm metricdata.ResourceMetrics, name string) []metricdata.HistogramDataPoint[N] {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				h, ok := m.Data.(metricdata.Histogram[N])
				require.True(t, ok, "metric %q is not a Histogram", name)
				return h.DataPoints
			}
		}
	}
	t.Fatalf("metric %q not found", name)
	return nil
}

func findGaugeDataPoints[N int64 | float64](t *testing.T, rm metricdata.ResourceMetrics, name string) []metricdata.DataPoint[N] {
	t.Helper()
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				g, ok := m.Data.(metricdata.Gauge[N])
				require.True(t, ok, "metric %q is not a Gauge", name)
				return g.DataPoints
			}
		}
	}
	t.Fatalf("metric %q not found", name)
	return nil
}

// --- NewInstrumentWithDBSystemAndDefaults tests ---

func TestInstrumentWithDBSystemAndDefaults_Float64(t *testing.T) {
	ctx := context.Background()

	t.Run("defaults applied", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		dur, err := dbconv.NewClientOperationDuration(meter)
		require.NoError(t, err)

		wrapped := NewInstrumentWithDBSystemAndDefaults(dur,
			attribute.String("db.namespace", "mydb"),
		)
		wrapped.Record(ctx, 0.5, dbconv.SystemNamePostgreSQL)

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findHistogramDataPoints[float64](t, rm, "db.client.operation.duration")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("db.namespace", "mydb"))
		assertHasAttr(t, dp[0].Attributes, attribute.String("db.system.name", "postgresql"))
	})

	t.Run("defaults merged with extra", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		dur, err := dbconv.NewClientOperationDuration(meter)
		require.NoError(t, err)

		wrapped := NewInstrumentWithDBSystemAndDefaults(dur,
			attribute.String("db.namespace", "mydb"),
		)
		wrapped.Record(ctx, 1.2, dbconv.SystemNamePostgreSQL,
			attribute.String("db.operation.name", "SELECT"),
		)

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findHistogramDataPoints[float64](t, rm, "db.client.operation.duration")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("db.namespace", "mydb"))
		assertHasAttr(t, dp[0].Attributes, attribute.String("db.system.name", "postgresql"))
		assertHasAttr(t, dp[0].Attributes, attribute.String("db.operation.name", "SELECT"))
	})
}

func TestInstrumentWithDBSystemAndDefaults_Int64(t *testing.T) {
	ctx := context.Background()

	t.Run("defaults applied", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		rows, err := dbconv.NewClientResponseReturnedRows(meter)
		require.NoError(t, err)

		wrapped := NewInstrumentWithDBSystemAndDefaults(rows,
			attribute.String("db.namespace", "mydb"),
		)
		wrapped.Record(ctx, 42, dbconv.SystemNameMySQL)

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findHistogramDataPoints[int64](t, rm, "db.client.response.returned_rows")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("db.namespace", "mydb"))
		assertHasAttr(t, dp[0].Attributes, attribute.String("db.system.name", "mysql"))
	})

	t.Run("defaults merged with extra", func(t *testing.T) {
		reader := sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		meter := provider.Meter("test")

		rows, err := dbconv.NewClientResponseReturnedRows(meter)
		require.NoError(t, err)

		wrapped := NewInstrumentWithDBSystemAndDefaults(rows,
			attribute.String("db.namespace", "mydb"),
		)
		wrapped.Record(ctx, 100, dbconv.SystemNameMySQL,
			attribute.String("db.operation.name", "FIND"),
		)

		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))

		dp := findHistogramDataPoints[int64](t, rm, "db.client.response.returned_rows")
		require.Len(t, dp, 1)
		assertHasAttr(t, dp[0].Attributes, attribute.String("db.namespace", "mydb"))
		assertHasAttr(t, dp[0].Attributes, attribute.String("db.system.name", "mysql"))
		assertHasAttr(t, dp[0].Attributes, attribute.String("db.operation.name", "FIND"))
	})
}

func TestInstrumentWithDBSystemAndDefaults_NoDefaults(t *testing.T) {
	ctx := context.Background()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	meter := provider.Meter("test")

	dur, err := dbconv.NewClientOperationDuration(meter)
	require.NoError(t, err)

	wrapped := NewInstrumentWithDBSystemAndDefaults(dur)
	wrapped.Record(ctx, 0.3, dbconv.SystemNamePostgreSQL,
		attribute.String("db.operation.name", "INSERT"),
	)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))

	dp := findHistogramDataPoints[float64](t, rm, "db.client.operation.duration")
	require.Len(t, dp, 1)
	assertHasAttr(t, dp[0].Attributes, attribute.String("db.system.name", "postgresql"))
	assertHasAttr(t, dp[0].Attributes, attribute.String("db.operation.name", "INSERT"))
}

func TestInstrumentWithDBSystemAndDefaults_MultipleDefaults(t *testing.T) {
	ctx := context.Background()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	meter := provider.Meter("test")

	dur, err := dbconv.NewClientOperationDuration(meter)
	require.NoError(t, err)

	wrapped := NewInstrumentWithDBSystemAndDefaults(dur,
		attribute.String("db.namespace", "mydb"),
		attribute.String("server.address", "localhost"),
	)
	wrapped.Record(ctx, 0.8, dbconv.SystemNamePostgreSQL)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))

	dp := findHistogramDataPoints[float64](t, rm, "db.client.operation.duration")
	require.Len(t, dp, 1)
	assertHasAttr(t, dp[0].Attributes, attribute.String("db.namespace", "mydb"))
	assertHasAttr(t, dp[0].Attributes, attribute.String("server.address", "localhost"))
	assertHasAttr(t, dp[0].Attributes, attribute.String("db.system.name", "postgresql"))
}

func assertHasAttr(t *testing.T, set attribute.Set, expected attribute.KeyValue) {
	t.Helper()
	val, ok := set.Value(expected.Key)
	assert.True(t, ok, "attribute %q not found", expected.Key)
	assert.Equal(t, expected.Value, val, "attribute %q has wrong value", expected.Key)
}
