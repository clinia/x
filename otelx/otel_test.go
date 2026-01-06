package otelx

import (
	"testing"

	"github.com/clinia/x/loggerx"
	"github.com/stretchr/testify/assert"
)

func TestNewWithTracerConfig(t *testing.T) {
	tracerConfig := &TracerConfig{
		ServiceName: "Clinia X",
		Name:        "X",
		Provider:    "otel",
		Providers: TracerProvidersConfig{
			OTLP: OTLPTracerConfig{
				Protocol: "grpc",
				Insecure: true,
				Sampling: OTLPSampling{
					SamplingRatio: 1,
				},
			},
		},
	}

	otel, err := New(t.Context(), loggerx.NewDefaultLogger(), WithTracer(tracerConfig))

	assert.NoError(t, err)
	assert.NotNil(t, otel.Tracer())
	assert.NotNil(t, otel.Tracer().Provider())
	assert.NotNil(t, otel.Tracer().TextMapPropagator())
}

func TestNewWithMeterConfig(t *testing.T) {
	testCases := []struct {
		config *MeterConfig
	}{
		{
			config: &MeterConfig{
				ServiceName: "Clinia X",
				Name:        "X",
				Provider:    "otel",
				Providers: MeterProvidersConfig{
					OTLP: OTLPMeterConfig{
						Protocol: "grpc",
						Insecure: true,
					},
				},
			},
		},
		{
			config: &MeterConfig{
				ServiceName: "Clinia X",
				Name:        "X",
				Provider:    "prometheus",
				Providers:   MeterProvidersConfig{},
			},
		},
	}

	for _, tc := range testCases {
		otel, err := New(t.Context(), loggerx.NewDefaultLogger(), WithMeterConfig(tc.config))

		assert.NoError(t, err)
		assert.NotNil(t, otel.MeterProvider())
	}
}
