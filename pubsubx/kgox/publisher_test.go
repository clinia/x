package kgox

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kerr"
)

func TestPublisher(t *testing.T) {
	l := getLogger()
	config := getPubsubConfig(t, false)

	receivedMessages := func(t *testing.T, group string, topic messagex.Topic) <-chan *messagex.Message {
		t.Helper()
		msgsCh := make(chan *messagex.Message, 10)
		ctx, cancel := context.WithCancel(context.Background())

		pubSub, err := NewPubSub(l, config, nil)
		require.NoError(t, err)

		sub, err := pubSub.Subscriber(group, []messagex.Topic{topic})
		require.NoError(t, err)

		err = sub.Subscribe(ctx, pubsubx.Handlers{
			topic: func(ctx context.Context, msgs []*messagex.Message) ([]error, error) {
				for _, msg := range msgs {
					msgsCh <- msg
				}
				return nil, nil
			},
		})
		require.NoError(t, err)

		t.Cleanup(func() {
			cancel()
			sub.Close()
			close(msgsCh)
		})

		return msgsCh
	}

	expectReceivedMessages := func(t *testing.T, msgsCh <-chan *messagex.Message, expectedMsgs ...*messagex.Message) {
		t.Helper()

		for _, expectedMsg := range expectedMsgs {
			select {
			case receivedMsg := <-msgsCh:
				require.Equal(t, expectedMsg, receivedMsg)
			case <-time.After(defaultExpectedReceiveTimeout):
				t.Fatal("timed out waiting for message")
			}
		}
	}

	expectNoMessages := func(t *testing.T, msgsCh <-chan *messagex.Message) {
		t.Helper()

		select {
		case <-msgsCh:
			t.Fatal("received unexpected message")
		case <-time.After(defaultExpectedNoReceiveTimeout):
			// Expected
		}
	}

	t.Run("PublishSync", func(t *testing.T) {
		p := getPublisher(t, l, config)
		t.Cleanup(func() { p.Close() })

		group, topics := getRandomGroupTopics(t, 1)
		testTopic := topics[0]
		createTopic(t, config, testTopic)

		receivedMsgsCh := receivedMessages(t, group, testTopic)

		msg := messagex.NewMessage([]byte("test"))
		errs, err := p.PublishSync(context.Background(), testTopic, msg)
		require.NoError(t, err)
		require.NoError(t, errs.FirstNonNil())

		expectReceivedMessages(t, receivedMsgsCh, msg)

		// Should be able to publish multiple messages
		msgs := []*messagex.Message{}
		for i := 0; i < 10; i++ {
			msgs = append(msgs, messagex.NewMessage([]byte(fmt.Sprintf("test-%d", i))))
		}

		errs, err = p.PublishSync(context.Background(), testTopic, msgs...)
		require.NoError(t, err)
		require.NoError(t, errs.FirstNonNil())

		expectReceivedMessages(t, receivedMsgsCh, msgs...)

		tooLargePayload := make([]byte, 1000000)
		for i := range tooLargePayload {
			tooLargePayload[i] = 'a'
		}

		msgTooLarge := messagex.NewMessage(tooLargePayload, messagex.WithMetadata(map[string]string{
			"key1": "value1",
			"key2": "value2",
		}))

		errs, err = p.PublishSync(context.Background(), testTopic, msgTooLarge)
		assert.Contains(t, err.Error(), kerr.MessageTooLarge.Error())
		assert.Contains(t, err.Error(), fmt.Sprintf("{_clinia_message_id %s}", msgTooLarge.ID))
		for _, errL := range errs {
			assert.Contains(t, errL.Error(), kerr.MessageTooLarge.Error())
			assert.Contains(t, errL.Error(), fmt.Sprintf("{_clinia_message_id %s}", msgTooLarge.ID))
		}

		// Should be able to publish a mix of correct and too large messages
		msgs = []*messagex.Message{}
		okMsgs := []*messagex.Message{}
		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				msgs = append(msgs, messagex.NewMessage([]byte(fmt.Sprintf("test-%d", i))))
				okMsgs = append(okMsgs, msgs[len(msgs)-1])
			} else {
				msgs = append(msgs, messagex.NewMessage(tooLargePayload, messagex.WithMetadata(map[string]string{
					"id": fmt.Sprintf("msg-%d", i),
				})))
			}
		}

		errs, err = p.PublishSync(context.Background(), testTopic, msgs...)
		assert.Error(t, err)
		for _, errL := range errs {
			if errL != nil {
				assert.Contains(t, errL.Error(), kerr.MessageTooLarge.Error())
			}
		}

		expectReceivedMessages(t, receivedMsgsCh, okMsgs...)
	})

	t.Run("PublishAsync", func(t *testing.T) {
		p := getPublisher(t, l, config)
		t.Cleanup(func() { p.Close() })

		group, topics := getRandomGroupTopics(t, 1)
		testTopic := topics[0]
		createTopic(t, config, testTopic)

		receivedMsgsCh := receivedMessages(t, group, testTopic)

		msg := messagex.NewMessage([]byte("test"))
		err := p.PublishAsync(context.Background(), testTopic, msg)
		require.NoError(t, err)

		expectReceivedMessages(t, receivedMsgsCh, msg)

		// Should be able to publish multiple messages
		msgs := []*messagex.Message{}
		for i := 0; i < 10; i++ {
			msgs = append(msgs, messagex.NewMessage([]byte(fmt.Sprintf("test-%d", i))))
		}

		err = p.PublishAsync(context.Background(), testTopic, msgs...)
		require.NoError(t, err)

		expectReceivedMessages(t, receivedMsgsCh, msgs...)
	})

	t.Run("should be able to cancel sending messages with PublishAsync", func(t *testing.T) {
		p := getPublisher(t, l, config)
		t.Cleanup(func() { p.Close() })

		group, topics := getRandomGroupTopics(t, 1)
		testTopic := topics[0]
		createTopic(t, config, testTopic)

		receivedMsgsCh := receivedMessages(t, group, testTopic)

		ctx, cancel := context.WithCancel(context.Background())

		msg := messagex.NewMessage([]byte("test"))
		err := p.PublishAsync(ctx, testTopic, msg)
		require.NoError(t, err)

		cancel()
		expectNoMessages(t, receivedMsgsCh)

		// Should be able to publish multiple messages
		msgs := []*messagex.Message{}
		for i := 0; i < 10; i++ {
			msgs = append(msgs, messagex.NewMessage([]byte(fmt.Sprintf("test-%d", i))))
		}

		// Context is now already cancelled, we should not receive any messages
		err = p.PublishAsync(ctx, testTopic, msgs...)
		require.NoError(t, err)

		expectNoMessages(t, receivedMsgsCh)
	})

	t.Run("should not be able to Publish after close", func(t *testing.T) {
		p := getPublisher(t, l, config)

		group, topics := getRandomGroupTopics(t, 1)
		testTopic := topics[0]
		createTopic(t, config, testTopic)

		receivedMsgsCh := receivedMessages(t, group, testTopic)

		msg := messagex.NewMessage([]byte("test"))
		errs, err := p.PublishSync(context.Background(), testTopic, msg)
		require.NoError(t, err)
		require.NoError(t, errs.FirstNonNil())

		expectReceivedMessages(t, receivedMsgsCh, msg)

		err = p.Close()
		require.NoError(t, err)

		msg2 := messagex.NewMessage([]byte("test2"))
		publishSync := func() <-chan error {
			errCh := make(chan error, 1)
			go func() {
				errs, _ := p.PublishSync(context.Background(), testTopic, msg2)
				errCh <- errs.Join()
			}()
			return errCh
		}
		select {
		case <-time.After(1 * time.Second):
			// Expected
		case <-publishSync():
			t.Fatal("publisher should not be useful after close")
		}
	})
}
