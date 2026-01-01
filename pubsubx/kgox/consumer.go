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
	"github.com/clinia/x/tracex"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"golang.org/x/sync/errgroup"
)

type (
	consumer struct {
		l            *logrusx.Logger
		cl           *kgo.Client
		conf         *pubsubx.Config
		group        messagex.ConsumerGroup
		topics       []messagex.Topic
		opts         *pubsubx.SubscriberOptions
		handlers     map[messagex.Topic]handlerExecutor
		kotelService *kotel.Kotel
		meter        metric.Meter
		erh          *eventRetryHandler
		pqh          PoisonQueueHandler
		tracer       trace.Tracer

		loopProcMu       sync.Mutex
		loopProcessingWg sync.WaitGroup

		mu     sync.RWMutex
		cancel context.CancelFunc
		wg     sync.WaitGroup
		*consumerMetric
	}
	consumerMetric struct {
		recordSize            metric.Int64Histogram
		recordProcessingCount metric.Int64Gauge
	}
)

// TODO: Add the ability to handle repartition and cancel specific topic handling if the parition is revoked
// TODO: Add gauge metric to record the current count of records that are being processed
// TODO: Add a histogram metric to record the latency of the handler
type handlerExecutor struct {
	l *logrusx.Logger
	h pubsubx.Handler
	t messagex.Topic
}

func (he *handlerExecutor) handle(ctx context.Context, msgs []*messagex.Message) (outErrs []error, outErr error) {
	defer func() {
		if r := recover(); r != nil {
			outErr = errorx.InternalErrorf("panic while handling messages: %v", r)
			stackTrace := tracex.GetStackTrace()
			he.l.WithContext(ctx).WithFields(logrusx.NewLogFields(semconv.ExceptionStacktrace(stackTrace))).Errorf("panic while handling messages: %v", r)
		}
	}()
	return he.h(ctx, msgs)
}

var _ pubsubx.Subscriber = (*consumer)(nil)

const (
	// We do not want to have a max elapsed time as we are counting on `maxRetryCount` to stop retrying
	maxElapsedTime = 0
	// We want to wait a max of 3 seconds between retries
	maxRetryInterval = 3 * time.Second
	maxRetryCount    = 3

	componentName = "kgox.consumer"
)

func newConsumerMetric(m metric.Meter) (*consumerMetric, error) {
	rs, err := m.Int64Histogram("recordSize", metric.WithUnit("byte"))
	if err != nil {
		return nil, err
	}
	rpc, err := m.Int64Gauge("recordProcessingCount")
	if err != nil {
		return nil, err
	}
	return &consumerMetric{
		recordSize:            rs,
		recordProcessingCount: rpc,
	}, nil
}

func newConsumer(ctx context.Context, l *logrusx.Logger, kotelService *kotel.Kotel, config *pubsubx.Config, group messagex.ConsumerGroup, topics []messagex.Topic, opts *pubsubx.SubscriberOptions, erh *eventRetryHandler, pqh PoisonQueueHandler, m metric.Meter, t trace.Tracer) (*consumer, error) {
	if l == nil {
		return nil, errorx.FailedPreconditionErrorf("logger is required")
	}

	if opts == nil {
		opts = pubsubx.NewDefaultSubscriberOptions()
	}

	cm, err := newConsumerMetric(m)
	if err != nil {
		l.WithContext(ctx).WithError(err).Errorf("failed to create consumer metrics, not recording any library metrics")
	}

	if t == nil {
		t = noop.NewTracerProvider().Tracer("kgox_consumer")
	}

	cons := &consumer{l: l, kotelService: kotelService, group: group, conf: config, topics: topics, opts: opts, erh: erh, pqh: pqh, consumerMetric: cm, tracer: t}

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
		kgo.WithLogger(&pubsubLogger{l: c.l}),
	}

	if c.opts.DialTimeout > 0 {
		kopts = append(kopts, kgo.DialTimeout(c.opts.DialTimeout))
	}

	if c.opts.RebalanceTimeout > 0 {
		kopts = append(kopts, kgo.RebalanceTimeout(c.opts.RebalanceTimeout))
	}

	if c.kotelService != nil {
		kopts = append(kopts, kgo.WithHooks(c.kotelService.Hooks()...))
	}

	if !c.conf.EnableAutoCommit {
		kopts = append(kopts,
			kgo.DisableAutoCommit(),
		)
		// TODO: Add manual handling of the rebalancing while poling, currently we just assume that we might
		// double process some records if it happens
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
	defer c.mu.Unlock()
	cancel := c.cancel
	if c.cl == nil && cancel == nil {
		// Already closed
		return nil
	}

	// We wait for the current loop to finish if there is ongoing work
	c.loopProcMu.Lock()
	c.loopProcessingWg.Wait()
	c.loopProcMu.Unlock()

	if cancel != nil {
		cancel()
	}

	// Wait for the consumer to be exited
	c.wg.Wait()

	c.cl = nil
	c.cancel = nil

	return nil
}

