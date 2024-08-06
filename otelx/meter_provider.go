// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/stringsx"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

// NewMeterProvider creates a new meter provider based on the configuration.
func NewMeterProvider(l *logrusx.Logger, c *MeterConfig) (metric.MeterProvider, error) {
	switch f := stringsx.SwitchExact(c.Provider); {
	case f.AddCase("prometheus"):
		if meterProvider, err := SetupPrometheusMeterProvider(c.Name, c); err != nil {
			return nil, err
		} else {
			l.Infof("Prometheus meter configured! Sending measurements to /metrics endpoint")
			return meterProvider, nil
		}

	case f.AddCase("otel"):
		if meterProvider, err := SetupOTLPMeterProvider(c.Name, c); err != nil {
			return nil, err
		} else {
			l.Infof("OTLP meter configured! Sending measurements to %s", c.Providers.OTLP.ServerURL)
			return meterProvider, nil
		}

	case f.AddCase("stdout"):
		if meterProvider, err := SetupStdoutMeterProvider(c.Name, c); err != nil {
			return nil, err
		} else {
			l.Infof("Stdout meter configured! Sending measurements to stdout")
			return meterProvider, nil
		}

	case f.AddCase(""):
		l.Infof("Missing provider in config - skipping meter setup")

		return noop.NewMeterProvider(), nil
	default:
		return nil, f.ToUnknownCaseErr()
	}
}
