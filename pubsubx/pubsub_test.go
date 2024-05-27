package pubsubx

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx/kafkax"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func TestSubscribersHealth(t *testing.T) {

	// We need to test the scenarios where a subscriber joins a group that already exists, where the subscribed topic has a max replication of 1
	// This means that when subscriber #2 joins, if subscriber #1 is already consuming the topic, subscriber #2 should not consume the same topic
	// However, whenever subscriber #1 leaves the group, subscriber #2 should start consuming the topic

	t.Run("should be healthy when no nacks", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		testMode = true
		t.Cleanup(func() {
			testMode = false
		})

		scope := strings.ReplaceAll(t.Name(), "/", "_")
		group := "test-group"
		topic := "my-messages"
		require.NoError(t, setupTestTopic(t, fmt.Sprintf("%s.%s", scope, topic)))
		registerConsumerGroupCleanup(t, fmt.Sprintf("%s.%s", scope, group))

		ps, err := setupTestSubscriber(t, scope)
		require.NoError(t, err)

		msgs := []*message.Message{}
		for i := range 10 {
			msgs = append(msgs, message.NewMessage(fmt.Sprintf("%d", i), nil))
		}

		err = ps.Publisher().Publish(ctx, topic, msgs[:4]...)
		require.NoError(t, err)

		sub1, err := ps.Subscriber(group, WithBatchConsumerModel(nil), WithNackResendSleep(kafkax.NoSleep))
		require.NoError(t, err)
		msgs1, err := sub1.Subscribe(ctx, topic)
		require.NoError(t, err)

		deadlineCtx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		t.Cleanup(cancel)
		for range 4 {
			select {
			case <-deadlineCtx.Done():
				t.Fatal("Timeout")
			case msg := <-msgs1:
				msg.Ack()
			}
		}

		// We gucci
		sub2, err := ps.Subscriber(group, WithBatchConsumerModel(nil), WithNackResendSleep(kafkax.NoSleep))
		require.NoError(t, err)
		t.Cleanup(func() {
			err := sub2.Close()
			require.NoError(t, err)
		})

		msgs2, err := sub2.Subscribe(ctx, topic)
		require.NoError(t, err)

		err = ps.Publisher().Publish(ctx, topic, msgs[4:6]...)
		require.NoError(t, err)

		deadlineCtx, cancel = context.WithDeadline(ctx, time.Now().Add(10*time.Second))
		t.Cleanup(cancel)
		for range 2 {
			select {
			case <-deadlineCtx.Done():
				t.Fatal("Timeout")
			case msg := <-msgs1:
				t.Logf("Received message from sub1")
				msg.Ack()
				continue
			case <-msgs2:
				t.Logf("Received message from sub2")
				t.Fatal("Should not have received message since a topic with 1 partition should only be consumed by one subscriber within a group")
			}
		}

		// we add a timeout before close to ensure that our MarkMessage from sub 1 has time to be processed
		time.Sleep(1 * time.Second)
		err = sub1.Close()
		require.NoError(t, err)

		err = ps.publisher.Publish(ctx, topic, msgs[6:]...)
		require.NoError(t, err)

		// following messages should be consumed by sub2
		deadlineCtx, cancel = context.WithDeadline(ctx, time.Now().Add(2*time.Second))
		t.Cleanup(cancel)
		for i := range 4 {
			select {
			case msg := <-msgs2:
				t.Logf("Received message #%s from sub2", msg.UUID)
				require.Equal(t, fmt.Sprintf("%d", i+6), msg.UUID)
				msg.Ack()
			case <-deadlineCtx.Done():
				t.Fatal("Timeout")
			}
		}
	})

	t.Run("should still be healthy when nacks have time to commit", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		testMode = true
		t.Cleanup(func() {
			testMode = false
		})

		scope := strings.ReplaceAll(t.Name(), "/", "_")
		group := "test-group"
		topic := "my-messages"
		require.NoError(t, setupTestTopic(t, fmt.Sprintf("%s.%s", scope, topic)))
		registerConsumerGroupCleanup(t, fmt.Sprintf("%s.%s", scope, group))

		ps, err := setupTestSubscriber(t, scope)
		require.NoError(t, err)

		msgs := []*message.Message{}
		for i := range 10 {
			msgs = append(msgs, message.NewMessage(fmt.Sprintf("%d", i), nil))
		}

		err = ps.Publisher().Publish(ctx, topic, msgs[:4]...)
		require.NoError(t, err)

		bCOpts := &BatchConsumerOptions{
			MaxBatchSize: 2,
			MaxWaitTime:  500 * time.Millisecond,
		}

		sub1, err := ps.Subscriber(group, WithBatchConsumerModel(bCOpts))
		require.NoError(t, err)
		msgs1, err := sub1.Subscribe(ctx, topic)
		require.NoError(t, err)

		deadlineCtx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		t.Cleanup(cancel)
		for i := range 4 {
			select {
			case <-deadlineCtx.Done():
				t.Fatal("Timeout")
			case msg := <-msgs1:
				t.Logf("Received message #%d from sub1", i)
				require.Equal(t, fmt.Sprintf("%d", i), msg.UUID)
				msg.Ack()
			}
		}

		// We gucci
		sub2, err := ps.Subscriber(group, WithBatchConsumerModel(bCOpts))
		require.NoError(t, err)
		t.Cleanup(func() {
			err := sub2.Close()
			require.NoError(t, err)
		})

		msgs2, err := sub2.Subscribe(ctx, topic)
		require.NoError(t, err)

		err = ps.Publisher().Publish(ctx, topic, msgs[4:8]...)
		require.NoError(t, err)

		errCh := make(chan error, 1)
		deadlineCtx, cancel = context.WithDeadline(ctx, time.Now().Add(10*time.Second))
		go func(ctx context.Context) {
			// We start trying to consume messages from sub2
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs2:
				errCh <- fmt.Errorf("Received message from sub2: %s", msg.UUID)
			}
		}(deadlineCtx)

		// NACK the first message to ensure sub 2 does not receive message in rebalance
		for range 2 {
			msg := <-msgs1
			msg.Nack()
		}

		for i := range 4 {
			select {
			case <-deadlineCtx.Done():
				t.Fatal("Timeout")
			case err := <-errCh:
				t.Fatal(err)
			case msg := <-msgs1:
				t.Logf("Received message #%d from sub1 (i = %d)", i+4, i)
				require.Equal(t, fmt.Sprintf("%d", i+4), msg.UUID)
				msg.Ack()
				t.Logf("Ack from sub1")
				continue
			}
		}

		cancel()

		// following messages should be consumed by sub2
		deadlineCtx, cancel = context.WithDeadline(ctx, time.Now().Add(2*time.Second))
		t.Cleanup(cancel)
		go func(ctx context.Context) {
			for i := range 2 {
				select {

				case msg := <-msgs2:
					t.Logf("Received message #%s from sub2", msg.UUID)
					if fmt.Sprintf("%d", i+8) != msg.UUID {
						t.Logf("Expected message #%d, got %s", i+8, msg.UUID)
						errCh <- fmt.Errorf("expected message #%d, got %s", i+8, msg.UUID)
						return
					}
					msg.Ack()
				case <-ctx.Done():
					errCh <- fmt.Errorf("timeout")
					return
				}
			}

			close(errCh)
		}(deadlineCtx)

		// we add a timeout before close to ensure that our MarkMessage from sub 1 has time to be processed
		time.Sleep(1 * time.Second)
		err = sub1.Close()
		require.NoError(t, err)

		err = ps.publisher.Publish(ctx, topic, msgs[8:]...)
		require.NoError(t, err)

		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("should still be healthy when acks/nacks do not have time to complete", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		testMode = true
		t.Cleanup(func() {
			testMode = false
		})

		scope := strings.ReplaceAll(fmt.Sprintf("%s_%s", t.Name(), ksuid.New().String()), "/", "_")
		group := "test-group"
		topic := "my-messages"
		require.NoError(t, setupTestTopic(t, fmt.Sprintf("%s.%s", scope, topic)))
		registerConsumerGroupCleanup(t, fmt.Sprintf("%s.%s", scope, group))

		ps, err := setupTestSubscriber(t, scope)
		require.NoError(t, err)

		msgs := []*message.Message{}
		for i := range 10 {
			msgs = append(msgs, message.NewMessage(fmt.Sprintf("%d", i), nil))
		}

		err = ps.Publisher().Publish(ctx, topic, msgs[:4]...)
		require.NoError(t, err)

		bCOpts := &BatchConsumerOptions{
			MaxBatchSize: 2,
			MaxWaitTime:  500 * time.Millisecond,
		}

		sub1, err := ps.Subscriber(group, WithBatchConsumerModel(bCOpts))
		require.NoError(t, err)
		msgs1, err := sub1.Subscribe(ctx, topic)
		require.NoError(t, err)

		deadlineCtx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		t.Cleanup(cancel)
		for i := range 4 {
			select {
			case <-deadlineCtx.Done():
				t.Fatal("Timeout")
			case msg := <-msgs1:
				t.Logf("Received message #%d from sub1", i)
				require.Equal(t, fmt.Sprintf("%d", i), msg.UUID)
				msg.Ack()
			}
		}

		// We gucci
		sub2, err := ps.Subscriber(group, WithBatchConsumerModel(bCOpts))
		require.NoError(t, err)
		t.Cleanup(func() {
			err := sub2.Close()
			require.NoError(t, err)
		})

		msgs2, err := sub2.Subscribe(ctx, topic)
		require.NoError(t, err)

		err = ps.Publisher().Publish(ctx, topic, msgs[4:8]...)
		require.NoError(t, err)

		errCh := make(chan error, 1)
		deadlineCtx, cancel = context.WithDeadline(ctx, time.Now().Add(10*time.Second))
		go func(ctx context.Context) {
			// We start trying to consume messages from sub2
			select {
			case <-ctx.Done():
				t.Logf("Context deadline exceeded")
				return
			case msg := <-msgs2:
				errCh <- fmt.Errorf("received message from sub2: %s", msg.UUID)
			}
		}(deadlineCtx)

		nacked := make(chan bool)
		go func(ctx context.Context) {
			defer close(nacked)
			// We keep nacking messages #4,5,6,7+ from sub1...
			// This is to simulate a scenario where the acks/nacks do not have time to complete since we close the subscriber
			for {
				select {
				case msg, ok := <-msgs1:
					if !ok {
						t.Logf("Ending nack loop since messages channel is closed")
						return
					}
					t.Logf("Nacking message #%s", msg.UUID)
					msg.Nack()
					nacked <- true
				case <-ctx.Done():
					t.Logf("Context deadline exceeded")
					return
				}
			}
		}(deadlineCtx)

		// Wait for at least 2 nacks to be ongoing
		for range 2 {
			<-nacked
		}

		t.Log("Waiting for sub1 to close")
		err = sub1.Close()
		require.NoError(t, err)
		t.Log("Sub1 closed")
		cancel()

		// following messages [4,10] should be consumed by sub2
		deadlineCtx, cancel = context.WithDeadline(ctx, time.Now().Add(2*time.Second))
		go func(ctx context.Context) {
			for i := range 6 {
				select {
				case msg := <-msgs2:
					t.Logf("Received message #%d from sub2 (i = %d)", i+4, i)
					if fmt.Sprintf("%d", i+4) != msg.UUID {
						t.Logf("Expected message #%d, got %s", i+4, msg.UUID)
						errCh <- fmt.Errorf("expected message #%d, got %s", i+4, msg.UUID)
						return
					}
					msg.Ack()
					t.Logf("Ack from sub2")
					continue
				case <-ctx.Done():
					errCh <- fmt.Errorf("timeout")
					return
				case err := <-errCh:
					errCh <- err
					return
				}
			}

			close(errCh)
		}(deadlineCtx)

		t.Cleanup(cancel)

		// We asynchronously send messages [8,10] to ensure that sub2 receives them
		err = ps.publisher.Publish(ctx, topic, msgs[8:]...)
		require.NoError(t, err)

		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	})
}

