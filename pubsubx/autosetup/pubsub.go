package autosetup

import (
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/kgox"
	"github.com/clinia/x/stringsx"
)

func New(l *logrusx.Logger, c *pubsubx.Config, opts ...pubsubx.PubSubOption) (pubsubx.PubSub, error) {
	pubSubOpts := &pubsubx.PubSubOptions{}
	for _, opt := range opts {
		opt(pubSubOpts)
	}
	return setup(l, c, pubSubOpts)
}

func setup(l *logrusx.Logger, c *pubsubx.Config, opts *pubsubx.PubSubOptions) (pubsubx.PubSub, error) {
	switch f := stringsx.SwitchExact(c.Provider); {
	case f.AddCase("kafka"):
		l.Infof("Kafka pubsub configured! Sending & receiving messages to %s", c.Providers.Kafka.Brokers)
		return kgox.NewPubSub(l, c, opts)

	case f.AddCase("inmemory"):
		ps, err := pubsubx.SetupInMemoryPubSub(l, c)
		if err != nil {
			return nil, err
		}
		l.Infof("InMemory publisher configured! Sending & receiving messages to in-memory")
		return ps, nil
	default:
		return nil, f.ToUnknownCaseErr()
	}
}
