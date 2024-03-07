package otelsaramax

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

// We need a fake tracer provider to ensure the one passed in options is the one used afterwards.
// In order to avoid adding the SDK as a dependency, we use this mock.
type fakeTracerProvider struct {
	embedded.TracerProvider
}

func (fakeTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return fakeTracer{
		name: name,
	}
}

type fakeTracer struct {
	embedded.Tracer
	name string
}

func (fakeTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return ctx, nil
}

func TestNewConfig(t *testing.T) {
	tp := fakeTracerProvider{}
	prop := propagation.NewCompositeTextMapPropagator()

	testCases := []struct {
		name     string
		opts     []Option
		expected config
	}{
		{
			name: "with provider",
			opts: []Option{
				WithTracerProvider(tp),
			},
			expected: config{
				TracerProvider: tp,
				Tracer:         tp.Tracer(defaultTracerName, trace.WithInstrumentationVersion(Version())),
				Propagators:    otel.GetTextMapPropagator(),
			},
		},
		{
			name: "with empty provider",
			opts: []Option{
				WithTracerProvider(nil),
			},
			expected: config{
				TracerProvider: otel.GetTracerProvider(),
				Tracer:         otel.GetTracerProvider().Tracer(defaultTracerName, trace.WithInstrumentationVersion(Version())),
				Propagators:    otel.GetTextMapPropagator(),
			},
		},
		{
			name: "with propagators",
			opts: []Option{
				WithPropagators(prop),
			},
			expected: config{
				TracerProvider: otel.GetTracerProvider(),
				Tracer:         otel.GetTracerProvider().Tracer(defaultTracerName, trace.WithInstrumentationVersion(Version())),
				Propagators:    prop,
			},
		},
		{
			name: "with empty propagators",
			opts: []Option{
				WithPropagators(nil),
			},
			expected: config{
				TracerProvider: otel.GetTracerProvider(),
				Tracer:         otel.GetTracerProvider().Tracer(defaultTracerName, trace.WithInstrumentationVersion(Version())),
				Propagators:    otel.GetTextMapPropagator(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := newConfig(tc.opts...)
			assert.Equal(t, tc.expected, result)
		})
	}
}
