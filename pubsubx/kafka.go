package pubsubx

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx/kafkax"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel/propagation"
)

type kafkaPublisher struct {
	scope      string
	publisher  *kafkax.Publisher
	propagator propagation.TextMapPropagator
}

var _ Publisher = (*kafkaPublisher)(nil)

// TODO: add publisher configs
func setupKafkaPublisher(l *logrusx.Logger, c *Config, opts *pubSubOptions) (Publisher, error) {
	conf := kafkax.PublisherConfig{
		Brokers:   c.Providers.Kafka.Brokers,
		Marshaler: kafka.DefaultMarshaler{},
	}
	// Setup tracer if provided
	if opts.tracerProvider != nil {
		conf.Tracer = NewOTELSaramaTracer(otelsarama.WithTracerProvider(opts.tracerProvider), otelsarama.WithPropagators(opts.propagator))
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
}

var _ Subscriber = (*kafkaSubscriber)(nil)

// TODO: add subscriber configs
func setupKafkaSubscriber(l *logrusx.Logger, c *Config, opts *pubSubOptions, group string, subOpts *subscriberOptions) (Subscriber, error) {
	saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	// saramaSubscriberConfig.Metadata.AllowAutoTopicCreation = false

	conf := kafkax.SubscriberConfig{
		Brokers:               c.Providers.Kafka.Brokers,
		Unmarshaler:           kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: saramaSubscriberConfig,
		Tracer:                NewOTELSaramaTracer(otelsarama.WithTracerProvider(opts.tracerProvider)),
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
		conf.Tracer = NewOTELSaramaTracer(otelsarama.WithTracerProvider(opts.tracerProvider), otelsarama.WithPropagators(opts.propagator))
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

type kafkaAdmin struct {
	saramAdmin sarama.ClusterAdmin
}

func setupKafkaAdmin(l *logrusx.Logger, c *Config) (Admin, error) {
	admin, err := sarama.NewClusterAdmin(c.Providers.Kafka.Brokers, sarama.NewConfig())
	if err != nil {
		return nil, err
	}

	return &kafkaAdmin{
		saramAdmin: admin,
	}, nil
}

func (a *kafkaAdmin) CreateTopic(ctx context.Context, topic string, detail *TopicDetail) error {
	topicDetail := sarama.TopicDetail{
		NumPartitions:     detail.NumPartitions,
		ReplicationFactor: detail.ReplicationFactor,
		ConfigEntries:     detail.ConfigEntries,
	}

	return a.saramAdmin.CreateTopic(topic, &topicDetail, false)
}

func (a *kafkaAdmin) ListTopics(ctx context.Context) (map[string]TopicDetail, error) {
	topics, err := a.saramAdmin.ListTopics()
	if err != nil {
		return nil, err
	}

	topicDetails := map[string]TopicDetail{}
	for topic, detail := range topics {
		topicDetails[topic] = TopicDetail{
			NumPartitions:     detail.NumPartitions,
			ReplicationFactor: detail.ReplicationFactor,
			ConfigEntries:     detail.ConfigEntries,
		}
	}

	return topicDetails, nil
}

func (a *kafkaAdmin) DeleteTopic(ctx context.Context, topic string) error {
	return a.saramAdmin.DeleteTopic(topic)
}

func (a *kafkaAdmin) ListSubscribers(ctx context.Context, topic string) (map[string]string, error) {
	cgroups, err := a.saramAdmin.ListConsumerGroups()
	if err != nil {
		return nil, err
	}

	subscribers := map[string]string{}
	for _, cgroup := range cgroups {
		subscribers[cgroup] = topic
	}

	return subscribers, nil
}

func (a *kafkaAdmin) DeleteSubscriber(ctx context.Context, subscriber string) error {
	return a.saramAdmin.DeleteConsumerGroup(subscriber)
}
