package messagex

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

// TraceContextPropagator is responsible for injecting and extracting trace context
// into and from message metadata.
type TraceContextPropagator struct {
	prop propagation.TextMapPropagator
}

func NewTraceContextPropagator() TraceContextPropagator {
	return TraceContextPropagator{
		prop: propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	}
}

// Inject injects the trace context from the provided context into the message metadata.
func (t *TraceContextPropagator) Inject(ctx context.Context, m *Message) {
	t.prop.Inject(ctx, propagation.MapCarrier(m.Metadata))
}

// Extract extracts the trace context from the message metadata and returns a new context
// with the extracted trace context.
func (t *TraceContextPropagator) Extract(ctx context.Context, metadata MessageMetadata) context.Context {
	return t.prop.Extract(ctx, propagation.MapCarrier(metadata))
}
