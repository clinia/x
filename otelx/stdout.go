package otelx

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

func SetupStdoutTracer(tracerName string, c *TracerConfig) (trace.Tracer, propagation.TextMapPropagator, error) {
	opts := []stdouttrace.Option{}

	if c.Providers.Stdout.Pretty {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	}

	exp, err := stdouttrace.New(opts...)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(c.ServiceName),
		)),
	}

	if c.SpanLimits != nil {
		tpOpts = append(tpOpts, sdktrace.WithRawSpanLimits(*c.SpanLimits))
	}

	tp := sdktrace.NewTracerProvider(tpOpts...)

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	return tp.Tracer(tracerName), prop, nil
}

func SetupStdoutMeterProvider(meterName string, c *MeterConfig) (metric.MeterProvider, error) {
	// Print with a JSON encoder that indents with two spaces.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	exp, err := stdoutmetric.New(
		stdoutmetric.WithEncoder(enc),
		stdoutmetric.WithoutTimestamps(),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	reader := sdkmetric.NewPeriodicReader(exp)

	mOpts := []sdkmetric.Option{
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(c.ServiceName),
		)),
	}

	return sdkmetric.NewMeterProvider(mOpts...), nil
}
