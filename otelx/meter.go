// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/stringsx"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
	"go.opentelemetry.io/otel/metric/noop"
)

type Meter struct {
	meter metric.Meter
}

// Setup constructs the meter based on the given configuration.
func (m *Meter) setup(l *logrusx.Logger, c *MeterConfig) error {
	switch f := stringsx.SwitchExact(c.Provider); {
	case f.AddCase("otel"):
		meter, err := SetupOTLPMeter(c.Name, c)
		if err != nil {
			return err
		}

		m.meter = meter
		l.Infof("OTLP meter configured! Sending measurements to %s", c.Providers.OTLP.ServerURL)
	default:
		return f.ToUnknownCaseErr()
	}
	return nil
}

func NewNoopMeter() *Meter {
	return &Meter{
		meter: noop.NewMeterProvider().Meter("NoopMeter"),
	}
}

// IsLoaded returns true if the meter has been loaded.
func (m *Meter) IsLoaded() bool {
	if m == nil || m.meter == nil {
		return false
	}
	return true
}

// Meter returns the underlying OpenTelemetry meter.
func (m *Meter) Meter() metric.Meter {
	return m.meter
}

// Provider returns a MeterProvder which in turn yieds this meter unmodified.
func (m *Meter) Provider() metric.MeterProvider {
	return meterProvider{m: m.Meter()}
}

type meterProvider struct {
	embedded.MeterProvider
	m metric.Meter
}

var _ metric.MeterProvider = meterProvider{}

// Meter implements metric.MeterProvder.
func (mp meterProvider) Meter(name string, options ...metric.MeterOption) metric.Meter {
	return mp.m
}
