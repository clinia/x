// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"context"
	"fmt"

	"github.com/clinia/x/loggerx"
	"github.com/clinia/x/stringsx"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

// NewMeterProvider creates a new meter provider based on the configuration.
func NewMeterProvider(ctx context.Context, l *loggerx.Logger, c *MeterConfig) (metric.MeterProvider, error) {
	switch f := stringsx.SwitchExact(c.Provider); {
	case f.AddCase("prometheus"):
		if meterProvider, err := SetupPrometheusMeterProvider(c.Name, c); err != nil {
			return nil, err
		} else {
			l.Info(ctx, "prometheus meter configured! Sending measurements to /metrics endpoint")
			return meterProvider, nil
		}

	case f.AddCase("otel"):
		if meterProvider, err := SetupOTLPMeterProvider(c.Name, c); err != nil {
			return nil, err
		} else {
			l.Info(ctx, fmt.Sprintf("OTLP meter configured! Sending measurements to %s", c.Providers.OTLP.ServerURL))
			return meterProvider, nil
		}

	case f.AddCase("stdout"):
		if meterProvider, err := SetupStdoutMeterProvider(c.Name, c); err != nil {
			return nil, err
		} else {
			l.Info(ctx, "stdout meter configured! Sending measurements to stdout")
			return meterProvider, nil
		}

	case f.AddCase(""):
		l.Info(ctx, "missing provider in config - skipping meter setup")

		return noop.NewMeterProvider(), nil
	default:
		return nil, f.ToUnknownCaseErr()
	}
}
