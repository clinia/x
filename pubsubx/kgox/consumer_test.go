package kgox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/goleak"
)

func TestConsumer_Subscribe_Handling(t *testing.T) {
	consumer_Subscribe_Handling_test(t, false)
}

func TestConsumer_Subscribe_Handling_async(t *testing.T) {
	consumer_Subscribe_Handling_test(t, true)
}

func TestConsumerLifecycle(t *testing.T) {
	const kafkaTimeouts = 5 * time.Second
	const msgFailureTimeout = kafkaTimeouts * 3
	l := getLogger()
	m := noop.NewMeterProvider().Meter("test")
	tf := newProxyFixture(t)
	tf.EnableAll()
	t.Cleanup(tf.EnableAll)
	config := getPubsubConfig(t, false)
	opts := &pubsubx.SubscriberOptions{MaxBatchSize: 10, MaxTopicRetryCount: 3, EnableAsyncExecution: false, RebalanceTimeout: kafkaTimeouts, DialTimeout: kafkaTimeouts}
	pubSub, err := NewPubSub(l, config, nil)
	require.NoError(t, err)
	pqh := pubSub.PoisonQueueHandler().(*poisonQueueHandler)
	defer pubSub.Close()
	getWriteClient := func(t *testing.T) *kgo.Client {
		t.Helper()
		wc, err := kgo.NewClient(
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
			kgo.WithLogger(&pubsubLogger{l: l}),
		)
		require.NoError(t, err)
		t.Cleanup(wc.Close)
		return wc
	}
	sendMessage := func(t *testing.T, ctx context.Context, wc *kgo.Client, topic messagex.Topic, msg *messagex.Message) {
		t.Helper()

		rec, err := defaultMarshaler.Marshal(ctx, msg, topic.TopicName(config.Scope))
		require.NoError(t, err)

		r := wc.ProduceSync(context.Background(), rec)
		require.NoError(t, r.FirstErr())
	}

	t.Run("should gracefully cleanup consumer on network failure and allow resubscribing", func(t *testing.T) {
		glopt := goleak.IgnoreCurrent()
		defer goleak.VerifyNone(t, glopt)
		wc := getWriteClient(t)
		defer wc.Close()
		group, topics := getRandomGroupTopics(t, 1)
		createTopic(t, config, topics[0])
		cg := messagex.ConsumerGroup(group)
		erh := pubSub.eventRetryHandler(cg, nil)
		c, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		require.NoError(t, err)
		defer c.Close()
		subCtx := context.Background()
		msgBuf := make(chan *messagex.Message)
		handler := func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
			for _, msg := range msgs {
				msgBuf <- msg
			}
			return nil, nil
		}
		err = c.Subscribe(subCtx, pubsubx.Handlers{
			topics[0]: handler,
		})
		require.NoError(t, err, "no error on first Subscribe")
		assert.NoError(t, c.Health())
		sendMessage(t, context.Background(), wc, topics[0], messagex.NewMessage([]byte("test_msg_1")))
		ticker := time.After(15 * time.Second)
		select {
		case <-ticker:
			t.Fatal("failed to receive first message in first subscriber")
		case <-msgBuf:
		}
		tf.DisableAll()
		defer tf.EnableAll()
		c.wg.Wait()
		require.Error(t, c.Health())
		subCtx = context.Background()
		tf.EnableAll()
		err = c.Subscribe(subCtx, pubsubx.Handlers{
			topics[0]: handler,
		})
		require.NoError(t, err, "no error on second Subscribe")
		assert.NoError(t, c.Health())
		sendMessage(t, context.Background(), wc, topics[0], messagex.NewMessage([]byte("test_msg_2")))
		select {
		case <-time.After(msgFailureTimeout):
			t.Fatal("failed to receive first message in second subscriber")
		case <-msgBuf:
		}
	})

	t.Run("should gracefully cleanup consumer on network failure while closing and allow resubscribing", func(t *testing.T) {
		glopt := goleak.IgnoreCurrent()
		defer goleak.VerifyNone(t, glopt)
		wc := getWriteClient(t)
		defer wc.Close()
		group, topics := getRandomGroupTopics(t, 1)
		createTopic(t, config, topics[0])
		cg := messagex.ConsumerGroup(group)
		erh := pubSub.eventRetryHandler(cg, nil)
		c, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		require.NoError(t, err)
		defer c.Close()
		subCtx := context.Background()
		msgBuf := make(chan *messagex.Message)
		handler := func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
			for _, msg := range msgs {
				msgBuf <- msg
			}
			return nil, nil
		}
		err = c.Subscribe(subCtx, pubsubx.Handlers{
			topics[0]: handler,
		})
		require.NoError(t, err, "no error on first Subscribe")
		assert.NoError(t, c.Health())
		sendMessage(t, context.Background(), wc, topics[0], messagex.NewMessage([]byte("test_msg_1")))
		ticker := time.After(15 * time.Second)
		select {
		case <-ticker:
			t.Fatal("failed to receive first message in first subscriber")
		case <-msgBuf:
		}
		c.loopProcessingWg.Wait()
		tf.DisableAll()
		t.Cleanup(tf.EnableAll)
		c.wg.Wait()

		c.Close()
		require.Error(t, c.Health())
		subCtx = context.Background()
		tf.EnableAll()
		start := time.Now()
		err = c.Subscribe(subCtx, pubsubx.Handlers{
			topics[0]: handler,
		})
		require.NoError(t, err, "no error on second Subscribe")
		assert.NoError(t, c.Health())
		sendMessage(t, context.Background(), wc, topics[0], messagex.NewMessage([]byte("test_msg_2")))
		select {
		case <-time.After(msgFailureTimeout):
			t.Fatal("failed to receive first message in second subscriber")
		case <-msgBuf:
			t.Logf("received message after %s", time.Since(start))
		}
	})

	t.Run("should allow to close a subscribe and reuse the subscriber", func(t *testing.T) {
		glopt := goleak.IgnoreCurrent()
		defer goleak.VerifyNone(t, glopt)
		wc := getWriteClient(t)
		defer wc.Close()
		group, topics := getRandomGroupTopics(t, 1)
		createTopic(t, config, topics[0])
		cg := messagex.ConsumerGroup(group)
		erh := pubSub.eventRetryHandler(cg, nil)
		c, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		require.NoError(t, err)
		defer c.Close()
		subCtx := context.Background()
		msgBuf := make(chan *messagex.Message)
		handler := func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
			for _, msg := range msgs {
				msgBuf <- msg
			}
			return nil, nil
		}
		err = c.Subscribe(subCtx, pubsubx.Handlers{
			topics[0]: handler,
		})
		require.NoError(t, err, "no error on first Subscribe")
		assert.NoError(t, c.Health())
		sendMessage(t, context.Background(), wc, topics[0], messagex.NewMessage([]byte("test_msg_1")))
		ticker := time.After(15 * time.Second)
		select {
		case <-ticker:
			t.Fatal("failed to receive first message in first subscriber")
		case <-msgBuf:
		}
		c.Close()
		require.Error(t, c.Health())
		subCtx = context.Background()
		start := time.Now()
		err = c.Subscribe(subCtx, pubsubx.Handlers{
			topics[0]: handler,
		})
		require.NoError(t, err, "no error on second Subscribe")
		assert.NoError(t, c.Health())
		sendMessage(t, context.Background(), wc, topics[0], messagex.NewMessage([]byte("test_msg_2")))
		select {
		case <-time.After(msgFailureTimeout):
			t.Fatal("failed to receive first message in second subscriber")
		case <-msgBuf:
			t.Logf("received message after %s", time.Since(start))
		}
	})
}

