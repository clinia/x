package kgox

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/clinia/x/pointerx"
	"github.com/clinia/x/pubsubx"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

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
	mp                              metric.MeterProvider

	mu        sync.RWMutex
	consumers map[messagex.ConsumerGroup]*consumer
}

var _ pubsubx.PubSub = (*PubSub)(nil)

type contextLoggerKey string

const ctxLoggerKey contextLoggerKey = "consumer_logger"

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
		kgo.WithLogger(&pubsubLogger{l: l}),
	}

	// Setup kotel
	var kotelService *kotel.Kotel
	defaultCreateTopicConfigEntries := map[string]*string{}
	var mp metric.MeterProvider
	if opts != nil {
		mp = opts.MeterProvider
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
	if mp == nil {
		l.Warnf("no meter provider was defined in pubsub options, using noop")
		mp = noop.NewMeterProvider()
	}

	wc, err := kgo.NewClient(kopts...)
	if err != nil {
		return nil, errorx.InternalErrorf("failed to create kafka client: %v", err)
	}

	ps := &PubSub{
		l:                               l,
		conf:                            config,
		mp:                              mp,
		kotelService:                    kotelService,
		kopts:                           kopts,
		defaultCreateTopicConfigEntries: defaultCreateTopicConfigEntries,
		writeClient:                     wc,
		consumers:                       make(map[messagex.ConsumerGroup]*consumer),
	}

	if err := ps.Bootstrap(); err != nil {
		return nil, err
	}
	return ps, nil
}

func (p *PubSub) Bootstrap() error {
	if !p.conf.PoisonQueue.Enabled {
		return nil
	}

	if p.conf.PoisonQueue.TopicName == "" {
		return errorx.InternalErrorf("failed to create poison queue since topic name is empty")
	}
	poisonQueueTopic := messagex.TopicFromName(p.conf.PoisonQueue.TopicName)
	adminClient := kadm.NewClient(p.writeClient)
	var replicationFactor int16 = math.MaxInt16
	if len(p.conf.Providers.Kafka.Brokers) <= math.MaxInt16 {
		//nolint:all
		replicationFactor = int16(len(p.conf.Providers.Kafka.Brokers))
	}
	m, err := adminClient.Metadata(context.Background())
	if err != nil {
		return errorx.InternalErrorf("failed to validate poison queue existance : %s", err.Error())
	}
	if !m.Topics.Has(poisonQueueTopic.TopicName(p.conf.Scope)) {
		_, err = adminClient.CreateTopic(context.Background(), 1, replicationFactor, p.defaultCreateTopicConfigEntries, poisonQueueTopic.TopicName(p.conf.Scope))
		if err != nil && err.Error() != kerr.TopicAlreadyExists.Error() {
			return errorx.InternalErrorf("failed to create poison queue: %s", err.Error())
		} else if err == nil && p.l != nil {
			p.l.Infof("poison queue topic %s created", poisonQueueTopic.TopicName(p.conf.Scope))
		}
	} else {
		if p.l != nil {
			p.l.Debugf("poison queue topic %s already exist", poisonQueueTopic.TopicName(p.conf.Scope))
		}
	}

	return nil
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
	return (*poisonQueueHandler)(p)
}

// Subscriber implements pubsubx.PubSub.
func (p *PubSub) Subscriber(group string, topics []messagex.Topic, opts ...pubsubx.SubscriberOption) (pubsubx.Subscriber, error) {
	p.mu.RLock()
	consumerGroup := messagex.ConsumerGroup(group)
	if c, ok := p.consumers[consumerGroup]; ok {
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

	var m metric.Meter
	if p.mp != nil {
		m = p.mp.Meter("pubsubx_consumer")
	}
	cs, err := newConsumer(context.Background(), p.l, p.kotelService, p.conf, consumerGroup, topics, o, p.eventRetryHandler(consumerGroup, o), p.PoisonQueueHandler(), m)
	if err != nil {
		p.l.Errorf("failed to create consumer: %v", err)
		return nil, errorx.InternalErrorf("failed to create consumer: %v", err)
	}

	p.consumers[consumerGroup] = cs
	return cs, nil
}

func (p *PubSub) eventRetryHandler(group messagex.ConsumerGroup, opts *pubsubx.SubscriberOptions) *eventRetryHandler {
	if opts == nil {
		opts = pubsubx.NewDefaultSubscriberOptions()
	}
	return &eventRetryHandler{
		p,
		group,
		opts,
	}
}

// AdminClient implements pubsubx.PubSub.
func (p *PubSub) AdminClient() (pubsubx.PubSubAdminClient, error) {
	wc, err := kgo.NewClient(p.kopts...)
	if err != nil {
		return nil, errorx.InternalErrorf("failed to create kafka client: %v", err)
	}

	admClient := NewPubSubAdminClient(wc, p.conf, p.defaultCreateTopicConfigEntries)
	return admClient, nil
}

// getContextLogger allows to extract the logger set in the context if we have some contextual logger
// that is used
func getContextLogger(ctx context.Context, fallback *logrusx.Logger) (l *logrusx.Logger) {
	if ctxL := ctx.Value(ctxLoggerKey); ctxL != nil {
		if ctxL, ok := ctxL.(*logrusx.Logger); ok {
			l = ctxL.WithContext(ctx)
		}
	}
	if l == nil {
		l = fallback.WithContext(ctx)
	}
	return l
}
