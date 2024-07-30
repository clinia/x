package pubsubx

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/otelx/instrumentation/otelsaramax"
	"github.com/clinia/x/pubsubx/kafkax"
	"go.opentelemetry.io/otel/propagation"
)

type kafkaPublisher struct {
	scope      string
	publisher  *kafkax.Publisher
	propagator propagation.TextMapPropagator
}

var _ Publisher = (*kafkaPublisher)(nil)

func setupKafkaPublisher(l *logrusx.Logger, c *Config, opts *pubSubOptions) (Publisher, error) {
	conf := kafkax.PublisherConfig{
		Brokers:               c.Providers.Kafka.Brokers,
		OverwriteSaramaConfig: opts.SaramaPublisherConfig,
		Marshaler:             kafkax.DefaultMarshaler{},
	}
	// Setup tracer if provided
	if opts.tracerProvider != nil {
		conf.Tracer = kafkax.NewOTELSaramaTracer(otelsaramax.WithTracerProvider(opts.tracerProvider), otelsaramax.WithPropagators(opts.propagator))
	}

	publisher, err := kafkax.NewPublisher(
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

func (p *kafkaPublisher) BulkPublish(ctx context.Context, topic string, messages ...*message.Message) error {
	if p.propagator != nil {
		for _, msg := range messages {
			p.propagator.Inject(ctx, propagation.MapCarrier(msg.Metadata))
		}
	}
	return p.publisher.BulkPublish(topicName(p.scope, topic), messages...)
}

func (p *kafkaPublisher) Close() error {
	return p.publisher.Close()
}

type kafkaSubscriber struct {
	scope      string
	subscriber *kafkax.Subscriber
	onClosed   func(error)
}

var _ Subscriber = (*kafkaSubscriber)(nil)

func setupKafkaSubscriber(l *logrusx.Logger, c *Config, opts *pubSubOptions, group string, subOpts *subscriberOptions, onClosed func(error)) (Subscriber, error) {

	conf := kafkax.SubscriberConfig{
		Brokers:               c.Providers.Kafka.Brokers,
		OverwriteSaramaConfig: opts.SaramaSubscriberConfig,
		Unmarshaler:           kafkax.DefaultMarshaler{},
		Tracer:                kafkax.NewOTELSaramaTracer(otelsaramax.WithTracerProvider(opts.tracerProvider)),
	}

	switch subOpts.consumerModel {
	case ConsumerModelDefault:
		conf.ConsumerModel = kafkax.Default
	case ConsumerModelBatch:
		conf.ConsumerModel = kafkax.Batch
		if subOpts.batchConsumerOptions != nil {
			conf.BatchConsumerConfig = &kafkax.BatchConsumerConfig{
				MaxBatchSize: subOpts.batchConsumerOptions.MaxBatchSize,
				MaxWaitTime:  subOpts.batchConsumerOptions.MaxWaitTime,
			}
		} else {
			// Will use defaults
			conf.BatchConsumerConfig = &kafkax.BatchConsumerConfig{}
		}
	default:
		conf.ConsumerModel = kafkax.Default
	}

	// Setup tracer if provided
	if opts.tracerProvider != nil {
		conf.Tracer = kafkax.NewOTELSaramaTracer(otelsaramax.WithTracerProvider(opts.tracerProvider), otelsaramax.WithPropagators(opts.propagator))
	}

	if group != "" {
		conf.ConsumerGroup = group
	}

	subscriber, err := kafkax.NewSubscriber(
		conf,
		NewLogrusLogger(l.Logger),
	)
	if err != nil {
		return nil, err
	}

	return &kafkaSubscriber{
		scope:      c.Scope,
		subscriber: subscriber,
		onClosed:   onClosed,
	}, nil
}

// Close implements Subscriber.
func (s *kafkaSubscriber) Close() error {
	err := s.subscriber.Close()
	s.onClosed(err)
	return err
}

// Subscribe implements Subscriber.
func (s *kafkaSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return s.subscriber.Subscribe(ctx, topicName(s.scope, topic))
}
