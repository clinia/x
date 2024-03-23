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
var _ Subscriber = (*memoryPubSub)(nil)
var _ Admin = (*memoryPubSub)(nil)

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

// CreateTopic implements Admin.
func (ps *memoryPubSub) CreateTopic(ctx context.Context, topic string, detail *TopicDetail) error {
	// Admin is not a concept in the in-memory pubsub.
	return nil
}

// DeleteSubscriber implements Admin.
func (ps *memoryPubSub) DeleteSubscriber(ctx context.Context, subscriber string) error {
	// Admin is not a concept in the in-memory pubsub.
	return nil
}

// DeleteTopic implements Admin.
func (ps *memoryPubSub) DeleteTopic(ctx context.Context, topic string) error {
	// Admin is not a concept in the in-memory pubsub.
	return nil
}

// ListSubscribers implements Admin.
func (ps *memoryPubSub) ListSubscribers(ctx context.Context, topic string) (map[string]string, error) {
	// Admin is not a concept in the in-memory pubsub.
	return map[string]string{}, nil
}

// ListTopics implements Admin.
func (ps *memoryPubSub) ListTopics(ctx context.Context) (map[string]TopicDetail, error) {
	// Admin is not a concept in the in-memory pubsub.
	return map[string]TopicDetail{}, nil
}
