package kgox

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/clinia/x/errorx"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	tracexx "github.com/clinia/x/tracex"
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

const (
	// We do not want to have a max elapsed time as we are counting on `maxRetryCount` to stop retrying
	maxElapsedTime = 0
	// We want to wait a max of 3 seconds between retries
	maxRetryInterval = 3 * time.Second
	maxRetryCount    = 3
)

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
	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = maxElapsedTime
	bc.MaxInterval = maxRetryInterval

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
			if lo.SomeBy(errs, func(err kgo.FetchError) bool { return errs[0].Err == context.Canceled }) {
				l.Infof("context canceled, stopping consumer")
				return
			}

			l.WithError(errors.Join(lo.Map(errs, func(err kgo.FetchError, i int) error { return err.Err })...)).Error("error while polling records, Stopping consumer")
			return
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
						outErr = errorx.InternalErrorf("panic while handling messages")
						stackTrace := tracexx.GetStackTrace()
						l.WithFields(logrusx.NewLogFields(semconv.ExceptionStacktrace(stackTrace))).Errorln("panic while handling messages")
					}
				}()
				return topicHandler(ctx, msgs)
			}

			bc.Reset()
			retries := 0
			err := backoff.Retry(func() error {
				errs, err := wrappedHandler(ctx, allMsgs)
				if err != nil {
					if err == pubsubx.AbortSubscribeError() {
						return backoff.Permanent(err)
					}

					l.WithError(err).Errorln("error while handling messages")

					retries++
					if retries > maxRetryCount {
						// In this case, we should abort the subscription as this is most likely a critical error
						return backoff.Permanent(pubsubx.AbortSubscribeError())
					}

					return err
				}

				if len(errs) > 0 {
					allErrs := errors.Join(errs...)
					if allErrs != nil {
						l.WithError(allErrs).Errorln("errors while handling messages")
					}

					// TODO: [ENG-1361] retry logic (resend to the same topic with a retry count per-event)
				}

				return nil
			}, bc)

			if err == pubsubx.AbortSubscribeError() {
				l.Infof("aborting consumer")
				c.mu.RLock()
				defer c.mu.RUnlock()
				if c.cancel != nil {
					c.cancel()
					l.Infof("cancelled consumer")
				} else {
					l.Warnln("abort requested but no cancel function found")
				}

				// TODO: [ENG-1361] resend the messages to the same topic with a retry count increased for each event
				return
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
			if r := recover(); r != nil {
				c.l.Errorf("panic while consuming messages: %v", r)
			}

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

// Health implements pubsubx.Subscriber.
func (c *consumer) Health() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cancel == nil {
		return errorx.InternalErrorf("not subscribed to topics")
	}

	return nil
}
