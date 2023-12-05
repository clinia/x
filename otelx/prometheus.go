package otelx

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func SetupPrometheusMeter(meterName string, c *MeterConfig) (metric.Meter, error) {

	// The exporter embeds a default OpenTelemetry Reader and implements prometheus.Collector
	// TODO: Customize exporter. Meter reader
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}
	atts := append([]attribute.KeyValue{}, semconv.ServiceNameKey.String(c.ServiceName))
	atts = append(atts, c.ResourceAttributes...)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			atts...,
		)))

	return mp.Meter(meterName), nil
}
