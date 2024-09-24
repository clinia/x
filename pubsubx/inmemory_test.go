package pubsubx

import (
	"context"
	"testing"

	"github.com/clinia/x/pubsubx/messagex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriber(t *testing.T) {
	t.Run("Multiple Subscribers", func(t *testing.T) {
		pubsub := &memoryPubSub{}
		sub1, err := pubsub.Subscriber("group1", []messagex.Topic{"topic1"})
		require.NoError(t, err)

		sub2, err := pubsub.Subscriber("group2", []messagex.Topic{"topic1", "topic2"})
		require.NoError(t, err)

		seen1 := make(chan []*messagex.Message, 10)
		handlers1 := Handlers{
			"topic1": func(ctx context.Context, messages []*messagex.Message) ([]error, error) {
				seen1 <- messages
				return make([]error, len(messages)), nil
			},
		}
		err = sub1.Subscribe(context.Background(), handlers1)
		require.NoError(t, err)

		seen2Topic1 := make(chan []*messagex.Message, 10)
		seen2Topic2 := make(chan []*messagex.Message, 10)
		handlers2 := Handlers{
			"topic1": func(ctx context.Context, messages []*messagex.Message) ([]error, error) {
				seen2Topic1 <- messages
				return make([]error, len(messages)), nil
			},
			"topic2": func(ctx context.Context, messages []*messagex.Message) ([]error, error) {
				seen2Topic2 <- messages
				return make([]error, len(messages)), nil
			},
		}
		err = sub2.Subscribe(context.Background(), handlers2)
		require.NoError(t, err)

		messages := []*messagex.Message{
			messagex.NewMessage([]byte("message1")),
			messagex.NewMessage([]byte("message2")),
		}
		errs, err := pubsub.PublishSync(context.Background(), "topic1", messages...)
		require.NoError(t, err)
		require.NoError(t, errs.FirstNonNil())

		select {
		case <-seen2Topic2:
			t.Fatal("expected no messages on topic2")
		default:
		}

		for _, ch := range []chan []*messagex.Message{seen1, seen2Topic1} {
			select {
			case m := <-ch:
				assert.Len(t, m, 2)
				assert.ElementsMatch(t, messages, m)
			default:
				t.Fatal("expected messages")
			}
		}

		err = sub1.Close()
		require.NoError(t, err)

		errs, err = pubsub.PublishSync(context.Background(), "topic1", messages...)
		require.NoError(t, err)
		require.NoError(t, errs.FirstNonNil())

		select {
		case m := <-seen2Topic1:
			assert.Len(t, m, 2)
			assert.ElementsMatch(t, messages, m)
		default:
			t.Fatal("expected messages")
		}

		for _, ch := range []chan []*messagex.Message{seen1, seen2Topic2} {
			select {
			case <-ch:
				t.Fatal("expected no messages")
			default:
			}
		}
	})
}
