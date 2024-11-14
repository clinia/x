package kgox

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/clinia/x/pointerx"
	"github.com/clinia/x/pubsubx"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
)

type PubSub struct {
	conf                            *pubsubx.Config
	kopts                           []kgo.Opt
	defaultCreateTopicConfigEntries map[string]*string
	kotelService                    *kotel.Kotel
	writeClient                     *kgo.Client
	l                               *logrusx.Logger

	mu        sync.RWMutex
	consumers map[messagex.ConsumerGroup]*consumer
}

var _ pubsubx.PubSub = (*PubSub)(nil)

func NewPubSub(l *logrusx.Logger, config *pubsubx.Config, opts *pubsubx.PubSubOptions) (*PubSub, error) {
	if l == nil {
		return nil, errorx.FailedPreconditionErrorf("logger is required")
	}

	if config == nil {
		return nil, errorx.FailedPreconditionErrorf("config is required")
	}

	if config.Provider != "kafka" {
		return nil, errorx.FailedPreconditionErrorf("unsupported provider %s", config.Provider)
	}

	kopts := []kgo.Opt{
		kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
	}

	// Setup kotel
	var kotelService *kotel.Kotel
	defaultCreateTopicConfigEntries := map[string]*string{}
	if opts != nil {
		kotelService = newKotel(opts.TracerProvider, opts.Propagator, opts.MeterProvider)
		kopts = append(kopts, kgo.WithHooks(kotelService.Hooks()...))

		if opts.MaxMessageByte != nil {
			defaultCreateTopicConfigEntries["max.message.bytes"] = pointerx.Ptr(fmt.Sprintf("%d", *opts.MaxMessageByte))
			kopts = append(kopts, kgo.ProducerBatchMaxBytes(int32(*opts.MaxMessageByte)))
		}
		if opts.RetentionMs != nil {
			defaultCreateTopicConfigEntries["retention.ms"] = pointerx.Ptr(fmt.Sprintf("%d", *opts.RetentionMs))
		}
	}

	wc, err := kgo.NewClient(kopts...)
	if err != nil {
		return nil, errorx.InternalErrorf("failed to create kafka client: %v", err)
	}

	if config.PoisonQueue.Enabled {
		poisonQueueTopic := messagex.TopicFromName(config.PoisonQueue.TopicName)
		adminClient := kadm.NewClient(wc)
		var replicationFactor int16 = math.MaxInt16
		if len(config.Providers.Kafka.Brokers) <= math.MaxInt16 {
			replicationFactor = int16(len(config.Providers.Kafka.Brokers))
		}
		_, err := adminClient.CreateTopic(context.Background(), 1, replicationFactor, defaultCreateTopicConfigEntries, poisonQueueTopic.TopicName(config.Scope))
		if err != nil && err.Error() != kerr.TopicAlreadyExists.Error() {
			return nil, errorx.InternalErrorf("failed to create poison queue: %v", err)
		}
	}

	return &PubSub{
		l:                               l,
		conf:                            config,
		kotelService:                    kotelService,
		kopts:                           kopts,
		defaultCreateTopicConfigEntries: defaultCreateTopicConfigEntries,
		writeClient:                     wc,
		consumers:                       make(map[messagex.ConsumerGroup]*consumer),
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

// PoisonQueueHandler implements pubsubx.PubSub.
func (p *PubSub) PoisonQueueHandler() PoisonQueueHandler {
	// We can safely cast here because we know that the pubSub struct is a Publisher.
	return (*poisonQueueHandler)(p)
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

	cs, err := newConsumer(p.l, p.kotelService, p.conf, group, topics, o, p.PoisonQueueHandler())
	if err != nil {
		p.l.Errorf("failed to create consumer: %v", err)
		return nil, errorx.InternalErrorf("failed to create consumer: %v", err)
	}

	p.consumers[messagex.ConsumerGroup(group)] = cs
	return cs, nil
}

// AdminClient implements pubsubx.PubSub.
func (p *PubSub) AdminClient() (pubsubx.PubSubAdminClient, error) {
	wc, err := kgo.NewClient(p.kopts...)
	if err != nil {
		return nil, errorx.InternalErrorf("failed to create kafka client: %v", err)
	}

	admClient := NewPubSubAdminClient(kadm.NewClient(wc), p.defaultCreateTopicConfigEntries)
	return admClient, nil
}
