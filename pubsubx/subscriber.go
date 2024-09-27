package pubsubx

import (
	"context"

	"github.com/clinia/x/pubsubx/messagex"
)

type (
	// Handler represents a function that handles messages received by a subscriber.
	// It takes a context.Context and a slice of *messagex.Message as input parameters.
	// The function should return a slice of errors, representing per-message failures,
	// and an error, representing the processing failure in general.
	//
	// For the singular error, the caller may return a pubsubx.AbortSubscribeError() to abort the subscription right away.
	// For any other errors happening, the handler will be retried up to a certain number of times. After that, the subscription will be aborted.
	// As such, it is the caller's responsibility to return an error that is related to the batch processing itself and not to a specific message.
	Handler  func(ctx context.Context, msgs []*messagex.Message) ([]error, error)
	Handlers map[messagex.Topic]Handler
)

type Subscriber interface {
	// Subscribe subscribes to all topics that are configured in the subscriber.
	// It takes a context and a map of topic handlers as input.
	// - If there are topics missing handlers, it will return an error immediately.
	Subscribe(ctx context.Context, topicHandlers Handlers) error
	// Health returns the health status of the subscriber.
	// It should return an error if the subscriber is unhealthy, nil otherwise (healthy).
	Health() error
	// Close closes the subscriber.
	Close() error
}
