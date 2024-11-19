package kgox

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/clinia/x/assertx"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestPublishMessagesToPoisonQueue(t *testing.T) {
	l := logrusx.New("test", "")

	t.Run("should not do anything if the messages list is empty", func(t *testing.T) {
		config := getPubSubConfigWithCustomPoisonQueue(t, true, "pq-test-1")
		pqTopic := messagex.TopicFromName(config.PoisonQueue.TopicName).TopicName(config.Scope)
		kopts := []kgo.Opt{
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
			kgo.ConsumerGroup("poison-queue-test-group-1"),
			kgo.ConsumeTopics(pqTopic),
		}
		testClient, err := kgo.NewClient(kopts...)
		require.NoError(t, err)
		defer testClient.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		records := make([]*kgo.Record, 0, 1)
		pqh := getPoisonQueueHandler(t, l, config)
		admCl := kadm.NewClient(testClient)
		defer func() error {
			// Delete the topic
			res, err := admCl.DeleteTopics(context.Background(), pqTopic)
			return errors.Join(err, res.Error())
		}()
		err = pqh.PublishMessagesToPoisonQueue(ctx, "failed-topic", "failed-group", []error{}, []*messagex.Message{})
		assert.NoError(t, err)
		mut := sync.Mutex{}
		go func() {
			for {
				fetches := testClient.PollFetches(ctx)
				if fetches == nil {
					return
				}
				fetches.EachTopic(func(tp kgo.FetchTopic) {
					mut.Lock()
					defer mut.Unlock()
					rs := tp.Records()
					records = append(records, rs...)
				})
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
		assert.Never(t, func() bool {
			mut.Lock()
			defer mut.Unlock()
			return len(records) > 0
		}, 5*time.Second, 500*time.Millisecond)
	})

	t.Run("should publish to the queue with an empty error", func(t *testing.T) {
		config := getPubSubConfigWithCustomPoisonQueue(t, true, "pq-test-2")
		pqTopic := messagex.TopicFromName(config.PoisonQueue.TopicName).TopicName(config.Scope)
		kopts := []kgo.Opt{
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
			kgo.ConsumerGroup("poison-queue-test-group-2"),
			kgo.ConsumeTopics(pqTopic),
		}
		testClient, err := kgo.NewClient(kopts...)
		require.NoError(t, err)
		defer testClient.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		require.NoError(t, err)
		pqh := getPoisonQueueHandler(t, l, config)
		admCl := kadm.NewClient(testClient)
		defer func() error {
			// Delete the topic
			res, err := admCl.DeleteTopics(context.Background(), pqTopic)
			return errors.Join(err, res.Error())
		}()

		failTopicName := "failed-topic"
		failGroupName := "failed-group"
		err = pqh.PublishMessagesToPoisonQueue(ctx, failTopicName, messagex.ConsumerGroup(failGroupName), []error{}, []*messagex.Message{
			messagex.NewMessage([]byte("test")),
		})
		assert.NoError(t, err)
		mut := sync.Mutex{}
		records := make([]*kgo.Record, 0, 1)
		go func() {
			for {
				fetches := testClient.PollFetches(ctx)
				if fetches == nil {
					return
				}
				fetches.EachTopic(func(tp kgo.FetchTopic) {
					mut.Lock()
					defer mut.Unlock()
					rs := tp.Records()
					records = append(records, rs...)
				})
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			mut.Lock()
			defer mut.Unlock()
			assert.Equal(c, 1, len(records))
			if len(records) < 1 {
				return
			}
			keyCheck := map[string]bool{}
			for _, h := range records[0].Headers {
				switch h.Key {
				case originConsumerGroupHeaderKey:
					assert.Equal(c, "failed-group", string(h.Value))
				case originTopicHeaderKey:
					assert.Equal(c, "failed-topic", string(h.Value))
				case originErrorHeaderKey:
					assert.Equal(c, defaultMissingErrorString, string(h.Value))
				}
				keyCheck[h.Key] = true
			}
			assert.True(c, keyCheck[originConsumerGroupHeaderKey])
			assert.True(c, keyCheck[originTopicHeaderKey])
			assert.True(c, keyCheck[originErrorHeaderKey])
		}, 5*time.Second, 500*time.Millisecond)
	})

	t.Run("should publish to the queue with one error and one messages", func(t *testing.T) {
		config := getPubSubConfigWithCustomPoisonQueue(t, true, "pq-test-3")
		pqTopic := messagex.TopicFromName(config.PoisonQueue.TopicName).TopicName(config.Scope)
		kopts := []kgo.Opt{
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
			kgo.ConsumerGroup("poison-queue-test-group-3"),
			kgo.ConsumeTopics(pqTopic),
		}
		testClient, err := kgo.NewClient(kopts...)
		require.NoError(t, err)
		defer testClient.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		pqh := getPoisonQueueHandler(t, l, config)
		admCl := kadm.NewClient(testClient)
		defer func() error {
			// Delete the topic
			res, err := admCl.DeleteTopics(context.Background(), pqTopic)
			return errors.Join(err, res.Error())
		}()
		failTopicName := "failed-topic"
		failGroupName := "failed-group"
		testErrorMessage := "Test-Error"
		err = pqh.PublishMessagesToPoisonQueue(ctx, failTopicName, messagex.ConsumerGroup(failGroupName), []error{errors.New(testErrorMessage)}, []*messagex.Message{
			messagex.NewMessage([]byte("test")),
		})
		assert.NoError(t, err)
		mut := sync.Mutex{}
		records := make([]*kgo.Record, 0, 1)
		go func() {
			for {
				fetches := testClient.PollFetches(ctx)
				if fetches == nil {
					return
				}
				fetches.EachTopic(func(tp kgo.FetchTopic) {
					mut.Lock()
					defer mut.Unlock()
					rs := tp.Records()
					records = append(records, rs...)
				})
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			mut.Lock()
			defer mut.Unlock()
			assert.Equal(c, 1, len(records))
			if len(records) < 1 {
				return
			}
			keyCheck := map[string]bool{}
			for _, h := range records[0].Headers {
				switch h.Key {
				case originConsumerGroupHeaderKey:
					assert.Equal(c, "failed-group", string(h.Value))
				case originTopicHeaderKey:
					assert.Equal(c, "failed-topic", string(h.Value))
				case originErrorHeaderKey:
					assert.Equal(c, testErrorMessage, string(h.Value))
				}
				keyCheck[h.Key] = true
			}
			assert.True(c, keyCheck[originConsumerGroupHeaderKey])
			assert.True(c, keyCheck[originTopicHeaderKey])
			assert.True(c, keyCheck[originErrorHeaderKey])
		}, 5*time.Second, 500*time.Millisecond)
	})

	t.Run("should publish to the queue with generic error when errors and messages mismatch", func(t *testing.T) {
		config := getPubSubConfigWithCustomPoisonQueue(t, true, "pq-test-4")
		pqTopic := messagex.TopicFromName(config.PoisonQueue.TopicName).TopicName(config.Scope)
		kopts := []kgo.Opt{
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
			kgo.ConsumerGroup("poison-queue-test-group-4"),
			kgo.ConsumeTopics(pqTopic),
		}
		testClient, err := kgo.NewClient(kopts...)
		require.NoError(t, err)
		defer testClient.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		pqh := getPoisonQueueHandler(t, l, config)
		admCl := kadm.NewClient(testClient)
		defer func() error {
			// Delete the topic
			res, err := admCl.DeleteTopics(context.Background(), pqTopic)
			return errors.Join(err, res.Error())
		}()
		failTopicName := "failed-topic"
		failGroupName := "failed-group"
		err = pqh.PublishMessagesToPoisonQueue(ctx, failTopicName, messagex.ConsumerGroup(failGroupName), []error{errors.New("Test-Error")}, []*messagex.Message{
			messagex.NewMessage([]byte("test")),
			messagex.NewMessage([]byte("test2")),
		})
		assert.NoError(t, err)
		mut := sync.Mutex{}
		records := make([]*kgo.Record, 0, 1)
		go func() {
			for {
				fetches := testClient.PollFetches(ctx)
				if fetches == nil {
					return
				}
				fetches.EachTopic(func(tp kgo.FetchTopic) {
					mut.Lock()
					defer mut.Unlock()
					rs := tp.Records()
					records = append(records, rs...)
				})
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			mut.Lock()
			defer mut.Unlock()
			assert.Equal(c, 2, len(records))
			for _, r := range records {
				keyCheck := map[string]bool{}
				for _, h := range r.Headers {
					switch h.Key {
					case originConsumerGroupHeaderKey:
						assert.Equal(c, "failed-group", string(h.Value))
					case originTopicHeaderKey:
						assert.Equal(c, "failed-topic", string(h.Value))
					case originErrorHeaderKey:
						assert.Equal(c, defaultMissingErrorString, string(h.Value))
					}
					keyCheck[h.Key] = true
				}
				assert.True(c, keyCheck[originConsumerGroupHeaderKey])
				assert.True(c, keyCheck[originTopicHeaderKey])
				assert.True(c, keyCheck[originErrorHeaderKey])

			}
		}, 5*time.Second, 500*time.Millisecond)
	})

	t.Run("should publish to the queue with all matching errors and messages", func(t *testing.T) {
		config := getPubSubConfigWithCustomPoisonQueue(t, true, "pq-test-5")
		pqTopic := messagex.TopicFromName(config.PoisonQueue.TopicName).TopicName(config.Scope)
		kopts := []kgo.Opt{
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
			kgo.ConsumerGroup("poison-queue-test-group-5"),
			kgo.ConsumeTopics(pqTopic),
		}
		testClient, err := kgo.NewClient(kopts...)
		require.NoError(t, err)
		defer testClient.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		admCl := kadm.NewClient(testClient)
		defer func() error {
			// Delete the topic
			res, err := admCl.DeleteTopics(context.Background(), pqTopic)
			return errors.Join(err, res.Error())
		}()
		pqh := getPoisonQueueHandler(t, l, config)
		failTopicName := "failed-topic"
		failGroupName := "failed-group"
		testErrorMessage := "Test-Error"
		testPayload := "test"
		msgErrors := []error{
			errors.New(testErrorMessage + "1"),
			errors.New(testErrorMessage + "2"),
			errors.New(testErrorMessage + "3"),
		}
		msgs := []*messagex.Message{
			messagex.NewMessage([]byte(testPayload + "1")),
			messagex.NewMessage([]byte(testPayload + "1")),
			messagex.NewMessage([]byte(testPayload + "3")),
		}
		type inlinePair struct {
			msg *messagex.Message
			err error
		}
		msgMapper := map[string]inlinePair{}
		for i, m := range msgs {
			msgMapper[m.ID] = inlinePair{m, msgErrors[i]}
		}
		err = pqh.PublishMessagesToPoisonQueue(ctx, failTopicName, messagex.ConsumerGroup(failGroupName),
			msgErrors,
			msgs)
		assert.NoError(t, err)
		mut := sync.Mutex{}
		records := make([]*kgo.Record, 0, 1)
		go func() {
			for {
				fetches := testClient.PollFetches(ctx)
				if fetches == nil {
					return
				}
				fetches.EachTopic(func(tp kgo.FetchTopic) {
					mut.Lock()
					defer mut.Unlock()
					rs := tp.Records()
					records = append(records, rs...)
				})
				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			mut.Lock()
			defer mut.Unlock()
			assert.Equal(c, 3, len(records))
			for _, r := range records {
				keyCheck := map[string]bool{}
				var eventPayload kgo.Record
				err = json.Unmarshal(r.Value, &eventPayload)
				assert.NoError(t, err)
				var id string
				for _, h := range eventPayload.Headers {
					if h.Key == messagex.IDHeaderKey {
						id = string(h.Value)
						break
					}
				}
				assert.NotEqual(c, "", id)
				for _, h := range r.Headers {
					switch h.Key {
					case originConsumerGroupHeaderKey:
						assert.Equal(c, "failed-group", string(h.Value))
					case originTopicHeaderKey:
						assert.Equal(c, "failed-topic", string(h.Value))
					case originErrorHeaderKey:
						assert.Equal(c, msgMapper[id].err.Error(), string(h.Value))
					}
					keyCheck[h.Key] = true
				}

				payload, err := defaultMarshaler.Marshal(ctx, msgMapper[id].msg, failTopicName)
				assert.NoError(c, err)
				assert.True(c, keyCheck[originConsumerGroupHeaderKey])
				assert.True(c, keyCheck[originTopicHeaderKey])
				assert.True(c, keyCheck[originErrorHeaderKey])
				assert.Equal(c, payload.Key, eventPayload.Key)
				assert.Equal(c, payload.Value, eventPayload.Value)
				assertx.Equal(c, payload.Headers, eventPayload.Headers)
			}
		}, 5*time.Second, 500*time.Millisecond)
	})
}
