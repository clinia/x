package autosetup

import (
	"context"
	"fmt"

	"github.com/clinia/x/loggerx"
	"github.com/clinia/x/pubsubx"
	inmemorypubsub "github.com/clinia/x/pubsubx/inmemory"
	"github.com/clinia/x/pubsubx/kgox"
	"github.com/clinia/x/stringsx"
	"go.opentelemetry.io/otel/attribute"
)

func New(ctx context.Context, c *pubsubx.Config, opts ...pubsubx.PubSubOption) (pubsubx.PubSub, error) {
	pubSubOpts := &pubsubx.PubSubOptions{}
	for _, opt := range opts {
		opt(pubSubOpts)
	}
	l := loggerx.NewDefaultLogger().WithFields(attribute.String("component", "pubsubx.autosetup"))
	return setup(ctx, l, c, pubSubOpts)
}

func setup(ctx context.Context, l *loggerx.Logger, c *pubsubx.Config, opts *pubsubx.PubSubOptions) (pubsubx.PubSub, error) {
	switch f := stringsx.SwitchExact(c.Provider); {
	case f.AddCase("kafka"):
		l.Info(ctx, fmt.Sprintf("Kafka pubsub configured! Sending & receiving messages to %s", c.Providers.Kafka.Brokers))
		return kgox.NewPubSub(ctx, l, c, opts)

	case f.AddCase("inmemory"):
		ps, err := inmemorypubsub.SetupInMemoryPubSub(l, c)
		if err != nil {
			return nil, err
		}
		l.Info(ctx, "InMemory publisher configured! Sending & receiving messages to in-memory")
		return ps, nil
	default:
		return nil, f.ToUnknownCaseErr()
	}
}
