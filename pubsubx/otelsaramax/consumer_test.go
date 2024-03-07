package otelsaramax

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/trace"
)

const (
	topic = "test-topic"
)

func TestConsumerConsumePartitionWithError(t *testing.T) {
	// Mock partition consumer controller
	mockConsumer := mocks.NewConsumer(t, sarama.NewConfig())
	mockConsumer.ExpectConsumePartition(topic, 0, 0)

	consumer := WrapConsumer(mockConsumer, ConsumerInfo{})
	_, err := consumer.ConsumePartition(topic, 0, 0)
	assert.NoError(t, err)
	// Consume twice
	_, err = consumer.ConsumePartition(topic, 0, 0)
	assert.Error(t, err)
}

func BenchmarkWrapPartitionConsumer(b *testing.B) {
	// Mock provider
	provider := trace.NewNoopTracerProvider()

	mockPartitionConsumer, partitionConsumer := createMockPartitionConsumer(b)

	partitionConsumer = WrapPartitionConsumer(partitionConsumer, ConsumerInfo{}, WithTracerProvider(provider))
	message := sarama.ConsumerMessage{Key: []byte("foo")}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mockPartitionConsumer.YieldMessage(&message)
		<-partitionConsumer.Messages()
	}
}

func BenchmarkMockPartitionConsumer(b *testing.B) {
	mockPartitionConsumer, partitionConsumer := createMockPartitionConsumer(b)

	message := sarama.ConsumerMessage{Key: []byte("foo")}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mockPartitionConsumer.YieldMessage(&message)
		<-partitionConsumer.Messages()
	}
}

func createMockPartitionConsumer(b *testing.B) (*mocks.PartitionConsumer, sarama.PartitionConsumer) {
	// Mock partition consumer controller
	consumer := mocks.NewConsumer(b, sarama.NewConfig())
	mockPartitionConsumer := consumer.ExpectConsumePartition(topic, 0, 0)

	// Create partition consumer
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, 0)
	require.NoError(b, err)
	return mockPartitionConsumer, partitionConsumer
}
