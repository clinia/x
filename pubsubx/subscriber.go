package pubsubx

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Subscriber interface {
	// Subscribe subscribes to the topic.
	Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
	// Close closes the subscriber.
	Close() error
}
