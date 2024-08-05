// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"github.com/clinia/x/logrusx"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

type Otel struct {
	t    *Tracer
	mp   metric.MeterProvider
	opts *OtelOptions
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
		t:    &Tracer{},
		opts: otelOpts,
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
		if mp, err := NewMeterProvider(l, otelOpts.MeterConfig); err != nil {
			return nil, err
		} else {
			o.mp = mp
		}
	} else {
		l.Infof("Meter config is missing! - skipping meter setup")
		o.mp = noop.NewMeterProvider()
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

func WithMeterConfig(config *MeterConfig) OtelOption {
	return func(opts *OtelOptions) {
		if config != nil {
			opts.MeterConfig = config
		}
	}
}

// Creates a new no-op tracer and meter.
func NewNoop() *Otel {
	return &Otel{
		t:  NewNoopTracer("no-op-tracer"),
		mp: noop.NewMeterProvider(),
	}
}

// Tracer returns the underlying OpenTelemetry tracer.
func (o *Otel) Tracer() *Tracer {
	return o.t
}

// MeterProvider returns the underlying OpenTelemetry meter provider.
func (o *Otel) MeterProvider() metric.MeterProvider {
	return o.mp
}

func (o *Otel) Meter() metric.Meter {
	return o.MeterProvider().Meter(o.opts.MeterConfig.Name, metric.WithInstrumentationAttributes(o.opts.MeterConfig.ResourceAttributes...))
}