func getKafkaUrl(t *testing.T) string {
	t.Helper()
	kafkaUrl := "localhost:19092"
	kafkaUrlFromEnv := os.Getenv("KAFKA")
	if len(kafkaUrlFromEnv) > 0 {
		kafkaUrl = kafkaUrlFromEnv
	}

	return kafkaUrl
}

func setupTestSubscriber(t *testing.T, scope string) (*pubSub, error) {
	t.Helper()

	kconf := &Config{
		Scope:    scope,
		Provider: "kafka",
		Providers: ProvidersConfig{
			Kafka: KafkaConfig{
				Brokers: []string{getKafkaUrl(t)},
			},
		},
	}
	ps, err := New(logrusx.New("test", "test"), kconf)
	if err != nil {
		return nil, err
	}

	cps, ok := ps.(*pubSub)
	if !ok {
		t.Fatal("failed to cast pubSub")
	}

	return cps, nil
}

func registerConsumerGroupCleanup(t *testing.T, group string) error {
	t.Helper()
	cadm, err := getClusterAdmin(t)
	if err != nil {
		return err
	}

	t.Cleanup(func() {
		err := cadm.DeleteConsumerGroup(group)
		if err != nil {
			t.Logf("Failed to delete consumer group %s: %v", group, err)
		}
	})

	return nil
}

func getClusterAdmin(t *testing.T) (sarama.ClusterAdmin, error) {
	t.Helper()
	conf := sarama.NewConfig()
	conf.Version = sarama.V2_8_2_0
	return sarama.NewClusterAdmin([]string{getKafkaUrl(t)}, conf)
}

func setupTestTopic(t *testing.T, topic string) error {
	t.Helper()
	cadm, err := getClusterAdmin(t)
	if err != nil {
		return err
	}

	t.Cleanup(func() {
		err := cadm.DeleteTopic(topic)
		if err != nil {
			t.Logf("Failed to delete topic %s: %v", topic, err)
		}
	})

	return cadm.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
}
