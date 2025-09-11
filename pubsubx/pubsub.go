package pubsubx

import (
	"github.com/clinia/x/pubsubx/messagex"
)

// PubSub represents a generic interface for a publish-subscribe system.
type PubSub interface {
	// Bootstrap the pubsub instance
	Bootstrap() error

	// Publisher returns a Publisher instance for publishing messages.
	Publisher() Publisher

	// Subscriber returns a Subscriber instance for subscribing to messages.
	// It takes a group name, a list of topics, and optional SubscriberOptions.
	// It returns a Subscriber and an error if any.
	// A subscriber should define ALL the topics it wants to subscribe to.
	Subscriber(group string, topics []messagex.Topic, opts ...SubscriberOption) (Subscriber, error)

	// PubSubAdminClient returns a PubSubAdminClient instance for managing topics and configurations.
	// A new client is always returned on each call. Caller is responsible for closing the client.
	AdminClient() (PubSubAdminClient, error)

	// Close closes all publishers and subscribers associated with the PubSub instance.
	// It returns an error if any.
	Close() error
}
