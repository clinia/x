package otelx

import (
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
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
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(
			c.Providers.OTLP.Sampling.SamplingRatio,
		))),
	}

	tp := sdktrace.NewTracerProvider(tpOpts...)

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	return tp.Tracer(tracerName), prop, nil
}