func (c *consumer) convertRecordsToMessages(rs []*kgo.Record) []*messagex.Message {
	// TODO: Handle the records that can't be Unmarhsal, now they are dropped and lost
	msgs := make([]*messagex.Message, 0, len(rs))
	for _, r := range rs {
		m, err := defaultMarshaler.Unmarshal(r)
		if err != nil {
			c.l.WithError(err).Errorf("failed to unmarshal message: %v", err)
			continue
		}

		msgs = append(msgs, m)
	}
	return msgs
}

func (c *consumer) handleTopic(ctx context.Context, tp kgo.FetchTopic) error {
	// Use base name to handle the retry topics under the same topic handler
	topic := messagex.BaseTopicFromName(tp.Topic)
	l := c.l.WithContext(ctx).WithFields(
		logrusx.NewLogFields(c.attributes(&topic)...),
	)
	ctx = context.WithValue(ctx, ctxLoggerKey, l)
	span := trace.SpanFromContext(ctx)

	// We do not protect the read to handlers here since we cannot get to a point where the handlers are reset and we are still in this consuming loop
	topicHandler, ok := c.handlers[topic]
	if !ok {
		l.Errorf("no handler for topic")
		return errorx.InternalErrorf("no handler for topic")
	}
	records := tp.Records()
	span.SetAttributes(
		semconv.MessagingBatchMessageCount(len(records)),
	)
	if len(records) == 0 {
		l.Debugf("no records to process")
		span.AddEvent("no records to process")
		return nil
	}
	go func() {
		if c.consumerMetric != nil {
			mtc := metric.WithAttributes(attribute.String("topic", tp.Topic))
			c.recordProcessingCount.Record(ctx, int64(len(records)), mtc)
			for _, r := range records {
				if r != nil {
					c.recordSize.Record(ctx, int64(len(r.Value)), mtc)
				}
			}
		}
	}()
	msgs := c.convertRecordsToMessages(records)

	bc := backoff.NewExponentialBackOff()
	bc.MaxElapsedTime = maxElapsedTime
	bc.MaxInterval = maxRetryInterval
	retries := 0
	err := backoff.Retry(func() (outErr error) {
		ctx, span := c.tracer.Start(ctx, componentName+".handleTopic.handle",
			trace.WithAttributes(
				attribute.Int("messaging.batch.handle.retry_count", retries),
			),
		)
		defer func() {
			span.RecordError(outErr)
			span.End()
		}()
		errs, err := topicHandler.handle(ctx, msgs)
		if err != nil {
			l.WithError(err).Errorf("error while handling messages")

			if err == pubsubx.AbortSubscribeError() {
				return backoff.Permanent(err)
			}

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
			c.handleRemoteRetryLogic(ctx, topic, errs, msgs)
		}

		return nil
	}, bc)

	if err == pubsubx.AbortSubscribeError() {
		l.WithError(err).Warnf("subscriber abort error returned from topic : %s", tp.Topic)
		c.handleRemoteRetryLogic(ctx, topic, []error{errorx.NewRetryableError(err)}, msgs)
	} else if err != nil {
		l.WithError(err).Warnf("unhandled error happened while handling topic : %s", tp.Topic)
	}
	return err
}

