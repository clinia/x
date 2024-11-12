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
	"github.com/samber/lo"
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
	name := strings.ReplaceAll(t.Name(), "/", "-")[:8]
	rand := ksuid.New().String()
	Group = fmt.Sprintf("%s-%s-group", name, rand)
	for i := 0; i < count; i++ {
		Topics = append(Topics, messagex.Topic(fmt.Sprintf("%s-%s-%d", name, rand, i)))
	}

	return
}

func getPubsubConfig(t *testing.T) *pubsubx.Config {
	t.Helper()
	kafkaURLs := []string{"localhost:9091", "localhost:9092", "localhost:9093", "localhost:9094", "localhost:9095"}
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
		TopicRetry: true,
	}
}

func createTopic(t *testing.T, conf *pubsubx.Config, topic messagex.Topic, consumerGroup messagex.ConsumerGroup) {
	topics := []messagex.Topic{topic, topic.GenerateRetryTopic(consumerGroup)}
	client, err := kgo.NewClient(kgo.SeedBrokers(conf.Providers.Kafka.Brokers...))
	require.NoError(t, err)

	admCl := kadm.NewClient(client)
	deleteTopic := func() error {
		// Delete the topic
		res, err := admCl.DeleteTopics(context.Background(), lo.Map(topics, func(t messagex.Topic, _ int) string {
			return t.TopicName(conf.Scope)
		})...)
		return errors.Join(err, res.Error())
	}

	// Ensure the topic is deleted
	deleteTopic()

	// Create the topic
	res, err := admCl.CreateTopics(context.Background(), 1, 1, map[string]*string{}, lo.Map(topics, func(t messagex.Topic, _ int) string {
		return t.TopicName(conf.Scope)
	})...)
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
