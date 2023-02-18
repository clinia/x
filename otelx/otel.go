// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"go.opentelemetry.io/otel/trace"

	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/stringsx"
)

type Tracer struct {
	tracer trace.Tracer
}

// Creates a new tracer. If name is empty, a default tracer name is used
// instead. See: https://godocs.io/go.opentelemetry.io/otel/sdk/trace#TracerProvider.Tracer
func New(name string, l *logrusx.Logger, c *Config) (*Tracer, error) {
	t := &Tracer{}

	if err := t.setup(name, l, c); err != nil {
		return nil, err
	}

	return t, nil
}

// Creates a new no-op tracer.
func NewNoop(_ *logrusx.Logger, c *Config) *Tracer {
	tp := trace.NewNoopTracerProvider()
	t := &Tracer{tracer: tp.Tracer("")}
	return t
}

// setup constructs the tracer based on the given configuration.
func (t *Tracer) setup(name string, l *logrusx.Logger, c *Config) error {
	switch f := stringsx.SwitchExact(c.Provider); {
	case f.AddCase("jaeger"):
		tracer, err := SetupJaeger(t, name, c)
		if err != nil {
			return err
		}

		t.tracer = tracer
		l.Infof("Jaeger tracer configured! Sending spans to %s", c.Providers.Jaeger.LocalAgentAddress)
	case f.AddCase("otel"):
		tracer, err := SetupOTLP(t, name, c)
		if err != nil {
			return err
		}

		t.tracer = tracer
		l.Infof("OTLP tracer configured! Sending spans to %s", c.Providers.OTLP.ServerURL)
	case f.AddCase("stdout"):
		tracer, err := SetupStdout(t, name, c)
		if err != nil {
			return err
		}

		t.tracer = tracer
		l.Infof("Stdout tracer configured! Sending spans to stdout")
	case f.AddCase(""):
		l.Infof("No tracer configured - skipping tracing setup")
		t.tracer = trace.NewNoopTracerProvider().Tracer(name)
	default:
		return f.ToUnknownCaseErr()
	}

	return nil
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

// WithOTLP returns a new tracer with the underlying OpenTelemetry Tracer
// replaced.
func (t *Tracer) WithOTLP(other trace.Tracer) *Tracer {
	return &Tracer{other}
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
