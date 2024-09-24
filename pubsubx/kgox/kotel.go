package kgox

import (
	"github.com/twmb/franz-go/plugin/kotel"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func newKotel(tracerProvider trace.TracerProvider, propagator propagation.TextMapPropagator, meterProvider metric.MeterProvider) *kotel.Kotel {
	// Create a new kotel tracer with the provided tracer provider and
	// propagator.
	tracerOpts := []kotel.TracerOpt{
		kotel.TracerProvider(tracerProvider),
		kotel.TracerPropagator(propagator),
	}
	tr := kotel.NewTracer(tracerOpts...)

	// Create a new kotel meter with the provided meter provider.
	meterOpts := []kotel.MeterOpt{
		kotel.MeterProvider(meterProvider),
	}
	m := kotel.NewMeter(meterOpts...)

	return kotel.NewKotel(kotel.WithTracer(tr), kotel.WithMeter(m))
}
