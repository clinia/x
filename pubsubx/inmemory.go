package pubsubx

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/clinia/x/logrusx"
)

type memorySubscriber struct {
	pubsubchan *gochannel.GoChannel
}

func SetupInMemorySubscriber(l *logrusx.Logger, c *Config) (Subscriber, error) {
	return &memorySubscriber{
		pubsubchan: gochannel.NewGoChannel(gochannel.Config{}, NewLogrusLogger(l.Logger)),
	}, nil
}

func (s *memorySubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return s.pubsubchan.Subscribe(ctx, topic)
}

func (s *memorySubscriber) Close() error {
	return s.pubsubchan.Close()
}

type memoryPublisher struct {
	pubsubchan *gochannel.GoChannel
}

func SetupInMemoryPublisher(l *logrusx.Logger, c *Config) (Publisher, error) {
	return &memoryPublisher{
		pubsubchan: gochannel.NewGoChannel(gochannel.Config{}, NewLogrusLogger(l.Logger)),
	}, nil
}

func (p *memoryPublisher) Publish(topic string, messages ...*message.Message) error {
	return p.pubsubchan.Publish(topic, messages...)
}

func (p *memoryPublisher) Close() error {
	return p.pubsubchan.Close()
}

type memoryPubSub struct {
	pubsubchan *gochannel.GoChannel
}

func SetupInMemoryPubSub(l *logrusx.Logger, c *Config) (*memoryPubSub, error) {
	return &memoryPubSub{
		pubsubchan: gochannel.NewGoChannel(gochannel.Config{}, NewLogrusLogger(l.Logger)),
	}, nil
}

func (ps *memoryPubSub) Publish(topic string, messages ...*message.Message) error {
	return ps.pubsubchan.Publish(topic, messages...)
}

func (ps *memoryPubSub) SetupSubscriber() func(group string) (Subscriber, error) {
	// The in-memory pubsub doesn't support grouping.
	return func(group string) (Subscriber, error) {
		return ps, nil
	}
}

func (ps *memoryPubSub) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return ps.pubsubchan.Subscribe(ctx, topic)
}

func (ps *memoryPubSub) Close() error {
	return ps.pubsubchan.Close()
}
