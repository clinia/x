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
			OTLP: OTLPTracerConfig{
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
		otel, err := New(logrusx.New("clinia/x", "1"), WithMeter(tc.config))

		assert.NoError(t, err)
		assert.NotNil(t, otel.Meter())
		assert.NotNil(t, otel.Meter().Provider())
	}
}
