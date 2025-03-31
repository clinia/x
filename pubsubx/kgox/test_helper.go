package kgox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
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
	defaultAssertTimeout            = 10 * time.Second
)

type proxyFixture struct {
	cl      *toxiproxy.Client
	proxies []*toxiproxy.Proxy
}

func newProxyFixture(t *testing.T) *proxyFixture {
	cl := toxiproxy.NewClient("localhost:8474")
	proxies := []*toxiproxy.Proxy{}
	for i := range 3 {
		proxy, err := cl.Proxy(fmt.Sprintf("redpanda_%d", i))
		require.NoError(t, err)
		proxies = append(proxies, proxy)
	}

	return &proxyFixture{cl: cl, proxies: proxies}
}

func (f *proxyFixture) EnableAll() {
	for _, p := range f.proxies {
		p.Enable()
	}
}

func (f *proxyFixture) DisableAll() {
	for _, p := range f.proxies {
		p.Disable()
	}
}

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
	getAdmCl := func() *kadm.Client {
		client, err := kgo.NewClient(kgo.SeedBrokers(conf.Providers.Kafka.Brokers...))
		require.NoError(t, err)

		return kadm.NewClient(client)
	}
	deleteTopic := func() error {
		// Delete the topic
		admCl := getAdmCl()
		defer admCl.Close()
		res, err := admCl.DeleteTopics(context.Background(), topic.TopicName(conf.Scope))
		return errors.Join(err, res.Error())
	}

	// Ensure the topic is deleted
	deleteTopic()

	admCl := getAdmCl()
	defer admCl.Close()
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

func getLogger() *logrusx.Logger {
	l := logrusx.New("Clinia x", "testing")
	l.Entry.Logger.SetOutput(io.Discard)
	return l
}

func getKadmClient(t *testing.T, config *pubsubx.Config, defaultCreateTopicConfigs ...map[string]*string) (*KgoxAdminClient, *kadm.Client, *kgo.Client) {
	wc, err := kgo.NewClient(
		kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
	)
	require.NoError(t, err)
	t.Cleanup(wc.Close)
	var defaultCreateTopicConfigEntries map[string]*string
	if len(defaultCreateTopicConfigs) > 0 {
		defaultCreateTopicConfigEntries = lo.Assign(defaultCreateTopicConfigs...)
	}

	kadmCl := kadm.NewClient(wc)
	return NewPubSubAdminClient(wc, config, defaultCreateTopicConfigEntries), kadmCl, wc
}