func (c *consumer) start(ctx context.Context) {
	l := c.l.WithContext(ctx).WithFields(
		logrusx.NewLogFields(c.attributes(nil)...),
	)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		workLoop := func() (shouldBreak bool) {
			// Sync task that hang until it receive at least one record
			// When records are pulled, metadata is also updated within the client
			fetches := c.cl.PollRecords(ctx, int(c.opts.MaxBatchSize))
			ctx, span := c.tracer.Start(ctx, componentName+".workLoop",
				trace.WithSpanKind(trace.SpanKindConsumer),
			)
			defer func() {
				if shouldBreak {
					span.SetStatus(codes.Error, "stopping consumer loop")
				}
				span.End()
			}()
			// This signifies we are currently processing records
			// If the context was cancelled here, this is a noop
			c.loopProcMu.Lock()
			select {
			case <-ctx.Done():
				c.loopProcMu.Unlock()
				l.Infof("context canceled in between loops, stopping consumer")
				return true
			default:
				c.loopProcessingWg.Add(1)
			}
			c.loopProcMu.Unlock()

			defer c.loopProcessingWg.Done()
			defer c.cl.AllowRebalance()

			if errs := fetches.Errors(); len(errs) > 0 {
				// All errors are retried internally when fetching, but non-retriable errors are
				// returned from polls so that users can notice and take action.
				// TODO: Handle errors
				// If its a context canceled error, we should return
				shouldBreak = true
				if lo.SomeBy(errs, func(err kgo.FetchError) bool { return errs[0].Err == context.Canceled }) {
					l.Warnf("context canceled, stopping consumer")
					return shouldBreak
				}
				err := errors.Join(lo.Map(errs, func(err kgo.FetchError, i int) error { return err.Err })...)
				span.RecordError(err)
				l.WithError(err).Errorf("error while polling records, Stopping consumer")
				return shouldBreak
			}

			// We start a goroutine that will simply log if the context is cancelled while we are processing
			// This is purely for logging purpose to know why closing the consumer might take a bit of time
			logCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				select {
				case <-ctx.Done():
					l.Infof("context canceled while processing, will finish processing before stopping consumer")
					return
				case <-logCtx.Done():
					return
				}
			}()
			var wg errgroup.Group
			if c.opts.MaxParallelAsyncExecution <= 0 {
				wg.SetLimit(-1)
			} else {
				wg.SetLimit(int(c.opts.MaxParallelAsyncExecution))
			}

			// From here on, we use a background context as the main context might be cancelled in between this processing occurs and our check for ctx.Done().
			// We want to ensure that the in-progress processing is done even if the main context is cancelled
			backgroundCtx := context.Background()
			var abortErr error
			fetches.EachTopic(func(tp kgo.FetchTopic) {
				backgroundCtx, span := c.tracer.Start(backgroundCtx, componentName+".handleTopic",
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						semconv.MessagingOperationProcess,
						semconv.MessagingSourceKindTopic,
						semconv.MessagingSourceName(tp.Topic),
					),
				)
				defer span.End()
				processTopic := func() error {
					if err := c.handleTopic(backgroundCtx, tp); err != nil {
						switch err {
						case pubsubx.AbortSubscribeError():
							l.WithFields(logrusx.NewLogFields(semconv.MessagingSourceName(tp.Topic))).WithError(err).Errorf("critical error received, stopping consumer")
							span.RecordError(err)
							span.SetStatus(codes.Error, "critical error received, stopping consumer")
							return err
						default:
							return nil
						}
					}
					return nil
				}

				switch {
				case c.opts.EnableAsyncExecution:
					wg.Go(processTopic)
				default:
					abortErr = processTopic()
				}
			})

			if c.opts.EnableAsyncExecution {
				abortErr = wg.Wait()
			}

			if abortErr != nil {
				shouldBreak = true
				return shouldBreak
			}

			if !c.conf.EnableAutoCommit {
				// If commiting the offsets fails, kill the loop by returning
				if err := c.cl.CommitUncommittedOffsets(backgroundCtx); err != nil {
					l.WithError(err).Errorf("failed to commit records")
					shouldBreak = true
					return shouldBreak
				}
			}

			return false
		}

		if shouldBreak := workLoop(); shouldBreak {
			// We stop the consumer if we the workLoop returns true
			return
		}
	}
}

func (c *consumer) handleRemoteRetryLogic(ctx context.Context, topic messagex.Topic, errs []error, msgs []*messagex.Message) {
	l := getContextLogger(ctx, c.l)
	if c.erh == nil {
		l.Errorf("EventRetryHandler should not be nil in the consumer")
		return
	}
	if !c.erh.canTopicRetry() && !c.pqh.CanUsePoisonQueue() {
		l.Debugf("topic retry and poison queue are disable, not executing retry logic")
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
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cl != nil || c.cancel != nil {
		return errorx.InternalErrorf("already subscribed, close the subscriber first")
	}

	handlers := make(map[messagex.Topic]handlerExecutor)
	for _, topic := range c.topics {
		handler, ok := topicHandlers[topic]
		if !ok {
			return errorx.FailedPreconditionErrorf("missing handler for topic %s", topic)
		} else if handler == nil {
			return errorx.FailedPreconditionErrorf("nil handler for topic %s", topic)
		}
		handlers[topic] = handlerExecutor{
			t: topic,
			h: handler,
			l: c.l,
		}
	}

	if err := c.bootstrapClient(ctx); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	c.handlers = handlers
	c.cancel = cancel

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				c.l.WithContext(ctx).Errorf("panic while consuming messages: %v", r)
			}

			// Teardown
			if c.cl != nil {
				c.cl.Close()
			}

			if c.mu.TryLock() {
				// This allows the "spontaneous" failures to set the clients/cancel to nil without
				// impacting a Close() call that already locks this mutex
				c.cl = nil
				c.cancel = nil
				c.mu.Unlock()
			}
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