func consumer_Subscribe_Handling_test(t *testing.T, eae bool) {
	l := getLogger()
	tf := newProxyFixture(t)
	tf.EnableAll()
	t.Cleanup(tf.EnableAll)
	config := getPubsubConfig(t, true)
	opts := &pubsubx.SubscriberOptions{MaxBatchSize: 10, MaxTopicRetryCount: 3, EnableAsyncExecution: eae, RebalanceTimeout: 1 * time.Second, DialTimeout: 1 * time.Second}
	pqh := getPoisonQueueHandler(t, l, config)
	m := noop.NewMeterProvider().Meter("test")

	getWriteClient := func(t *testing.T) *kgo.Client {
		t.Helper()
		wc, err := kgo.NewClient(
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
		)
		require.NoError(t, err)
		t.Cleanup(wc.Close)
		return wc
	}

	sendMessage := func(t *testing.T, ctx context.Context, wc *kgo.Client, topic messagex.Topic, msg *messagex.Message) {
		t.Helper()

		rec, err := defaultMarshaler.Marshal(ctx, msg, topic.TopicName(config.Scope))
		require.NoError(t, err)

		r := wc.ProduceSync(context.Background(), rec)
		require.NoError(t, r.FirstErr())
	}

	t.Run("should push back messages on retryable error", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 1)
		createTopic(t, config, topics[0])
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		consumer, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		if err != nil {
			t.Fatalf("failed to create consumer: %v", err)
		}
		t.Cleanup(func() { consumer.Close() })
		wClient := getWriteClient(t)

		ctx := context.Background()
		recv := make(chan string, 3)

		topicHandlers := pubsubx.Handlers{
			topics[0]: func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				errs := make([]error, len(msgs))
				for i, msg := range msgs {
					recv <- msg.Metadata[messagex.RetryCountHeaderKey]
					errs[i] = errorx.NewRetryableError(errors.New("Retry Me"))
				}
				return errs, nil
			},
		}
		err = consumer.Subscribe(ctx, topicHandlers)
		require.NoError(t, err)

		expectedMsg := messagex.NewMessage([]byte("test"))
		sendMessage(t, ctx, wClient, topics[0], expectedMsg)

		ctx, cancel := context.WithTimeout(ctx, defaultAssertTimeout)
		t.Cleanup(cancel)
		actual := make([]string, 0, 3)
	loop:
		for range cap(recv) {
			select {
			case <-ctx.Done():
				break loop
			case headerVal := <-recv:
				actual = append(actual, headerVal)
			}
		}

		require.Equal(t, []string{"0", "1", "2"}, actual)
	})

	t.Run("should not push back messages on non retryable error", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 1)
		createTopic(t, config, topics[0])
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		consumer, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		if err != nil {
			t.Fatalf("failed to create consumer: %v", err)
		}
		t.Cleanup(func() { consumer.Close() })
		wClient := getWriteClient(t)

		ctx := context.Background()
		recv := make(chan string, 3)
		topicHandlers := pubsubx.Handlers{
			topics[0]: func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				errs := make([]error, len(msgs))
				for i, msg := range msgs {
					recv <- msg.Metadata[messagex.RetryCountHeaderKey]
					errs[i] = errors.New("Do not retry Me")
				}
				return errs, nil
			},
		}
		err = consumer.Subscribe(ctx, topicHandlers)
		require.NoError(t, err)

		expectedMsg := messagex.NewMessage([]byte("test"))
		sendMessage(t, ctx, wClient, topics[0], expectedMsg)

		ctx, cancel := context.WithTimeout(ctx, defaultExpectedNoReceiveTimeout)
		t.Cleanup(cancel)
		actual := make([]string, 0, 3)
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case headerVal := <-recv:
				actual = append(actual, headerVal)
				if len(actual) > 1 {
					break loop
				}
			}
		}

		require.Equal(t, []string{"0"}, actual)
	})

	t.Run("should not commit the records if the connection with the pubsub service fails", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 1)
		createTopic(t, config, topics[0])
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		ctx := context.Background()
		consumer, err := newConsumer(ctx, l, nil, config, cg, topics, opts, erh, pqh, m)
		require.NoError(t, err)
		wClient := getWriteClient(t)
		cMu := sync.Mutex{}
		shouldClose := false
		waitForClose := make(chan struct{})
		wg := sync.WaitGroup{}
		require.NoError(t, consumer.Subscribe(ctx, pubsubx.Handlers{
			topics[0]: func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				defer func() {
					for range msgs {
						wg.Done()
					}
				}()
				cMu.Lock()
				defer cMu.Unlock()
				if shouldClose {
					<-waitForClose
				}
				return nil, nil
			},
		}))
		expectedMsg := messagex.NewMessage([]byte("test"))
		expectedMsg2 := messagex.NewMessage([]byte("test2"))

		closeConsumer1 := func() {
			cMu.Lock()
			defer cMu.Unlock()
			shouldClose = true
		}

		wg.Add(1)
		sendMessage(t, ctx, wClient, topics[0], expectedMsg)
		// This makes sure the first message has been processed and commited
		wg.Wait()
		wg.Add(1)
		closeConsumer1()
		sendMessage(t, ctx, wClient, topics[0], expectedMsg2)
		tf.DisableAll()
		waitForClose <- struct{}{}

		// This waits for the failed execution
		wg.Wait()

		consumer.Close()

		// We reenable the proxies to make sure we can consume again with the new consumer
		tf.EnableAll()

		ctx = context.Background()
		consumer2, err := newConsumer(ctx, l, nil, config, cg, topics, opts, erh, pqh, m)
		require.NoError(t, err)
		t.Cleanup(func() { consumer2.Close() })

		receivedMsg := make(chan string, 1)
		err = consumer2.Subscribe(ctx, pubsubx.Handlers{
			topics[0]: func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				if len(msgs) != 1 {
					t.Fail()
				}
				receivedMsg <- string(msgs[0].Payload)
				return make([]error, 0), nil
			},
		})
		require.NoError(t, err)
		select {
		case <-time.After(defaultAssertTimeout):
			t.Fail()
		case msg := <-receivedMsg:
			require.Equal(t, "test2", msg)
		}
	})

	t.Run("should push back messages on retryable error for complete error", func(t *testing.T) {
		t.Skipf("test takes at least 45 seconds to run, skipping to reduce test suite time, please manually trigger this test")
		group, topics := getRandomGroupTopics(t, 1)
		createTopic(t, config, topics[0])
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		consumer, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		if err != nil {
			t.Fatalf("failed to create consumer: %v", err)
		}
		t.Cleanup(func() { consumer.Close() })
		wClient := getWriteClient(t)
		ctx := context.Background()

		headerResult := map[string]int{
			"0": 0,
			"1": 0,
			"2": 0,
		}

		mu := sync.Mutex{}
		topicHandlers := pubsubx.Handlers{
			topics[0]: func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				mu.Lock()
				defer mu.Unlock()
				errs := make([]error, len(msgs))
				for i, msg := range msgs {
					headerResult[msg.Metadata[messagex.RetryCountHeaderKey]] += 1
					errs[i] = errorx.NewRetryableError(errors.New("Retry Me"))
				}
				return errs, errs[0]
			},
		}
		err = consumer.Subscribe(context.Background(), topicHandlers)
		require.NoError(t, err)

		expectedMsg := messagex.NewMessage([]byte("test"))
		sendMessage(t, ctx, wClient, topics[0], expectedMsg)

		time.Sleep(15 * time.Second)
		err = consumer.Subscribe(context.Background(), topicHandlers)
		require.NoError(t, err)

		time.Sleep(15 * time.Second)
		err = consumer.Subscribe(context.Background(), topicHandlers)
		require.NoError(t, err)

		time.Sleep(15 * time.Second)
		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, 4, headerResult["0"])
		assert.Equal(t, 4, headerResult["1"])
		assert.Equal(t, 4, headerResult["2"])
	})
}

