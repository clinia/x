// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

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

// Creates a new tracer. If name is empty, a default tracer name is used
// instead. See: https://godocs.io/go.opentelemetry.io/otel/sdk/trace#TracerProvider.Tracer
func New(l *logrusx.Logger, opts OtelOptions) (*Otel, error) {
	t := &Tracer{}
	if opts.TracerConfig != nil {
		if err := t.setup(l, opts.TracerConfig); err != nil {
			return nil, err
		}
	} else {
		t.tracer = trace.NewNoopTracerProvider().Tracer("NoopTracer")
		t.propagator = propagation.NewCompositeTextMapPropagator()
	}

	o := &Otel{
		t: t,
	}
	return o, nil
}

// Creates a new no-op tracer and meter.
func NewNoop() *Otel {
	t := &Tracer{tracer: trace.NewNoopTracerProvider().Tracer("NoopTracer")}

	// mp := sdk.NewMeterProvider()
	// m := &Meter{meter: mp.Meter("")}

	return &Otel{
		t: t,
		// Metric: m,
	}
}
