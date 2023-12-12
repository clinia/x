package pubsubx

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/clinia/x/logrusx"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel/propagation"
)

type kafkaPublisher struct {
	scope      string
	publisher  *kafka.Publisher
	propagator propagation.TextMapPropagator
}

var _ Publisher = (*kafkaPublisher)(nil)

// TODO: add publisher configs
func SetupKafkaPublisher(l *logrusx.Logger, c *Config, opts *pubSubOptions) (Publisher, error) {
	conf := kafka.PublisherConfig{
		Brokers:   c.Providers.Kafka.Brokers,
		Marshaler: kafka.DefaultMarshaler{},
	}
	// Setup tracer if provided
	if opts.tracerProvider != nil {
		conf.Tracer = NewOTELSaramaTracer(otelsarama.WithTracerProvider(opts.tracerProvider), otelsarama.WithPropagators(opts.propagator))
	}

	publisher, err := kafka.NewPublisher(
		conf,
		NewLogrusLogger(l.Logger),
	)
	if err != nil {
		return nil, err
	}

	return &kafkaPublisher{
		scope:      c.Scope,
		publisher:  publisher,
		propagator: opts.propagator,
	}, nil
}

func (p *kafkaPublisher) Publish(ctx context.Context, topic string, messages ...*message.Message) error {
	if p.propagator != nil {
		for _, msg := range messages {
			p.propagator.Inject(ctx, propagation.MapCarrier(msg.Metadata))
		}
	}
	return p.publisher.Publish(topicName(p.scope, topic), messages...)

}

func (p *kafkaPublisher) Close() error {
	return p.publisher.Close()
}

type kafkaSubscriber struct {
	scope      string
	subscriber *kafka.Subscriber
}

var _ Subscriber = (*kafkaSubscriber)(nil)

// TODO: add subscriber configs
func SetupKafkaSubscriber(l *logrusx.Logger, c *Config, opts *pubSubOptions, group string) (Subscriber, error) {
	saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	conf := kafka.SubscriberConfig{
		Brokers:               c.Providers.Kafka.Brokers,
		Unmarshaler:           kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: saramaSubscriberConfig,
		Tracer:                NewOTELSaramaTracer(otelsarama.WithTracerProvider(opts.tracerProvider)),
	}
	// Setup tracer if provided
	if opts.tracerProvider != nil {
		conf.Tracer = NewOTELSaramaTracer(otelsarama.WithTracerProvider(opts.tracerProvider), otelsarama.WithPropagators(opts.propagator))
	}

	if group != "" {
		conf.ConsumerGroup = group
	}

	subscriber, err := kafka.NewSubscriber(
		conf,
		NewLogrusLogger(l.Logger),
	)
	if err != nil {
		return nil, err
	}

	return &kafkaSubscriber{
		scope:      c.Scope,
		subscriber: subscriber,
	}, nil
}

// Close implements Subscriber.
func (s *kafkaSubscriber) Close() error {
	return s.subscriber.Close()
}

// Subscribe implements Subscriber.
func (s *kafkaSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return s.subscriber.Subscribe(ctx, topicName(s.scope, topic))
}
