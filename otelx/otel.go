// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"github.com/clinia/x/logrusx"
)

type Otel struct {
	t *Tracer
	m *Meter
}

type OtelOptions struct {
	TracerConfig *TracerConfig
	MeterConfig  *MeterConfig
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
		m: &Meter{},
	}

	if otelOpts.TracerConfig != nil {
		if err := o.t.setup(l, otelOpts.TracerConfig); err != nil {
			return nil, err
		}
	} else {
		l.Infof("Tracer config is missing! - skipping tracer setup")
		o.t = NewNoopTracer("no-op-tracer")
	}

	if otelOpts.MeterConfig != nil {
		if err := o.m.setup(l, otelOpts.MeterConfig); err != nil {
			return nil, err
		}
	} else {
		l.Infof("Meter config is missing! - skipping meter setup")
		o.m = NewNoopMeter("no-op-meter")
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

func WithMeter(config *MeterConfig) OtelOption {
	return func(opts *OtelOptions) {
		if config != nil {
			opts.MeterConfig = config
		}
	}
}

// Creates a new no-op tracer and meter.
func NewNoop() *Otel {
	return &Otel{
		t: NewNoopTracer("no-op-tracer"),
		m: NewNoopMeter("no-op-meter"),
	}
}

// Tracer returns the underlying OpenTelemetry tracer.
func (o *Otel) Tracer() *Tracer {
	return o.t
}

// Meter returns the underlying OpenTelemetry meter.
func (o *Otel) Meter() *Meter {
	return o.m
}
