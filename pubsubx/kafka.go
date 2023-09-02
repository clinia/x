package pubsubx

import (
	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/clinia/x/logrusx"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
)

// TODO: add publisher configs
func SetupKafkaPublisher(l *logrusx.Logger, c *Config, opts *pubSubOptions) (Publisher, error) {
	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   c.Providers.Kafka.Brokers,
			Marshaler: kafka.DefaultMarshaler{},
			Tracer:    NewOTELSaramaTracer(otelsarama.WithTracerProvider(opts.tracerProvider)),
		},
		NewLogrusLogger(l.Logger),
	)
	if err != nil {
		return nil, err
	}

	return publisher, nil
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
