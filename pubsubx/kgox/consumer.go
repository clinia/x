package kgox

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type consumer struct {
	l            *logrusx.Logger
	cl           *kgo.Client
	conf         *pubsubx.Config
	group        messagex.ConsumerGroup
	topics       []messagex.Topic
	opts         *pubsubx.SubscriberOptions
	handlers     pubsubx.Handlers
	kotelService *kotel.Kotel

	mu     sync.RWMutex
	cancel context.CancelFunc
	closed bool
	wg     sync.WaitGroup
}

var _ pubsubx.Subscriber = (*consumer)(nil)

func newConsumer(l *logrusx.Logger, kotelService *kotel.Kotel, config *pubsubx.Config, group string, topics []messagex.Topic, opts *pubsubx.SubscriberOptions) (*consumer, error) {
	if l == nil {
		return nil, errorx.FailedPreconditionErrorf("logger is required")
	}

	if opts == nil {
		opts = pubsubx.NewDefaultSubscriberOptions()
	}

	cons := &consumer{l: l, kotelService: kotelService, group: messagex.ConsumerGroup(group), conf: config, topics: topics, opts: opts}

	if err := cons.bootstrapClient(); err != nil {
		return nil, err
	}

	return cons, nil
}

func (c *consumer) attributes(topic *messagex.Topic) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.MessagingSystemKey.String("kafka"),
		semconv.MessagingKafkaConsumerGroup(c.group.ConsumerGroup(c.conf.Scope)),
	}

	if topic != nil {
		attrs = append(attrs,
			semconv.MessagingSourceKindTopic,
			semconv.MessagingSourceName(topic.TopicName(c.conf.Scope)),
		)
	}

	return attrs
}

func (c *consumer) bootstrapClient() error {
	scopedTopics := lo.Map(c.topics, func(topic messagex.Topic, _ int) string {
		return topic.TopicName(c.conf.Scope)
	})
	kopts := []kgo.Opt{
		kgo.ConsumerGroup(c.group.ConsumerGroup(c.conf.Scope)),
		kgo.SeedBrokers(c.conf.Providers.Kafka.Brokers...),
		kgo.ConsumeTopics(scopedTopics...),
	}

	if c.kotelService != nil {
		kopts = append(kopts, kgo.WithHooks(c.kotelService.Hooks()...))
	}

	client, err := kgo.NewClient(kopts...)
	if err != nil {
		return err
	}

	c.cl = client
	return nil
}

// Close implements pubsubx.Subscriber.
func (c *consumer) Close() error {
	c.mu.Lock()
	if c.closed && c.cancel == nil {
		c.mu.Unlock()
		// Already closed
		return nil
	}

	c.closed = true

	if cancel := c.cancel; cancel != nil {
		cancel()
	}
	c.mu.Unlock()

	// Wait for the consumer to be done
	c.wg.Wait()

	// Close the client
	c.cl.Close()

	return nil
}

func (c *consumer) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		fetches := c.cl.PollRecords(ctx, int(c.opts.MaxBatchSize))
		if errs := fetches.Errors(); len(errs) > 0 {
			// All errors are retried internally when fetching, but non-retriable errors are
			// returned from polls so that users can notice and take action.
			// TODO: Handle errors
			// If its a context canceled error, we should return
			l := c.l.WithFields(
				logrusx.NewLogFields(c.attributes(nil)...),
			)
			if errs[0].Err == context.Canceled {
				l.Infof("context canceled, stopping consumer")
				return
			}

			l.WithError(errs[0].Err).Error("error while polling records")
			panic(fmt.Sprintf("unexpected error: %v", errs[0].Err))
		}

		fetches.EachTopic(func(tp kgo.FetchTopic) {
			topic := messagex.TopicFromName(tp.Topic)
			l := c.l.WithFields(
				logrusx.NewLogFields(c.attributes(&topic)...),
			)
			records := tp.Records()
			allMsgs := make([]*messagex.Message, 0, len(records))
			for _, record := range records {
				msg, err := defaultMarshaler.Unmarshal(record)
				if err != nil {
					l.Warnf("failed to unmarshal message: %v", err)
					continue
				}

				allMsgs = append(allMsgs, msg)
			}

			// We do not protect the read to handlers here since we cannot get to a point where the handlers are reset and we are still in this consuming loop
			topicHandler, ok := c.handlers[topic]
			if !ok || topicHandler == nil {
				l.Errorln("no handler for topic")
				return
			}
			wrappedHandler := func(ctx context.Context, msgs []*messagex.Message) (outErrs []error, outErr error) {
				defer func() {
					if r := recover(); r != nil {
						l.Errorf("panic while handling messages: %v", r)
						// TODO: this should either be retried, or if it continues failing we should have a way to signal that the consumer is not healthy
						outErr = errorx.InternalErrorf("panic while handling messages")
					}
				}()
				return topicHandler(ctx, msgs)
			}
			errs, err := wrappedHandler(ctx, allMsgs)
			if err != nil {
				l.WithError(err).Errorln("error while handling messages")
				// TODO: if this is a critical error, we should return close the consumer
				// TODO: Check if this is a critical error. If it is not critical, we need to resend the messages right away
				return
			}

			if len(errs) > 0 {
				allErrs := errors.Join(errs...)
				if allErrs != nil {
					l.WithError(allErrs).Errorln("errors while handling messages")
				}

				// TODO: retry logic
			}
		})
	}
}

// Subscribe implements pubsubx.Subscriber.
func (c *consumer) Subscribe(ctx context.Context, topicHandlers pubsubx.Handlers) error {
	for _, topic := range c.topics {
		if handler, ok := topicHandlers[topic]; !ok {
			return errorx.FailedPreconditionErrorf("missing handler for topic %s", topic)
		} else if handler == nil {
			return errorx.FailedPreconditionErrorf("nil handler for topic %s", topic)
		}
	}

	c.mu.RLock()
	if c.closed {
		// This means that we have closed the consumer in between calls
		// Make sure we wait for the consumer to be done (and thus cancel to be reset)
		c.wg.Wait()

		// We can create a new client
		if err := c.bootstrapClient(); err != nil {
			c.mu.RUnlock()
			return err
		}
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		return errorx.InternalErrorf("already subscribed to topics")
	}
	ctx, cancel := context.WithCancel(ctx)

	c.handlers = topicHandlers
	c.cancel = cancel

	c.wg.Add(1)
	go func() {
		defer func() {
			// Teardown
			c.mu.Lock()
			c.cancel = nil
			c.mu.Unlock()

			c.wg.Done()
		}()

		c.start(ctx)
	}()

	return nil
}
