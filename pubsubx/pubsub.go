package pubsubx

import (
	"sync"

	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/stringsx"
)

type PubSub interface {
	Publisher() Publisher
	Subscriber(group string) (Subscriber, error)
	// CLoses all publishers and subscribers.
	Close() error
}

type pubSub struct {
	publisher  Publisher
	subscriber func(group string) (Subscriber, error)
	subs       sync.Map
}

var _ PubSub = (*pubSub)(nil)

func New(l *logrusx.Logger, c *Config) (PubSub, error) {
	ps := &pubSub{
		subs: sync.Map{},
	}

	if err := ps.setup(l, c); err != nil {
		return nil, err
	}

	return ps, nil
}

func (ps *pubSub) setup(l *logrusx.Logger, c *Config) error {
	switch f := stringsx.SwitchExact(c.Provider); {
	case f.AddCase("kafka"):
		publisher, err := SetupKafkaPublisher(l, c)
		if err != nil {
			return err
		}

		ps.publisher = publisher
		ps.subscriber = func(group string) (Subscriber, error) {
			if ms, ok := ps.subs.Load(group); !ok {
				s, e := SetupKafkaSubscriber(l, c, group)
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

		ps.publisher = pubsub
		ps.subscriber = pubsub.SetupSubscriber()
		l.Infof("InMemory publisher configured! Sending & receiving messages to in-memory")
	default:
		return f.ToUnknownCaseErr()
	}

	return nil
}

func (ps *pubSub) Publisher() Publisher {
	return ps.publisher
}

func (ps *pubSub) Subscriber(group string) (Subscriber, error) {
	return ps.subscriber(group)
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
