package kgox

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	defaultExpectedReceiveTimeout   = 30 * time.Second
	defaultExpectedNoReceiveTimeout = 5 * time.Second
)

func getRandomGroupTopics(t *testing.T, count int) (Group string, Topics []messagex.Topic) {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "-")
	name = name[:len(name)/2]
	rand := ksuid.New().String()
	rand = rand[:len(rand)/2]
	Group = fmt.Sprintf("%s-group", rand)
	topicName := fmt.Sprintf("%s-%s", name, rand)
	for i := 0; i < count; i++ {
		Topics = append(Topics, messagex.Topic(fmt.Sprintf("%s-%d", topicName, i)))
	}

	return
}

func getPubsubConfig(t *testing.T, retry bool) *pubsubx.Config {
	t.Helper()

	kafkaURLs := []string{"localhost:19092", "localhost:29092", "localhost:39092"}
	kafkaURLsFromEnv := os.Getenv("KAFKA")
	if len(kafkaURLsFromEnv) > 0 {
		kafkaURLs = strings.Split(kafkaURLsFromEnv, ",")
	}

	return &pubsubx.Config{
		Scope:    "test-scope",
		Provider: "kafka",
		Providers: pubsubx.ProvidersConfig{
			Kafka: pubsubx.KafkaConfig{
				Brokers: kafkaURLs,
			},
		},
		TopicRetry: retry,
		PoisonQueue: pubsubx.PoisonQueueConfig{
			TopicName: "poison-queue",
			Enabled:   false,
		},
		EnableAutoCommit: false,
	}
}

func getPubSubConfigWithCustomPoisonQueue(t *testing.T, retry bool, customPoisonQueue string) *pubsubx.Config {
	t.Helper()
	config := getPubsubConfig(t, retry)
	config.PoisonQueue.TopicName = customPoisonQueue
	config.PoisonQueue.Enabled = true
	return config
}

func createTopic(t *testing.T, conf *pubsubx.Config, topic messagex.Topic) {
	client, err := kgo.NewClient(kgo.SeedBrokers(conf.Providers.Kafka.Brokers...))
	require.NoError(t, err)

	admCl := kadm.NewClient(client)
	deleteTopic := func() error {
		// Delete the topic
		res, err := admCl.DeleteTopics(context.Background(), topic.TopicName(conf.Scope))
		return errors.Join(err, res.Error())
	}

	// Ensure the topic is deleted
	deleteTopic()

	// Create the topic
	res, err := admCl.CreateTopics(context.Background(), 1, 1, map[string]*string{}, topic.TopicName(conf.Scope))
	require.NoError(t, err, "failed to create topic '%s'", topic.TopicName(conf.Scope))
	require.NoError(t, res.Error(), "failed to create topic '%s'", topic.TopicName(conf.Scope))
	t.Logf("created topic '%s'", topic.TopicName(conf.Scope))

	t.Cleanup(func() {
		err := deleteTopic()
		require.NoError(t, err)
	})
}

func getPublisher(t *testing.T, l *logrusx.Logger, conf *pubsubx.Config) *publisher {
	t.Helper()
	pubSub, err := NewPubSub(l, conf, nil)
	require.NoError(t, err)

	return pubSub.Publisher().(*publisher)
}

func getEventRetryHandler(t *testing.T, l *logrusx.Logger, conf *pubsubx.Config, group messagex.ConsumerGroup, opts *pubsubx.SubscriberOptions) *eventRetryHandler {
	t.Helper()
	pubSub, err := NewPubSub(l, conf, nil)
	require.NoError(t, err)

	return pubSub.eventRetryHandler(group, opts)
}

func getPoisonQueueHandler(t *testing.T, l *logrusx.Logger, conf *pubsubx.Config) *poisonQueueHandler {
	t.Helper()
	pubSub, err := NewPubSub(l, conf, nil)
	require.NoError(t, err)

	return pubSub.PoisonQueueHandler().(*poisonQueueHandler)
}
