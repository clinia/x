package kgox

import (
	"errors"
	"sync"

	"github.com/clinia/x/pubsubx"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
)

type PubSub struct {
	conf         *pubsubx.Config
	kotelService *kotel.Kotel
	writeClient  *kgo.Client
	l            *logrusx.Logger

	mu        sync.RWMutex
	consumers map[messagex.ConsumerGroup]*consumer
}

func NewPubSub(l *logrusx.Logger, config *pubsubx.Config, opts *pubsubx.PubSubOptions) (*PubSub, error) {
	// TODO: Use propagator to inject trace context
	if l == nil {
		return nil, errorx.FailedPreconditionErrorf("logger is required")
	}

	if config.Provider != "kafka" {
		return nil, errorx.FailedPreconditionErrorf("unsupported provider %s", config.Provider)
	}

	kopts := []kgo.Opt{
		kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
	}

	// Setup kotel
	var kotelService *kotel.Kotel
	if opts != nil {
		kotelService = newKotel(opts.TracerProvider, opts.Propagator, opts.MeterProvider)
		kopts = append(kopts, kgo.WithHooks(kotelService.Hooks()...))
	}

	wc, err := kgo.NewClient(kopts...)
	if err != nil {
		return nil, errorx.InternalErrorf("failed to create kafka client: %v", err)
	}

	return &PubSub{
		l:            l,
		conf:         config,
		kotelService: kotelService,
		writeClient:  wc,
		consumers:    make(map[messagex.ConsumerGroup]*consumer),
	}, nil
}

// Close implements pubsubx.PubSub.
func (p *PubSub) Close() error {
	p.writeClient.Close()
	errs := make([]error, 0, len(p.consumers))
	// TODO: iterate through all consumer clients and close them.
	for _, c := range p.consumers {
		err := c.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Publisher implements pubsubx.PubSub.
func (p *PubSub) Publisher() pubsubx.Publisher {
	// We can safely cast here because we know that the pubSub struct is a Publisher.
	return (*publisher)(p)
}

// Subscriber implements pubsubx.PubSub.
func (p *PubSub) Subscriber(group string, topics []messagex.Topic, opts ...pubsubx.SubscriberOption) (pubsubx.Subscriber, error) {
	p.mu.RLock()
	if c, ok := p.consumers[messagex.ConsumerGroup(group)]; ok {
		p.mu.RUnlock()
		return c, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	o := pubsubx.NewDefaultSubscriberOptions()
	for _, opt := range opts {
		opt(o)
	}

	cs, err := newConsumer(p.l, p.kotelService, p.conf, group, topics, o)
	if err != nil {
		p.l.Errorf("failed to create consumer: %v", err)
		return nil, errorx.InternalErrorf("failed to create consumer: %v", err)
	}

	p.consumers[messagex.ConsumerGroup(group)] = cs
	return cs, nil
}
