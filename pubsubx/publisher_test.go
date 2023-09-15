package pubsubx

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type MockPublisher struct {
	PublishFunc func(topic string, messages ...*message.Message) error
}

func (m *MockPublisher) Publish(topic string, messages ...*message.Message) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(topic, messages...)
	}
	return nil
}

func (m *MockPublisher) Close() error {
	return nil
}
func TestOtelPublishPropagator(t *testing.T) {
	msg := message.NewMessage("test", []byte("payload"))
	traceID := trace.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	spanID := trace.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	sc := trace.NewSpanContext(
		trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		},
	)

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
	)

	mockPublisher := &MockPublisher{
		// Function to assess if message contains span context in metadata
		PublishFunc: func(topic string, messages ...*message.Message) error {
			ctx := prop.Extract(context.Background(), propagation.MapCarrier(messages[0].Metadata))
			sc := trace.SpanContextFromContext(ctx)
			assert.Equal(t, sc.TraceID(), traceID)
			assert.Equal(t, sc.SpanID(), spanID)
			return nil
		},
	}
	otelPublisher := NewOTelPublisher(mockPublisher, prop)
	otelPublisher.Publish(trace.ContextWithSpanContext(context.Background(), sc), "test", msg)
}
