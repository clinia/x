package otelx

import (
	"testing"

	"github.com/clinia/x/logrusx"
	"github.com/stretchr/testify/assert"
)

func TestNewWithTracerConfig(t *testing.T) {
	tracerConfig := &TracerConfig{
		ServiceName: "Clinia X",
		Name:        "X",
		Provider:    "otel",
		Providers: TracerProvidersConfig{
			OTLP: OTLPConfig{
				Protocol: "grpc",
				Insecure: true,
				Sampling: OTLPSampling{
					SamplingRatio: 1,
				},
			},
		},
	}

	otel, err := New(logrusx.New("clinia/x", "1"), WithTracer(tracerConfig))

	assert.NoError(t, err)
	assert.NotNil(t, otel.Tracer())
	assert.NotNil(t, otel.Tracer().Provider())
	assert.NotNil(t, otel.Tracer().TextMapPropagator())
}
