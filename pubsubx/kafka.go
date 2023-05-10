package pubsubx

import (
	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/clinia/x/logrusx"
)

// TODO: add publisher configs
func SetupKafkaPublisher(l *logrusx.Logger, c *Config) (Publisher, error) {
	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   c.Providers.Kafka.Brokers,
			Marshaler: kafka.DefaultMarshaler{},
		},
		NewLogrusLogger(l.Logger),
	)
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

// TODO: add subscriber configs
func SetupKafkaSubscriber(l *logrusx.Logger, c *Config) (Subscriber, error) {
	saramaSubscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	saramaSubscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	subscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               []string{"kafka:9092"},
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: saramaSubscriberConfig,
		},
		NewLogrusLogger(l.Logger),
	)
	if err != nil {
		return nil, err
	}

	return subscriber, nil
}
