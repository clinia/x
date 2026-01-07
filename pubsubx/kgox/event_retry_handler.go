package kgox

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
)

type eventRetryHandler struct {
	*PubSub
	consumerGroup messagex.ConsumerGroup
	opts          *pubsubx.SubscriberOptions
}

func (erh *eventRetryHandler) generateRetryTopics(ctx context.Context, topics ...messagex.Topic) ([]messagex.Topic, []error, error) {
	l := erh.l
	if !erh.canTopicRetry() || len(topics) == 0 {
		l.Debug(ctx, "not generating any retry topics, configuration is either disable, max topic count is <= 0 or no topic are passed in")
		return []messagex.Topic{}, []error{}, nil
	}
	retryTopics := lo.Map(topics, func(topic messagex.Topic, _ int) messagex.Topic {
		return topic.GenerateRetryTopic(erh.consumerGroup)
	})

	pbac, err := erh.AdminClient()
	if err != nil {
		l.WithError(err).Error(ctx, "failed to generate admin client on retry topic creation")
		return []messagex.Topic{}, []error{}, err
	}
	defer pbac.Close()

	scopedRetryTopics := lo.Map(retryTopics, func(retryTopic messagex.Topic, _ int) string {
		return retryTopic.TopicName(erh.conf.Scope)
	})
	var replicationFactor int16 = math.MaxInt16
	if len(erh.conf.Providers.Kafka.Brokers) <= math.MaxInt16 {
		//nolint:all
		replicationFactor = int16(len(erh.conf.Providers.Kafka.Brokers))
	}
	res, err := pbac.CreateTopics(ctx, 1, replicationFactor, scopedRetryTopics, erh.defaultCreateTopicConfigEntries)
	if err != nil {
		l.WithError(err).Error(ctx, "failed to execute retry topic creation")
		return []messagex.Topic{}, []error{}, err
	}
	out := make([]messagex.Topic, 0, len(topics))
	errs := make([]error, len(topics))
	for i, retryTopic := range retryTopics {
		scopeTopic := retryTopic.TopicName(erh.conf.Scope)
		tr, ok := res[scopeTopic]
		if !ok {
			l.Error(ctx, fmt.Sprintf("retry topic [%s] not included in topic creation response", scopeTopic))
			errs[i] = errorx.InvalidArgumentErrorf("retry topic [%s] not included in topic creation response", scopeTopic)
			continue
		}
		if tr.Err != nil && tr.Err.Error() != kerr.TopicAlreadyExists.Error() {
			l.WithError(tr.Err).Error(ctx, fmt.Sprintf("failed to create retry topic [%s]", scopeTopic))
			errs[i] = tr.Err
			continue
		}
		out = append(out, retryTopic)
	}
	return out, errs, errors.Join(errs...)
}

func (c *eventRetryHandler) canTopicRetry() bool {
	return c.conf != nil && c.conf.TopicRetry && c.opts != nil && c.opts.MaxTopicRetryCount > 0
}

func (c *eventRetryHandler) parseRetryMessages(ctx context.Context, errs []error, allMsgs []*messagex.Message) ([]*messagex.Message, []*messagex.Message, []error) {
	l := getContextLogger(ctx, c.l)
	retryableMessages := make([]*messagex.Message, 0)
	poisonQueueMessages := make([]*messagex.Message, 0)
	poisonQueueErrs := make([]error, 0)

	checkErrs := len(errs) == len(allMsgs)
	retryable := false
	var referErr error
	if !checkErrs {
		if len(errs) == 1 {
			l.Debug(ctx, "using first error as reference to if we should retry the batch")
			_, retryable = errorx.IsRetryableError(errs[0])
			referErr = errs[0]
		} else {
			l.Warn(ctx, "errors handler result mismatch messages length, can't identify which message failed, sending them all back")
		}
	}
	for i, msg := range allMsgs {
		if msg == nil {
			continue
		}
		localRetryable := retryable
		if checkErrs {
			referErr = errs[i]
			if referErr == nil {
				continue
			}
			_, localRetryable = errorx.IsRetryableError(referErr)
		}
		copiedMsg := msg.Copy()
		if !localRetryable {
			poisonQueueMessages = append(poisonQueueMessages, copiedMsg)
			poisonQueueErrs = append(poisonQueueErrs, referErr)
			continue
		}
		retryCount, ok := msg.Metadata[messagex.RetryCountHeaderKey]
		if !ok {
			l.Warn(ctx, fmt.Sprintf("message is missing %s header, setting it to '1'", messagex.RetryCountHeaderKey))
			copiedMsg.Metadata[messagex.RetryCountHeaderKey] = "1"
		} else {
			numericRetryCount, err := strconv.Atoi(retryCount)
			if !c.canTopicRetry() || err != nil || numericRetryCount >= int(c.opts.MaxTopicRetryCount)-1 {
				l.WithError(err).Error(ctx, "not retrying, adding message to poison queue messages")
				poisonQueueMessages = append(poisonQueueMessages, copiedMsg)
				poisonQueueErrs = append(poisonQueueErrs, referErr)
				continue
			}
			copiedMsg.Metadata[messagex.RetryCountHeaderKey] = strconv.Itoa(numericRetryCount + 1)
		}
		retryableMessages = append(retryableMessages, copiedMsg)
	}
	return retryableMessages, poisonQueueMessages, poisonQueueErrs
}

func (c *eventRetryHandler) publishRetryMessages(ctx context.Context, retryableMessages []*messagex.Message, topic messagex.Topic) error {
	retryRecords := make([]*kgo.Record, len(retryableMessages))
	marshalErrs := make(pubsubx.Errors, len(retryableMessages))
	scopedTopic := topic.GenerateRetryTopic(c.consumerGroup).TopicName(c.conf.Scope)
	for i, m := range retryableMessages {
		retryRecords[i], marshalErrs[i] = defaultMarshaler.Marshal(ctx, m, scopedTopic)
	}
	marshalErr := errors.Join(marshalErrs...)
	retryRecords = slices.DeleteFunc(retryRecords, func(r *kgo.Record) bool {
		return r == nil
	})
	produceResults := c.writeClient.ProduceSync(ctx, retryRecords...)
	produceErrs := make(pubsubx.Errors, len(produceResults))
	for i, record := range produceResults {
		produceErrs[i] = record.Err
	}
	produceErr := errors.Join(produceErrs...)
	return errors.Join(marshalErr, produceErr)
}
