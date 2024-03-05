package pubsubx

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/clinia/x/logrusx"
)

type memoryPubSub struct {
	scope      string
	pubsubchan *gochannel.GoChannel
}

var _ Publisher = (*memoryPubSub)(nil)

func SetupInMemoryPubSub(l *logrusx.Logger, c *Config) (*memoryPubSub, error) {
	return &memoryPubSub{
		scope:      c.Scope,
		pubsubchan: gochannel.NewGoChannel(gochannel.Config{}, NewLogrusLogger(l.Logger)),
	}, nil
}

func (ps *memoryPubSub) Publish(ctx context.Context, topic string, messages ...*message.Message) error {
	return ps.pubsubchan.Publish(topicName(ps.scope, topic), messages...)
}

// BulkPublish implements Publisher.
func (ps *memoryPubSub) BulkPublish(ctx context.Context, topic string, messages ...*message.Message) error {
	return ps.pubsubchan.Publish(topicName(ps.scope, topic), messages...)
}

func (ps *memoryPubSub) setupSubscriber() func(group string, subOpts *subscriberOptions) (Subscriber, error) {
	// The in-memory pubsub doesn't support grouping.
	return func(group string, subOpts *subscriberOptions) (Subscriber, error) {
		// TODO - Support batch consumer model
		return ps, nil
	}
}

func (ps *memoryPubSub) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return ps.pubsubchan.Subscribe(ctx, topicName(ps.scope, topic))
}

func (ps *memoryPubSub) Close() error {
	return ps.pubsubchan.Close()
}