func TestConsumer_Subscribe_Concurrency(t *testing.T) {
	consumer_Subscribe_Concurrency_test(t, false)
}

func TestConsumer_Subscribe_Concurrency_async(t *testing.T) {
	consumer_Subscribe_Concurrency_test(t, true)
}

func consumer_Subscribe_Concurrency_test(t *testing.T, eae bool) {
	l := getLogger()
	tf := newProxyFixture(t)
	tf.EnableAll()
	t.Cleanup(tf.EnableAll)
	config := getPubsubConfig(t, false)
	pqh := getPoisonQueueHandler(t, l, config)
	m := noop.NewMeterProvider().Meter("test")
	opts := &pubsubx.SubscriberOptions{MaxBatchSize: 10, EnableAsyncExecution: eae, RebalanceTimeout: 1 * time.Second, DialTimeout: 1 * time.Second}

	getWriteClient := func(t *testing.T) *kgo.Client {
		t.Helper()
		wc, err := kgo.NewClient(
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
		)
		require.NoError(t, err)
		t.Cleanup(wc.Close)
		return wc
	}

	sendMessage := func(t *testing.T, ctx context.Context, wc *kgo.Client, topic messagex.Topic, msg *messagex.Message) {
		t.Helper()

		rec, err := defaultMarshaler.Marshal(ctx, msg, topic.TopicName(config.Scope))
		require.NoError(t, err)

		r := wc.ProduceSync(context.Background(), rec)
		require.NoError(t, r.FirstErr())
	}

	t.Run("should return an error if there are missing handlers", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 3)
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		consumer, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		if err != nil {
			t.Fatalf("failed to create consumer: %v", err)
		}
		t.Cleanup(func() { consumer.Close() })

		noopHandler := func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
			return nil, nil
		}
		ctx := context.Background()
		err = consumer.Subscribe(ctx, pubsubx.Handlers{
			topics[0]: noopHandler,
			topics[1]: noopHandler,
			// Missing handler for topics[2]
			// topics[2]: noopHandler,
		})
		require.Error(t, err)

		err = consumer.Subscribe(ctx, pubsubx.Handlers{
			topics[0]: noopHandler,
			topics[1]: noopHandler,
			topics[2]: noopHandler,
		})
		require.NoError(t, err)
	})

	t.Run("should handle concurrent Subscribe calls and close", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 1)
		testTopic := topics[0]
		createTopic(t, config, testTopic)

		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		consumer, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		if err != nil {
			t.Fatalf("failed to create consumer: %v", err)
		}
		t.Cleanup(func() { consumer.Close() })

		ctx := context.Background()
		topicHandlers := pubsubx.Handlers{
			testTopic: func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				return nil, nil
			},
		}

		var wg sync.WaitGroup
		wg.Add(2)

		var firstErr, secondErr error

		go func() {
			defer wg.Done()
			firstErr = consumer.Subscribe(ctx, topicHandlers)
		}()

		go func() {
			defer wg.Done()
			secondErr = consumer.Subscribe(ctx, topicHandlers)
		}()

		wg.Wait()

		if firstErr == nil && secondErr == nil {
			t.Fatalf("expected one of the Subscribe calls to fail, but both succeeded")
		}

		if firstErr != nil && secondErr != nil {
			t.Fatalf("expected one of the Subscribe calls to succeed, but both failed")
		}

		if consumer.cancel == nil {
			t.Fatalf("expected consumer to be subscribed, but it is not")
		}

		// After we cancel the consumer, we should be able to subscribe again
		err = consumer.Close()
		require.NoError(t, err)

		err = consumer.Subscribe(ctx, topicHandlers)
		require.NoError(t, err)
	})

	t.Run("should be able to consume messages again after closing", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 1)
		testTopic := topics[0]
		wClient := getWriteClient(t)
		createTopic(t, config, testTopic)

		receivedMsgs := make(chan *messagex.Message, 10)
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		consumer, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		if err != nil {
			t.Fatalf("failed to create consumer: %v", err)
		}

		ctx := context.Background()
		topicHandlers := pubsubx.Handlers{
			testTopic: func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				for _, msg := range msgs {
					receivedMsgs <- msg
				}

				return nil, nil
			},
		}

		err = consumer.Subscribe(ctx, topicHandlers)
		require.NoError(t, err)

		// Produce a message
		expectedMsg := messagex.NewMessage([]byte("test"))
		sendMessage(t, ctx, wClient, testTopic, expectedMsg)

		// Wait for the message to be consumed
		select {
		case <-time.After(defaultExpectedReceiveTimeout):
			t.Fatalf("timed out waiting for message to be consumed")
		case msg := <-receivedMsgs:
			assert.Equal(t, expectedMsg, msg)
		}

		// Close the consumer
		err = consumer.Close()
		require.NoError(t, err)

		// Produce another message
		expectedMsg2 := messagex.NewMessage([]byte("test2"))
		sendMessage(t, ctx, wClient, testTopic, expectedMsg2)

		// We should not receive the message since the consumer is closed
		select {
		case <-time.After(defaultExpectedNoReceiveTimeout):
			t.Log("no message consumed, all good")
		case <-receivedMsgs:
			t.Fatalf("expected no message to be consumed")
		}

		// Re-open the consumer
		err = consumer.Subscribe(ctx, topicHandlers)
		require.NoError(t, err)

		// Wait for the message to be consumed
		select {
		case <-time.After(defaultExpectedReceiveTimeout):
			t.Fatalf("timed out waiting for message to be consumed")
		case msg := <-receivedMsgs:
			assert.Equal(t, expectedMsg2, msg)
		}

		// Close the consumer
		err = consumer.Close()
		require.NoError(t, err)

		// If we subscribe with another consumer group, we should receive both messages
		anotherGroup, _ := getRandomGroupTopics(t, 1)
		acg := messagex.ConsumerGroup(anotherGroup)
		aerh := getEventRetryHandler(t, l, config, acg, nil)
		anotherConsumer, err := newConsumer(context.Background(), l, nil, config, acg, topics, opts, aerh, pqh, m)
		t.Cleanup(func() {
			anotherConsumer.Close()
		})
		require.NoError(t, err)

		anotherReceivedMsgs := make(chan *messagex.Message, 10)
		anotherTopicHandlers := pubsubx.Handlers{
			testTopic: func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				for _, msg := range msgs {
					anotherReceivedMsgs <- msg
				}

				return nil, nil
			},
		}

		err = anotherConsumer.Subscribe(ctx, anotherTopicHandlers)
		require.NoError(t, err)

		// Wait for the messages to be consumed
		expectedMsgs := []*messagex.Message{expectedMsg, expectedMsg2}
		for _, expectedMsg := range expectedMsgs {
			select {
			case <-time.After(defaultExpectedReceiveTimeout):
				t.Fatalf("timed out waiting for message to be consumed")
			case msg := <-anotherReceivedMsgs:
				assert.Equal(t, expectedMsg, msg)
			}
		}
	})

	t.Run("should be able to consume multiple topics and abort from one topic failure", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 3)
		counter := make([]atomic.Int32, len(topics))
		wClient := getWriteClient(t)
		hs := make(pubsubx.Handlers, len(topics))
		for i, topic := range topics {
			createTopic(t, config, topic)
			hs[topic] = func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				counter[i].Add(int32(len(msgs)))
				var err error = nil
				errs := make([]error, len(msgs))
				for _, msg := range msgs {
					if msg != nil && string(msg.Payload) == "abort" {
						err = pubsubx.AbortSubscribeError()
					}
					errs[i] = errors.New("abort error on this message")
				}
				return errs, err
			}
		}
		ctx := context.Background()
		for i := range 10 * len(topics) {
			expectedMsg := messagex.NewMessage([]byte("test" + strconv.Itoa(i)))
			sendMessage(t, ctx, wClient, topics[i%len(topics)], expectedMsg)
		}
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		c, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		require.NoError(t, err)
		require.NoError(t, c.Subscribe(ctx, hs))
		t.Cleanup(func() { c.Close() })

		expected := []int32{10, 10, 10}
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			for i, topic := range topics {
				assert.Equal(c, expected[i], counter[i].Load(),
					fmt.Sprintf("topic '%s' did not receive the expected msg count", topic))
			}
		}, defaultAssertTimeout, 250*time.Millisecond)

		for i := range len(topics) {
			expectedMsg := messagex.NewMessage([]byte("test" + strconv.Itoa(i)))
			if i == len(topics)/2 {
				expectedMsg = messagex.NewMessage([]byte("abort"))
			}
			sendMessage(t, ctx, wClient, topics[i%len(topics)], expectedMsg)
		}

		// This will wait until the consumer closed after receiving the AbortSubscribeError
		t.Logf("waiting for consumer to close")
		c.wg.Wait()

		assert.Equal(t, expected[len(expected)/2]+4, counter[len(counter)/2].Load())
	})

	t.Run("should be able to retry up to the max retry attempts", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 1)
		testTopic := topics[0]
		wClient := getWriteClient(t)
		createTopic(t, config, testTopic)

		// Buffer to intercept logs
		buf := newConcurrentBuffer(t)
		l.Entry.Logger.SetOutput(buf)

		receivedMsgs := make(chan *messagex.Message, 10)
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		consumer, err := newConsumer(context.Background(), l, nil, config, cg, topics, opts, erh, pqh, m)
		if err != nil {
			t.Fatalf("failed to create consumer: %v", err)
		}
		t.Cleanup(func() { consumer.Close() })

		ctx := context.Background()
		shouldFail := make(chan bool)
		shouldPanic := make(chan bool)

		topicHandlers := pubsubx.Handlers{
			testTopic: func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				for _, msg := range msgs {
					if msg.Metadata[messagex.RetryCountHeaderKey] == "0" {
						receivedMsgs <- msg
					}
				}

				select {
				case <-shouldPanic:
					panic("failed to process message")
				case shouldFail := <-shouldFail:
					if shouldFail {
						return nil, assert.AnError
					}

					return nil, nil
				}
			},
		}

		err = consumer.Subscribe(ctx, topicHandlers)
		require.NoError(t, err)

		// Produce a message
		expectedMsg := messagex.NewMessage([]byte("test"))
		sendMessage(t, ctx, wClient, testTopic, expectedMsg)

		// We should retry up to the max retry attempts
		for i := range maxRetryCount + 1 {
			if i == maxRetryCount {
				// This is the last attempt, we should not fail and consume the message properly
				shouldFail <- false
			} else {
				shouldFail <- true
			}

			// Wait for the message to be consumed
			select {
			case <-time.After(defaultExpectedReceiveTimeout):
				t.Fatalf("timed out waiting for message to be consumed")
			case msg := <-receivedMsgs:
				assert.Equal(t, expectedMsg, msg)
			}
		}

		// Consumer should still be alive
		assert.NoError(t, consumer.Health())

		expectedMsg = messagex.NewMessage([]byte("test2"))
		sendMessage(t, ctx, wClient, testTopic, expectedMsg)

		shouldFail <- false
		// Wait for the message to be consumed
		select {
		case <-time.After(defaultExpectedReceiveTimeout):
			t.Fatalf("timed out waiting for message to be consumed")
		case msg := <-receivedMsgs:
			assert.Equal(t, expectedMsg, msg)
		}

		// The consumer should stop if we fail up to the max retry attempts
		expectedMsg = messagex.NewMessage([]byte("test3"))
		sendMessage(t, ctx, wClient, testTopic, expectedMsg)

		for i := range maxRetryCount + 1 {
			if i == 0 {
				buf.b.Reset()
				shouldPanic <- true
			} else {
				shouldFail <- true
			}

			// Wait for the message to be consumed
			select {
			case <-time.After(defaultExpectedReceiveTimeout):
				t.Fatalf("timed out waiting for message to be consumed")
			case msg := <-receivedMsgs:
				assert.Equal(t, expectedMsg, msg)
			}

			if i == 0 {
				// Expect stack trace to be logged
				assert.EventuallyWithT(t, func(t *assert.CollectT) {
					str := buf.String()
					containsPanic := strings.Contains(str, "panic while handling messages")
					containsStackTrace := strings.Contains(str, "consumer_test.go:")
					assert.True(t, containsPanic, "expected panic message to be logged")
					assert.True(t, containsStackTrace, "expected stack trace to be logged")
				}, defaultAssertTimeout, 100*time.Millisecond)
			}
		}

		// The consumer should be closed
		assert.EventuallyWithT(t, func(t *assert.CollectT) {
			expectErr := consumer.Health()
			assert.Error(t, expectErr)
		}, defaultAssertTimeout, 100*time.Millisecond)
	})
}

type concurrentBuffer struct {
	b *bytes.Buffer
	m sync.RWMutex
	t *testing.T
}

func newConcurrentBuffer(t *testing.T) *concurrentBuffer {
	return &concurrentBuffer{
		b: new(bytes.Buffer),
		t: t,
	}
}

func (c *concurrentBuffer) Write(p []byte) (n int, err error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.b.Write(p)
}

func (c *concurrentBuffer) String() string {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.b.String()
}
