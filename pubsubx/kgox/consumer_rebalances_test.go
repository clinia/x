package kgox

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/stretchr/testify/require"
)

func TestConsumerRebalancesWithAutoCommit(t *testing.T) {
	l := getLogger()
	config := getPubsubConfig(t, true)
	config.EnableAutoCommit = true
	config.TopicRetry = false

	t.Run("should not commit offsets when handler has a critical error and restarts", func(t *testing.T) {
		ctx := context.Background()
		group, topics := getRandomGroupTopics(t, 1)
		createTopic(t, config, topics[0])

		cg := messagex.ConsumerGroup(group)
		require.NotEmpty(t, cg)

		pubSub, err := NewPubSub(ctx, l, config, nil)
		require.NoError(t, err)

		admCl, err := pubSub.AdminClient()
		require.NoError(t, err)
		kgoxAdmCl, ok := admCl.(*KgoxAdminClient)
		require.True(t, ok)
		require.NotNil(t, kgoxAdmCl)

		msgs := make([]*messagex.Message, 0, 100)
		for i := range cap(msgs) {
			msgs = append(msgs, &messagex.Message{
				ID:      fmt.Sprintf("id-%d", i),
				Payload: []byte("value"),
			})
		}
		errs, err := pubSub.Publisher().PublishSync(ctx, topics[0], msgs...)
		require.NoError(t, err)
		require.NoError(t, errs.FirstNonNil())

		topicRetryCount := 2
		sub, err := pubSub.Subscriber(group, topics, pubsubx.WithMaxTopicRetryCount(topicRetryCount), pubsubx.WithMaxBatchSize(100))
		require.NoError(t, err)

		handlerCh := make(chan struct{})
		t.Cleanup(func() {
			close(handlerCh)
		})

		handlers := pubsubx.Handlers{
			topics[0]: func(ctx context.Context, m []*messagex.Message) ([]error, error) {
				handlerCh <- struct{}{}
				return []error{}, errorx.InternalErrorf("fatal handler error")
			},
		}
		err = sub.Subscribe(ctx, handlers)
		require.NoError(t, err)

		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			go func() {
				for {
					select {
					case <-handlerCh:
					case <-ctx.Done():
						return
					}
				}
			}()
			sub.Close()
		})
		for i := range topicRetryCount + 1 {
			select {
			case <-handlerCh:
				t.Logf("handler called %d times", i+1)
			case <-time.After(10 * time.Second):
				t.Fatalf("handler did not get called")
			}
		}

		go func() {
			<-handlerCh
		}()
		err = sub.Close()
		require.NoError(t, err)

	drain:
		for {
			select {
			case <-handlerCh:
			case <-time.After(2 * time.Second):
				break drain
			}
		}

		// If we restart the consumer, it should start from the last committed offset and not trigger the handler
		err = sub.Subscribe(ctx, handlers)
		require.NoError(t, err)

		select {
		case <-handlerCh:
			t.Logf("handler was called")
		case <-time.After(10 * time.Second):
			t.Fatalf("handler should have been called after re-subscribing")
		}
	})

	t.Run("should commit offsets when handler does not error but close rapidly", func(t *testing.T) {
		ctx := context.Background()
		group, topics := getRandomGroupTopics(t, 1)
		createTopic(t, config, topics[0])

		cg := messagex.ConsumerGroup(group)
		require.NotEmpty(t, cg)

		pubSub, err := NewPubSub(ctx, l, config, nil)
		require.NoError(t, err)

		admCl, err := pubSub.AdminClient()
		require.NoError(t, err)
		kgoxAdmCl, ok := admCl.(*KgoxAdminClient)
		require.True(t, ok)
		require.NotNil(t, kgoxAdmCl)

		msgs := make([]*messagex.Message, 0, 100)
		for i := range cap(msgs) {
			msgs = append(msgs, &messagex.Message{
				ID:      fmt.Sprintf("id-%d", i),
				Payload: []byte("value"),
			})
		}
		errs, err := pubSub.Publisher().PublishSync(ctx, topics[0], msgs...)
		require.NoError(t, err)
		require.NoError(t, errs.FirstNonNil())

		topicRetryCount := 2
		sub, err := pubSub.Subscriber(group, topics, pubsubx.WithMaxTopicRetryCount(topicRetryCount), pubsubx.WithMaxBatchSize(100))
		require.NoError(t, err)

		handlerCh := make(chan struct{})
		waiting := make(chan struct{}, 10)
		t.Cleanup(func() {
			close(handlerCh)
		})

		handlers := pubsubx.Handlers{
			topics[0]: func(ctx context.Context, m []*messagex.Message) ([]error, error) {
				waiting <- struct{}{}
				handlerCh <- struct{}{}
				return []error{}, nil
			},
		}
		err = sub.Subscribe(ctx, handlers)
		require.NoError(t, err)

		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			go func() {
				for {
					select {
					case <-handlerCh:
					case <-ctx.Done():
						return
					}
				}
			}()
			sub.Close()
		})

		<-waiting
		go func() {
			// Simulate slow processing
			// <-time.After(1 * time.Second)
			<-handlerCh
		}()
		err = sub.Close()
		require.NoError(t, err)

	drain:
		for {
			select {
			case <-handlerCh:
			case <-time.After(2 * time.Second):
				break drain
			}
		}

		// If we restart the consumer, it should start from the last committed offset and not trigger the handler
		err = sub.Subscribe(ctx, handlers)
		require.NoError(t, err)

		select {
		case <-handlerCh:
			t.Fatalf("handler should not have been called after re-subscribing")
		case <-time.After(10 * time.Second):
			t.Logf("handler did not get called")
		}
	})
}
