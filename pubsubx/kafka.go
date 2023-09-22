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

	kafkaPub := &kafkaPublisher{publisher: *publisher, propagator: opts.propagator}
	return kafkaPub, nil
}

type kafkaPublisher struct {
	publisher  kafka.Publisher
	propagator propagation.TextMapPropagator
}

func (p *kafkaPublisher) Publish(ctx context.Context, topic string, messages ...*message.Message) error {
	if p.propagator != nil {
		for _, msg := range messages {
			p.propagator.Inject(ctx, propagation.MapCarrier(msg.Metadata))
		}
	}
	return p.publisher.Publish(topic, messages...)

}

func (p *kafkaPublisher) Close() error {
	return p.publisher.Close()
}

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

	return subscriber, nil
}
