package kafkax_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/clinia/x/pubsubx/kafkax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/subscriber"
	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
)

func kafkaBrokers() []string {
	brokers := os.Getenv("WATERMILL_TEST_KAFKA_BROKERS")
	if brokers != "" {
		return strings.Split(brokers, ",")
	}
	return []string{"localhost:9091", "localhost:9092", "localhost:9093", "localhost:9094", "localhost:9095"}
}

func newPubSub(t *testing.T, marshaler kafkax.MarshalerUnmarshaler, consumerGroup string, consumerModel kafkax.ConsumerModel) (*kafkax.Publisher, *kafkax.Subscriber) {
	logger := watermill.NewStdLogger(true, true)

	var err error
	var publisher *kafkax.Publisher

	retriesLeft := 5
	for {
		publisher, err = kafkax.NewPublisher(kafkax.PublisherConfig{
			Brokers:   kafkaBrokers(),
			Marshaler: marshaler,
		}, logger)
		if err == nil || retriesLeft == 0 {
			break
		}

		retriesLeft--
		fmt.Printf("cannot create kafka Publisher: %s, retrying (%d retries left)", err, retriesLeft)
		time.Sleep(time.Second * 2)
	}
	require.NoError(t, err)

	saramaConfig := kafkax.DefaultSaramaSubscriberConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	saramaConfig.Admin.Timeout = time.Second * 30
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.ChannelBufferSize = 10240
	saramaConfig.Consumer.Group.Heartbeat.Interval = time.Millisecond * 500
	saramaConfig.Consumer.Group.Rebalance.Timeout = time.Second * 3

	var subscriber *kafkax.Subscriber

	retriesLeft = 5
	for {
		subscriber, err = kafkax.NewSubscriber(
			kafkax.SubscriberConfig{
				Brokers:               kafkaBrokers(),
				Unmarshaler:           marshaler,
				OverwriteSaramaConfig: saramaConfig,
				ConsumerGroup:         consumerGroup,
				InitializeTopicDetails: &sarama.TopicDetail{
					NumPartitions:     8,
					ReplicationFactor: 1,
				},
				BatchConsumerConfig: &kafkax.BatchConsumerConfig{
					MaxBatchSize: 10,
					MaxWaitTime:  100 * time.Millisecond,
				},
			},
			logger,
		)
		if err == nil || retriesLeft == 0 {
			break
		}

		retriesLeft--
		fmt.Printf("cannot create kafka Subscriber: %s, retrying (%d retries left)", err, retriesLeft)
		time.Sleep(time.Second * 2)
	}

	require.NoError(t, err)

	return publisher, subscriber
}

func generatePartitionKey(topic string, msg *message.Message) (string, error) {
	return msg.Metadata.Get("partition_key"), nil
}

func createPubSubWithConsumerGroup(consumerModel kafkax.ConsumerModel) func(t *testing.T, consumerGroup string) (message.Publisher, message.Subscriber) {
	return func(t *testing.T, consumerGroup string) (message.Publisher, message.Subscriber) {
		return newPubSub(t, kafkax.DefaultMarshaler{}, consumerGroup, consumerModel)
	}
}

func createPubSub(consumerModel kafkax.ConsumerModel) func(t *testing.T) (message.Publisher, message.Subscriber) {
	return func(t *testing.T) (message.Publisher, message.Subscriber) {
		return createPubSubWithConsumerGroup(consumerModel)(t, "test")
	}
}

func createPartitionedPubSub(consumerModel kafkax.ConsumerModel) func(*testing.T) (message.Publisher, message.Subscriber) {
	return func(t *testing.T) (message.Publisher, message.Subscriber) {
		return newPubSub(t, kafkax.NewWithPartitioningMarshaler(generatePartitionKey), "test", consumerModel)
	}
}

func createNoGroupPubSub(consumerModel kafkax.ConsumerModel) func(t *testing.T) (message.Publisher, message.Subscriber) {
	return func(t *testing.T) (message.Publisher, message.Subscriber) {
		return newPubSub(t, kafkax.DefaultMarshaler{}, "", consumerModel)
	}
}

func TestPublishSubscribe(t *testing.T) {
	testCases := []struct {
		name          string
		consumerModel kafkax.ConsumerModel
	}{
		{
			name:          "with default consumer",
			consumerModel: kafkax.Default,
		},
		{
			name:          "with batch config",
			consumerModel: kafkax.Batch,
		},
		{
			name:          "with partition concurrent config",
			consumerModel: kafkax.PartitionConcurrent,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			features := tests.Features{
				ConsumerGroups:      true,
				ExactlyOnceDelivery: false,
				GuaranteedOrder:     false,
				Persistent:          true,
			}

			tests.TestPubSub(
				t,
				features,
				createPubSub(test.consumerModel),
				createPubSubWithConsumerGroup(test.consumerModel),
			)
		})
	}
}

