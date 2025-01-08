package kgox

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/clinia/x/errorx"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/clinia/x/tracex"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kadm"
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
	erh          *eventRetryHandler
	pqh          PoisonQueueHandler

	mu     sync.RWMutex
	cancel context.CancelFunc
	closed bool
	wg     sync.WaitGroup

	state state

	adminClient *kadm.Client
}

type state struct {
	lastConsumptionTimePerTopic map[string]time.Time
	previousDescribedGroupLag   kadm.DescribedGroupLag
	currentDescribedGroupLag    kadm.DescribedGroupLag
}

var _ pubsubx.Subscriber = (*consumer)(nil)

const (
	// We do not want to have a max elapsed time as we are counting on `maxRetryCount` to stop retrying
	maxElapsedTime = 0
	// We want to wait a max of 3 seconds between retries
	maxRetryInterval = 3 * time.Second
	maxRetryCount    = 3
)

func newConsumer(l *logrusx.Logger, kotelService *kotel.Kotel, config *pubsubx.Config, group messagex.ConsumerGroup, topics []messagex.Topic, opts *pubsubx.SubscriberOptions, erh *eventRetryHandler, pqh PoisonQueueHandler) (*consumer, error) {
	if l == nil {
		return nil, errorx.FailedPreconditionErrorf("logger is required")
	}

	if opts == nil {
		opts = pubsubx.NewDefaultSubscriberOptions()
	}

	cons := &consumer{l: l, kotelService: kotelService, group: group, conf: config, topics: topics, opts: opts, erh: erh, pqh: pqh}

	if err := cons.bootstrapClient(context.Background()); err != nil {
		return nil, err
	}

	if config.ConsumerGroupMonitoring.IsEnabled() {
		cons.state = state{
			lastConsumptionTimePerTopic: make(map[string]time.Time, len(topics)),
		}
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

func (c *consumer) bootstrapClient(ctx context.Context) error {
	if c.erh == nil {
		return errorx.InternalErrorf("EventRetryHandler should not be nil in the consumer")
	}
	retryTopics, _, err := c.erh.generateRetryTopics(context.Background(), c.topics...)
	if err != nil {
		c.l.WithContext(ctx).WithError(err).Errorf("event retry mechanism might not work properly")
	}
	scopedTopics := make([]string, len(c.topics)+len(retryTopics))
	for i, t := range c.topics {
		scopedTopics[i] = t.TopicName(c.conf.Scope)
	}
	for i, t := range retryTopics {
		scopedTopics[len(c.topics)+i] = t.TopicName(c.conf.Scope)
	}
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

	c.adminClient = kadm.NewClient(c.cl)

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

	if c.conf.ConsumerGroupMonitoring.IsEnabled() {
		go c.monitor(ctx)
	}

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
			l := c.l.WithContext(ctx).WithFields(
				logrusx.NewLogFields(c.attributes(nil)...),
			)
			if lo.SomeBy(errs, func(err kgo.FetchError) bool { return errs[0].Err == context.Canceled }) {
				l.Infof("context canceled, stopping consumer")
				return
			}

			l.WithError(errors.Join(lo.Map(errs, func(err kgo.FetchError, i int) error { return err.Err })...)).Errorf("error while polling records, Stopping consumer")
			return
		}

		fetches.EachTopic(func(tp kgo.FetchTopic) {
			if c.conf.ConsumerGroupMonitoring.IsEnabled() {
				defer func() {
					c.mu.Lock()
					c.state.lastConsumptionTimePerTopic[tp.Topic] = time.Now()
					c.mu.Unlock()
				}()
			}
			// Use base name to handle the retry topics under the same topic handler
			topic := messagex.BaseTopicFromName(tp.Topic)
			l := c.l.WithContext(ctx).WithFields(
				logrusx.NewLogFields(c.attributes(&topic)...),
			)
			ctx := context.WithValue(ctx, ctxLoggerKey, l)
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
				l.Errorf("no handler for topic")
				return
			}
			wrappedHandler := func(ctx context.Context, msgs []*messagex.Message) (outErrs []error, outErr error) {
				defer func() {
					if r := recover(); r != nil {
						outErr = errorx.InternalErrorf("panic while handling messages")
						stackTrace := tracex.GetStackTrace()
						l.WithContext(ctx).WithFields(logrusx.NewLogFields(semconv.ExceptionStacktrace(stackTrace))).Errorf("panic while handling messages")
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

					l.WithError(err).Errorf("error while handling messages")

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
						l.WithError(allErrs).Errorf("errors while handling messages")
					}
					c.handleRemoteRetryLogic(ctx, topic, errs, allMsgs)
				}

				return nil
			}, bc)

			if err == pubsubx.AbortSubscribeError() {
				l.Infof("aborting consumer")
				c.mu.RLock()
				defer c.mu.RUnlock()
				c.handleRemoteRetryLogic(ctx, topic, []error{errorx.NewRetryableError(err)}, allMsgs)
				// This is required since the context is cancelled by the backoff
				if c.cancel != nil {
					c.cancel()
					l.Infof("cancelled consumer")
				} else {
					l.Warnf("abort requested but no cancel function found")
				}
				return
			}
		})
		fmt.Println("end of polling records")
	}
}

