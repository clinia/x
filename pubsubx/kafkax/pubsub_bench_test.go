package kafkax_test

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/clinia/x/pubsubx/kafkax"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
)

func BenchmarkSubscriber(b *testing.B) {
	runBenchmark(b, kafkax.Default)
}

func BenchmarkSubscriberBatch(b *testing.B) {
	runBenchmark(b, kafkax.Batch)
}

func BenchmarkSubscriberPartitionConcurrent(b *testing.B) {
	runBenchmark(b, kafkax.PartitionConcurrent)
}

func runBenchmark(b *testing.B, consumerModel kafkax.ConsumerModel) {
	tests.BenchSubscriber(b, func(n int) (message.Publisher, message.Subscriber) {
		logger := watermill.NopLogger{}

		publisher, err := kafkax.NewPublisher(kafkax.PublisherConfig{
			Brokers:   kafkaBrokers(),
			Marshaler: kafkax.DefaultMarshaler{},
		}, logger)
		if err != nil {
			panic(err)
		}

		saramaConfig := kafkax.DefaultSaramaSubscriberConfig()
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

		subscriber, err := kafkax.NewSubscriber(
			kafkax.SubscriberConfig{
				Brokers:               kafkaBrokers(),
				Unmarshaler:           kafkax.DefaultMarshaler{},
				OverwriteSaramaConfig: saramaConfig,
				ConsumerGroup:         "test",
				ConsumerModel:         consumerModel,
			},
			logger,
		)
		if err != nil {
			panic(err)
		}

		return publisher, subscriber
	})
}
