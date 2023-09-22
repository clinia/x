// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/stringsx"
)

type Tracer struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

// setup constructs the tracer based on the given configuration.
func (t *Tracer) setup(l *logrusx.Logger, c *TracerConfig) error {
	switch f := stringsx.SwitchExact(c.Provider); {
	case f.AddCase("jaeger"):
		tracer, prop, err := SetupJaegerTracer(c.Name, c)
		if err != nil {
			return err
		}

		t.tracer = tracer
		t.propagator = prop
		l.Infof("Jaeger tracer configured! Sending spans to %s", c.Providers.Jaeger.LocalAgentAddress)
	case f.AddCase("otel"):
		tracer, prop, err := SetupOTLPTracer(c.Name, c)
		if err != nil {
			return err
		}

		t.tracer = tracer
		t.propagator = prop
		l.Infof("OTLP tracer configured! Sending spans to %s", c.Providers.OTLP.ServerURL)
	case f.AddCase("stdout"):
		tracer, prop, err := SetupStdoutTracer(c.Name, c)
		if err != nil {
			return err
		}

		t.tracer = tracer
		t.propagator = prop
		l.Infof("Stdout tracer configured! Sending spans to stdout")
	case f.AddCase(""):
		l.Infof("Missing provider in config - skipping tracing setup")
		t.tracer = trace.NewNoopTracerProvider().Tracer("NoopTracer")
		t.propagator = propagation.NewCompositeTextMapPropagator()
	default:
		return f.ToUnknownCaseErr()
	}
	return nil
}

func NewNoopTracer() *Tracer {
	return &Tracer{
		tracer:     trace.NewNoopTracerProvider().Tracer("NoopTracer"),
		propagator: propagation.NewCompositeTextMapPropagator(),
	}
}

// IsLoaded returns true if the tracer has been loaded.
func (t *Tracer) IsLoaded() bool {
	if t == nil || t.tracer == nil {
		return false
	}
	return true
}

// Tracer returns the underlying OpenTelemetry tracer.
func (t *Tracer) Tracer() trace.Tracer {
	return t.tracer
}

// Provider returns a TracerProvider which in turn yieds this tracer unmodified.
func (t *Tracer) Provider() trace.TracerProvider {
	return tracerProvider{t.Tracer()}
}

type tracerProvider struct {
	t trace.Tracer
}

var _ trace.TracerProvider = tracerProvider{}

// Tracer implements trace.TracerProvider.
func (tp tracerProvider) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	return tp.t
}

// TextMapPropagator returns the underlying OpenTelemetry textMapPropagator.
func (t *Tracer) TextMapPropagator() propagation.TextMapPropagator {
	return t.propagator
}

// Inject set tracecontext from the Context into the carrier.
func (t *Tracer) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	t.propagator.Inject(ctx, carrier)
}

// Extract reads tracecontext from the carrier into a returned Context.
// If the extracted tracecontext is invalid, the passed ctx will be returned directly instead.
func (t *Tracer) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return t.propagator.Extract(ctx, carrier)
}
