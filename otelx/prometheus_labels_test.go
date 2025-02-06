package otelx

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestNewPrometheusLabels(t *testing.T) {
	tests := []struct {
		name string
		kvs  []attribute.KeyValue
		want prometheus.Labels
	}{
		{
			name: "single attribute with dot",
			kvs:  []attribute.KeyValue{attribute.String("foo.bar", "baz")},
			want: prometheus.Labels{"foo__bar": "baz"},
		},
		{
			name: "multiple attributes with dots",
			kvs: []attribute.KeyValue{
				attribute.String("foo.bar", "baz"),
				attribute.String("baz.qux", "quux"),
			},
			want: prometheus.Labels{
				"foo__bar": "baz",
				"baz__qux": "quux",
			},
		},
		{
			name: "attribute without dot",
			kvs:  []attribute.KeyValue{attribute.String("foobar", "baz")},
			want: prometheus.Labels{"foobar": "baz"},
		},
		{
			name: "empty attributes",
			kvs:  []attribute.KeyValue{},
			want: prometheus.Labels{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPrometheusLabels(tt.kvs...)
			require.Equal(t, tt.want, got)
		})
	}
}