func TestPublishSubscribe_ordered(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long tests")
	}

	testCases := []struct {
		name          string
		consumerModel kafkax.ConsumerModel
	}{
		{
			name:          "with default consumer",
			consumerModel: kafkax.Default,
		},
		{
			name:          "with partition concurrent config",
			consumerModel: kafkax.PartitionConcurrent,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			tests.TestPubSub(
				t,
				tests.Features{
					ConsumerGroups:      true,
					ExactlyOnceDelivery: false,
					GuaranteedOrder:     true,
					Persistent:          true,
				},
				createPartitionedPubSub(test.consumerModel),
				createPubSubWithConsumerGroup(test.consumerModel),
			)
		})
	}

	t.Run("with batch consumer config", func(t *testing.T) {
		testBulkMessageHandlerPubSub(
			t,
			tests.Features{
				ConsumerGroups:      true,
				ExactlyOnceDelivery: false,
				GuaranteedOrder:     true,
				Persistent:          true,
			},
			createPartitionedPubSub(kafkax.Batch),
			createPubSubWithConsumerGroup(kafkax.Batch),
		)
	})
}

func TestNoGroupSubscriber(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long tests")
	}
	testCases := []struct {
		name          string
		consumerModel kafkax.ConsumerModel
	}{
		{
			name:          "with default consumer",
			consumerModel: kafkax.Default,
		},
		{
			name:          "with batch config",
			consumerModel: kafkax.Batch,
		},
		{
			name:          "with partition concurrent config",
			consumerModel: kafkax.PartitionConcurrent,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			tests.TestPubSub(
				t,
				tests.Features{
					ConsumerGroups:                   false,
					ExactlyOnceDelivery:              false,
					GuaranteedOrder:                  false,
					Persistent:                       true,
					NewSubscriberReceivesOldMessages: true,
				},
				createNoGroupPubSub(test.consumerModel),
				nil,
			)
		})
	}
}

func TestCtxValues(t *testing.T) {
	tests := []struct {
		name          string
		consumerModel kafkax.ConsumerModel
	}{
		{
			name:          "with default consumer",
			consumerModel: kafkax.Default,
		},
		{
			name:          "with batch config",
			consumerModel: kafkax.Batch,
		},
		{
			name:          "with partition concurrent config",
			consumerModel: kafkax.PartitionConcurrent,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pub, sub := newPubSub(t, kafkax.DefaultMarshaler{}, "", test.consumerModel)
			topicName := "topic_" + watermill.NewUUID()

			var messagesToPublish []*message.Message

			for i := 0; i < 20; i++ {
				id := watermill.NewUUID()
				messagesToPublish = append(messagesToPublish, message.NewMessage(id, nil))
			}
			err := pub.Publish(topicName, messagesToPublish...)
			require.NoError(t, err, "cannot publish message")

			messages, err := sub.Subscribe(context.Background(), topicName)
			require.NoError(t, err)

			receivedMessages, all := subscriber.BulkReadWithDeduplication(messages, len(messagesToPublish), time.Second*10)
			require.True(t, all)

			expectedPartitionsOffsets := map[int32]int64{}
			for _, msg := range receivedMessages {
				partition, ok := kafkax.MessagePartitionFromCtx(msg.Context())
				assert.True(t, ok)

				messagePartitionOffset, ok := kafkax.MessagePartitionOffsetFromCtx(msg.Context())
				assert.True(t, ok)

				kafkaMsgTimestamp, ok := kafkax.MessageTimestampFromCtx(msg.Context())
				assert.True(t, ok)
				assert.NotZero(t, kafkaMsgTimestamp)

				_, ok = kafkax.MessageKeyFromCtx(msg.Context())
				assert.True(t, ok)

				if expectedPartitionsOffsets[partition] <= messagePartitionOffset {
					// kafka partition offset is offset of the last message + 1
					expectedPartitionsOffsets[partition] = messagePartitionOffset + 1
				}
			}
			assert.NotEmpty(t, expectedPartitionsOffsets)

			offsets, err := sub.PartitionOffset(topicName)
			require.NoError(t, err)
			assert.NotEmpty(t, offsets)

			assert.EqualValues(t, expectedPartitionsOffsets, offsets)

			require.NoError(t, pub.Close())
		})
	}
}
