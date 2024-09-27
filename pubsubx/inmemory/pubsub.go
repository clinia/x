package inmemorypubsub

import (
	"context"
	"sync"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/samber/lo"
)

type (
	memoryPubSub struct {
		scope       string
		mu          sync.RWMutex
		subscribers []*memorySubscriber
		admClient   pubsubx.PubSubAdminClient
	}
	memorySubscriber struct {
		m             *memoryPubSub
		topics        []messagex.Topic
		topicHandlers pubsubx.Handlers
		closeFn       func() error
	}
)

var (
	_ pubsubx.PubSub     = (*memoryPubSub)(nil)
	_ pubsubx.Publisher  = (*memoryPubSub)(nil)
	_ pubsubx.Subscriber = (*memorySubscriber)(nil)
)

func SetupInMemoryPubSub(l *logrusx.Logger, c *pubsubx.Config) (*memoryPubSub, error) {
	return &memoryPubSub{
		scope:       c.Scope,
		subscribers: make([]*memorySubscriber, 0),
	}, nil
}

// Close implements Publisher.
func (m *memoryPubSub) Close() error {
	return nil
}

// PublishAsync implements Publisher.
func (m *memoryPubSub) PublishAsync(ctx context.Context, topic messagex.Topic, messages ...*messagex.Message) error {
	go m.PublishSync(ctx, topic, messages...)
	return nil
}

// PublishSync implements Publisher.
func (m *memoryPubSub) PublishSync(ctx context.Context, topic messagex.Topic, messages ...*messagex.Message) (pubsubx.Errors, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sub := range m.subscribers {
		if handler, ok := sub.topicHandlers[topic]; !ok || handler == nil {
			continue
		} else {
			handler(ctx, messages)
		}
	}

	errs := make(pubsubx.Errors, len(messages))
	return errs, nil
}

// Publisher implements PubSub.
func (m *memoryPubSub) Publisher() pubsubx.Publisher {
	return m
}

// Subscriber implements PubSub.
func (m *memoryPubSub) Subscriber(group string, topics []messagex.Topic, opts ...pubsubx.SubscriberOption) (pubsubx.Subscriber, error) {
	var out *memorySubscriber
	out = &memorySubscriber{
		m:      m,
		topics: topics,
		closeFn: func() error {
			m.mu.Lock()
			defer m.mu.Unlock()
			_, idx, ok := lo.FindIndexOf(m.subscribers, func(s *memorySubscriber) bool {
				return s == out
			})
			if !ok {
				return errorx.NotFoundErrorf("subscriber not found")
			}

			m.subscribers = append(m.subscribers[:idx], m.subscribers[idx+1:]...)

			return nil
		},
	}

	m.mu.Lock()
	m.subscribers = append(m.subscribers, out)
	m.mu.Unlock()

	return out, nil
}

// Close implements Subscriber.
func (m *memorySubscriber) Close() error {
	return m.closeFn()
}

// Subscribe implements Subscriber.
func (m *memorySubscriber) Subscribe(ctx context.Context, topicHandlers pubsubx.Handlers) error {
	m.m.mu.Lock()
	defer m.m.mu.Unlock()
	for _, topic := range m.topics {
		handler, exists := topicHandlers[topic]
		if !exists {
			return errorx.InternalErrorf("missing handler for topic %s", topic)
		} else if handler == nil {
			return errorx.InternalErrorf("nil handler for topic %s", topic)
		}
	}

	m.topicHandlers = topicHandlers

	return nil
}

// AdminClient implements PubSub.
func (m *memoryPubSub) AdminClient() pubsubx.PubSubAdminClient {
	if m.admClient == nil {
		m.admClient = NewNoopAdminClient()
	}

	return m.admClient
}
