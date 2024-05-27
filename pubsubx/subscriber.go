package pubsubx

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Subscriber interface {
	// Subscribe subscribes to the topic.
	Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
	// Close closes the subscriber.
	Close() error
}

type BatchConsumerOptions struct {
	// MaxBatchSize max amount of elements the batch will contain.
	// Default value is 100 if nothing is specified.
	MaxBatchSize int16
	// MaxWaitTime max time that it will be waited until MaxBatchSize elements are received.
	// Default value is 100ms if nothing is specified.
	MaxWaitTime time.Duration
}

type subscriberOptions struct {
	consumerModel        ConsumerModel
	batchConsumerOptions *BatchConsumerOptions
	nackResendSleep      *time.Duration
}

type SubscriberOption func(*subscriberOptions)

// WithNackResendSleep sets the sleep time between resending NACKed messages.
// If not set, the default value is 100ms.
// You can also set it to kafkax.NoSleep to disable the sleep.
func WithNackResendSleep(nackResendSleep time.Duration) SubscriberOption {
	return func(o *subscriberOptions) {
		o.nackResendSleep = &nackResendSleep
	}
}

func WithDefaultConsumerModel() SubscriberOption {
	return func(o *subscriberOptions) {
		o.consumerModel = ConsumerModelDefault
		o.batchConsumerOptions = nil
	}
}

func WithBatchConsumerModel(batchOptions *BatchConsumerOptions) SubscriberOption {
	return func(o *subscriberOptions) {
		o.consumerModel = ConsumerModelBatch
		o.batchConsumerOptions = batchOptions
	}
}