func (c *consumer) handleRemoteRetryLogic(ctx context.Context, topic messagex.Topic, errs []error, msgs []*messagex.Message) {
	l := getContextLogger(ctx, c.l)
	if c.erh == nil {
		l.Errorf("EventRetryHandler should not be nil in the consumer")
		return
	}
	if !c.erh.canTopicRetry() && !c.pqh.CanUsePoisonQueue() {
		l.Debugf("topic retry and poison queue are disable, not exeucting retry logic")
		return
	}
	retryableMessages, poisonQueueMessages, poisonQueueErrs := c.erh.parseRetryMessages(ctx, errs, msgs)
	if c.erh.canTopicRetry() && len(retryableMessages) > 0 {
		if publishErr := c.erh.publishRetryMessages(ctx, retryableMessages, topic); publishErr != nil {
			l.WithError(publishErr).Errorf("failed to publish as some or all retry messages")
		}
	}
	if c.pqh.CanUsePoisonQueue() && len(poisonQueueMessages) > 0 {
		if publishErr := c.publishPoisonQueueMessages(ctx, topic, poisonQueueMessages, poisonQueueErrs); publishErr != nil {
			l.WithError(publishErr).Errorf("failed to publish some or all retry messages to the poison queue")
		}
	}
}

func (c *consumer) publishPoisonQueueMessages(ctx context.Context, topic messagex.Topic, msgs []*messagex.Message, errs []error) error {
	topicName := topic.TopicName(c.conf.Scope)
	localErrs := errs
	if len(errs) > 1 && len(errs) != len(msgs) {
		c.l.WithContext(ctx).Errorf("tried to publish poison queue messages but error don't match messages number, failing back to empty error")
		localErrs = []error{}
	}
	if len(errs) == 1 {
		return c.pqh.PublishMessagesToPoisonQueueWithGenericError(ctx, topicName, c.group, errs[0], msgs...)
	} else {
		return c.pqh.PublishMessagesToPoisonQueue(ctx, topicName, c.group, localErrs, msgs)
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
		if err := c.bootstrapClient(ctx); err != nil {
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
				c.l.WithContext(ctx).Errorf("panic while consuming messages: %v", r)
			}

			// Teardown
			c.mu.Lock()
			c.cancel()
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

	err := errorx.InternalErrorf("consumer group %s is not healthy", c.group.ConsumerGroup(c.conf.Scope))

	if c.cancel == nil {
		return err
	}

	if !c.conf.ConsumerGroupMonitoring.IsEnabled() {
		return nil
	}

	if c.state.previousDescribedGroupLag.Group == "" || c.state.currentDescribedGroupLag.Group == "" {
		return nil
	}

	for topic, lastConsumptionTime := range c.state.lastConsumptionTimePerTopic {
		timeElapsed := time.Since(lastConsumptionTime)

		// If last consumption time for at least one topic is less than the health timeout, return healthy
		if timeElapsed < c.conf.ConsumerGroupMonitoring.HealthTimeout {
			return nil
		}

		previousTopicLags := c.state.previousDescribedGroupLag.Lag[topic]
		currentTopicLags := c.state.currentDescribedGroupLag.Lag[topic]

		for i, currentPartitionLag := range currentTopicLags {
			currentPartitionOffset := currentPartitionLag.Commit.At
			previousPartitionOffset := previousTopicLags[i].Commit.At

			// If offset is moving for at least one partition, return healthy
			if currentPartitionOffset > previousPartitionOffset {
				return nil
			}

			if currentPartitionLag.Lag > 0 {
				// If we reach here, we have a case for unhealthy for that specific partition
				err.WithDetails(errorx.InternalErrorf("topic '%s' partition %d: no consumption for %s, lag is %d, and offset %d is not moving", topic, currentPartitionLag.Partition, timeElapsed, currentPartitionLag.Lag, currentPartitionOffset))
			}

		}
	}

	// If no early returns, it means that:
	// 1. All topics of the consumer group did not consume any messages for the last `HealthTimeout` duration
	// 2. All the partitions of the topics of the consumer group have offsets that remain constant (not moving)
	// 3. A lag still exists for at least one partition
	// From the above, we can infer that the consumer group is unhealthy ("stuck")
	if len(err.Details) > 0 {
		return err
	}

	return nil
}

// monitor refreshes the lag state of the consumer group at the interval set in the config
func (c *consumer) monitor(ctx context.Context) {
	ticker := time.NewTicker(c.conf.ConsumerGroupMonitoring.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := c.refreshLagState(ctx)
			if err != nil {
				c.l.WithContext(ctx).WithError(errorx.InternalErrorf("failed to fetch lag for consumer group %s: %v", c.group.ConsumerGroup(c.conf.Scope), err))

				// Teardown
				// Note: Next call to Health() will return error and will trigger consumer group restart
				c.mu.Lock()
				c.cancel = nil
				c.mu.Unlock()

				c.wg.Done()
				fmt.Println("Teardown done")

				return
			}
		}
	}
}

// refreshLagState fetches the lag with retry mechanism from the admin client and updates the state
func (c *consumer) refreshLagState(ctx context.Context) error {
	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = 3 * time.Second

	var groupLags map[string]kadm.DescribedGroupLag
	var err error
	retryErr := backoff.Retry(func() error {
		groupLags, err = c.adminClient.Lag(ctx, c.group.ConsumerGroup(c.conf.Scope))
		return err
	}, bc)
	if retryErr != nil {
		return retryErr
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// copy the current state to the previous state
	c.state.previousDescribedGroupLag = c.state.currentDescribedGroupLag
	// update the current state
	c.state.currentDescribedGroupLag = groupLags[c.group.ConsumerGroup(c.conf.Scope)]

	// If all the lags are 0, reset all the last consumption times
	if c.allLagsZero() {
		c.state.lastConsumptionTimePerTopic = make(map[string]time.Time, len(c.topics))
	}

	return nil
}

// allLagsZero checks if all the lags in the current described group lag are 0
func (c *consumer) allLagsZero() bool {
	for _, topicLags := range c.state.currentDescribedGroupLag.Lag {
		for _, partitionLag := range topicLags {
			if partitionLag.Lag > 0 {
				return false
			}
		}
	}

	return true
}
