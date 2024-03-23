package pubsubx

import (
	"fmt"
	"sync"

	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/stringsx"
)

type PubSub interface {
	Admin() Admin

	Publisher() Publisher

	Subscriber(group string, opts ...SubscriberOption) (Subscriber, error)

	// CLoses all publishers and subscribers.
	Close() error
}

type pubSub struct {
	admin      Admin
	publisher  Publisher
	subscriber func(group string, opts *subscriberOptions) (Subscriber, error)
	subs       sync.Map
}

var _ PubSub = (*pubSub)(nil)

func New(l *logrusx.Logger, c *Config, opts ...PubSubOption) (PubSub, error) {
	pubSubOpts := &pubSubOptions{}
	for _, opt := range opts {
		opt(pubSubOpts)
	}

	ps := &pubSub{
		subs: sync.Map{},
	}

	if err := ps.setup(l, c, pubSubOpts); err != nil {
		return nil, err
	}

	return ps, nil
}

func (ps *pubSub) setup(l *logrusx.Logger, c *Config, opts *pubSubOptions) error {
	switch f := stringsx.SwitchExact(c.Provider); {
	case f.AddCase("kafka"):
		publisher, err := setupKafkaPublisher(l, c, opts)
		if err != nil {
			return err
		}

		admin, err := setupKafkaAdmin(l, c)
		if err != nil {
			return err
		}
		ps.admin = admin

		ps.publisher = publisher
		ps.subscriber = func(group string, subOpts *subscriberOptions) (Subscriber, error) {
			if ms, ok := ps.subs.Load(group); !ok {
				s, e := setupKafkaSubscriber(l, c, opts, group, subOpts)
				if e != nil {
					return nil, e
				}

				ps.subs.Store(group, s)
				return s, nil
			} else {
				return ms.(Subscriber), nil
			}
		}
		l.Infof("Kafka pubsub configured! Sending & receiving messages to %s", c.Providers.Kafka.Brokers)

	case f.AddCase("inmemory"):
		pubsub, err := SetupInMemoryPubSub(l, c)
		if err != nil {
			return err
		}

		ps.admin = pubsub
		ps.publisher = pubsub
		ps.subscriber = pubsub.setupSubscriber()
		l.Infof("InMemory publisher configured! Sending & receiving messages to in-memory")
	default:
		return f.ToUnknownCaseErr()
	}

	return nil
}

func (ps *pubSub) Admin() Admin {
	return ps.admin
}

func (ps *pubSub) Publisher() Publisher {
	return ps.publisher
}

func (ps *pubSub) Subscriber(group string, opts ...SubscriberOption) (Subscriber, error) {
	sOpts := &subscriberOptions{}
	for _, opt := range opts {
		opt(sOpts)
	}
	return ps.subscriber(group, sOpts)
}

func (ps *pubSub) Close() error {
	errors := []error{}
	keys := []string{}
	ps.subs.Range(func(k any, value interface{}) bool {
		keys = append(keys, k.(string))
		s := value.(Subscriber)
		if err := s.Close(); err != nil {
			errors = append(errors, err)
		}

		return false
	})

	if len(errors) > 0 {
		return errors[0]
	}

	for _, k := range keys {
		ps.subs.Delete(k)
	}

	err := ps.publisher.Close()
	if err != nil {
		return err
	}

	return nil
}

func topicName(scope, topic string) string {
	if scope != "" {
		return fmt.Sprintf("%s.%s", scope, topic)
	}

	return topic
}
