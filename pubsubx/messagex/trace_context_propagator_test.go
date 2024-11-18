package messagex

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestTraceContextPropagator(t *testing.T) {
	propagator := NewTraceContextPropagator()

	t.Run("Test Inject and Extract", func(t *testing.T) {
		// Create a new context with a trace ID and span ID
		traceID := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
		spanID := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
		ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    traceID,
			SpanID:     spanID,
			TraceFlags: trace.FlagsSampled,
		}))

		// Create a new message
		msg := &Message{
			Metadata: make(MessageMetadata),
		}

		// Inject the trace context into the message metadata
		propagator.Inject(ctx, msg)

		// Extract the trace context from the message metadata
		extractedCtx := propagator.Extract(context.Background(), msg.Metadata)
		extractedSpanContext := trace.SpanContextFromContext(extractedCtx)

		// Assert that the extracted trace ID and span ID match the original
		assert.Equal(t, traceID, extractedSpanContext.TraceID())
		assert.Equal(t, spanID, extractedSpanContext.SpanID())
	})

	t.Run("Extract with empty metadata", func(t *testing.T) {
		// Create an empty message metadata
		metadata := make(MessageMetadata)

		// Extract the trace context from the empty metadata
		extractedCtx := propagator.Extract(context.Background(), metadata)
		extractedSpanContext := trace.SpanContextFromContext(extractedCtx)

		// Assert that the extracted trace context is empty
		assert.False(t, extractedSpanContext.IsValid())
	})
}
