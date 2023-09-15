package pubsubx

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel/propagation"
)

// publisher is the interface that wraps the Publish method.
type Publisher interface {
	// Publish publishes a message to the topic.
	Publish(topic string, messages ...*message.Message) error
	// Close closes the publisher.
	Close() error
}

type OTelPublisher struct {
	publisher  Publisher
	propagator propagation.TextMapPropagator
}

// NewOTelPublisher creates a new OTelPublisher with the given publisher and OpenTelemetry propagator.
func NewOTelPublisher(pub Publisher, prop propagation.TextMapPropagator) *OTelPublisher {
	return &OTelPublisher{
		publisher:  pub,
		propagator: prop,
	}
}

func (p *OTelPublisher) Publish(ctx context.Context, topic string, messages ...*message.Message) error {
	if p.propagator != nil {
		for _, msg := range messages {
			p.propagator.Inject(ctx, propagation.MapCarrier(msg.Metadata))
		}
	}
	return p.publisher.Publish(topic, messages...)

}

func (p *OTelPublisher) Close() error {
	return p.publisher.Close()
}
