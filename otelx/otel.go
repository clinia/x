// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"github.com/clinia/x/logrusx"
)

type Otel struct {
	t *Tracer
	// m *Meter
}

type OtelOptions struct {
	TracerConfig *TracerConfig
	// MeterConfig  *MeterConfig
}

type OtelOption func(*OtelOptions)

// Creates opentelemetry tools (meter and tracer). Leaving
func New(l *logrusx.Logger, opts ...OtelOption) (*Otel, error) {
	otelOpts := &OtelOptions{}
	for _, opt := range opts {
		opt(otelOpts)
	}

	o := &Otel{
		t: &Tracer{},
	}
	if otelOpts.TracerConfig != nil {
		if err := o.t.setup(l, otelOpts.TracerConfig); err != nil {
			return nil, err
		}
	} else {
		l.Infof("Tracing config is missing! - skipping tracing setup")
		o.t = NewNoopTracer()
	}

	return o, nil
}

func WithTracer(config *TracerConfig) OtelOption {
	return func(opts *OtelOptions) {
		if config != nil {
			opts.TracerConfig = config
		}
	}
}

// Creates a new no-op tracer and meter.
func NewNoop() *Otel {
	t := NewNoopTracer()

	// mp := sdk.NewMeterProvider()
	// m := &Meter{meter: mp.Meter("")}

	return &Otel{
		t: t,
		// Metric: m,
	}
}

// Tracer returns the underlying OpenTelemetry tracer.
func (o *Otel) Tracer() *Tracer {
	return o.t
}
