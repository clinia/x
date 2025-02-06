package otelx

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
)

// NewPrometheusLabels converts a variadic list of attribute.KeyValue pairs into a prometheus.Labels map.
// Each attribute.KeyValue's key is transformed by replacing all dots ('.') with underscores ('_').
// The resulting map can be used with prometheus for structured metrics.
//
// Note: This transformation is necessary since prometheus labels cannot contain dots. This also enables us to reuse the same attributes for both tracing and metrics.
func NewPrometheusLabels(kvs ...attribute.KeyValue) prometheus.Labels {
	labels := prometheus.Labels{}
	for _, kv := range kvs {
		k := strings.ReplaceAll(string(kv.Key), ".", "_")
		labels[k] = kv.Value.AsString()
	}

	return labels
}
