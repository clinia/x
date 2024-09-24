package pubsubx

import (
	"context"

	"github.com/clinia/x/pubsubx/messagex"
)

// Publisher is an interface for publishing messages to a topic.
type Publisher interface {
	// PublishSync publishes messages synchronously to the specified topic.
	// It returns an error if the operation fails.
	PublishSync(ctx context.Context, topic messagex.Topic, messages ...*messagex.Message) (Errors, error)

	// PublishAsync publishes messages asynchronously to the specified topic.
	// It returns an error if the operation fails.
	//
	// WARNING: The context should stay alive until the messages are published.
	// When using a fire-n-forget approach in a request-response scenario, a new Context.Background() should be preferred since
	// the request context might be canceled before the messages are actually published.
	PublishAsync(ctx context.Context, topic messagex.Topic, messages ...*messagex.Message) error

	// Close closes the publisher.
	// Once a publisher is closed, it cannot be used to publish messages anymore.
	//
	// WARNING: Since the PubSub.Publisher() method always returns the same instance of the publisher,
	// closing the publisher will force the user to create a new PubSub instance to get a new publisher.
	// This should really be done only when the application is shutting down, or you no longer have any needs to publish messages.
	Close() error
}
