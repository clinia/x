package pubsubx

import "github.com/ThreeDotsLabs/watermill/message"

// Publisher is the interface that wraps the Publish method.
type Publisher interface {
	// Publish publishes a message to the topic.
	Publish(topic string, messages ...*message.Message) error
	// Close closes the publisher.
	Close() error
}
