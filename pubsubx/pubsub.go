package pubsubx

import (
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/stringsx"
)

type PubSub struct {
	Publisher  Publisher
	Subscriber Subscriber
}

func New(l *logrusx.Logger, c *Config) (*PubSub, error) {
	ps := &PubSub{}

	if err := ps.setup(l, c); err != nil {
		return nil, err
	}

	return ps, nil
}

func (ps *PubSub) setup(l *logrusx.Logger, c *Config) error {
	switch f := stringsx.SwitchExact(c.Provider); {
	case f.AddCase("kafka"):
		publisher, err := SetupKafkaPublisher(l, c)
		if err != nil {
			return err
		}

		subscriber, err := SetupKafkaSubscriber(l, c)
		if err != nil {
			return err
		}

		l.Infof("Kafka pubsub configured! Sending & receiving messages to %s", c.Providers.Kafka.Brokers)
		ps.Publisher = publisher
		ps.Subscriber = subscriber

	case f.AddCase("inmemory"):
		pubsub, err := SetupInMemoryPubSub(l, c)
		if err != nil {
			return err
		}

		l.Infof("InMemory publisher configured! Sending & receiving messages to in-memory")
		ps.Publisher = pubsub
		ps.Subscriber = pubsub
	default:
		return f.ToUnknownCaseErr()
	}

	return nil
}
