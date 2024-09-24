package kgox

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func getRandomGroupTopics(t *testing.T, count int) (Group string, Topics []messagex.Topic) {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "-")
	rand := ksuid.New().String()
	Group = fmt.Sprintf("%s-%s-group", name, rand)
	for i := 0; i < count; i++ {
		Topics = append(Topics, messagex.Topic(fmt.Sprintf("%s-%s-%d", name, rand, i)))
	}

	return
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
