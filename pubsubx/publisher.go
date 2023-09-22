package pubsubx

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

// publisher is the interface that wraps the Publish method.
type Publisher interface {
	// Publish publishes a message to the topic.
	Publish(ctx context.Context, topic string, messages ...*message.Message) error
	// Close closes the publisher.
	Close() error
}
