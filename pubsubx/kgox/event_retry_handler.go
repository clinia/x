package kgox

import (
	"context"
	"errors"
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

func (erh *eventRetryHandler) generateRetryTopics(ctx context.Context, topics ...messagex.Topic) []messagex.Topic {
	if !erh.conf.TopicRetry || erh.opts.MaxTopicRetryCount <= 0 || len(topics) == 0 {
		erh.l.Debugf("not generating any retry topics, configuration is either disable, max topic count is <= 0 or no topic are passed in")
		return []messagex.Topic{}
	}
	retryTopics := lo.Map(topics, func(topic messagex.Topic, _ int) messagex.Topic {
		return topic.GenerateRetryTopic(erh.consumerGroup)
	})

	pbac, err := erh.AdminClient()
	if err != nil {
		erh.l.WithError(err).Errorf("failed to generate admin client on retry topic creation")
		return []messagex.Topic{}
	}
	var replicationFactor int16 = math.MaxInt16
	if len(erh.conf.Providers.Kafka.Brokers) <= math.MaxInt16 {
		//nolint:all
		replicationFactor = int16(len(erh.conf.Providers.Kafka.Brokers))
	}

	scopedRetryTopics := lo.Map(retryTopics, func(retryTopic messagex.Topic, _ int) string {
		return retryTopic.TopicName(erh.conf.Scope)
	})
	res, err := pbac.CreateTopics(ctx, 1, replicationFactor, scopedRetryTopics, erh.defaultCreateTopicConfigEntries)
	if err != nil {
		erh.l.WithError(err).Errorf("failed to execute retry topic creation")
		return []messagex.Topic{}
	}
	return lo.Filter(retryTopics, func(retryTopic messagex.Topic, _ int) bool {
		scopeTopic := retryTopic.TopicName(erh.conf.Scope)
		tr, ok := res[scopeTopic]
		if !ok {
			erh.l.Errorf("retry topic [%s] not included in topic creation response", scopeTopic)
			return false
		}
		if tr.Err != nil && tr.Err.Error() != kerr.TopicAlreadyExists.Error() {
			erh.l.WithError(tr.Err).Errorf("failed to create retry topic [%s]", scopeTopic)
			return false
		}
		return true
	})
}

func (c *eventRetryHandler) canTopicRetry() bool {
	return c.conf != nil && c.conf.TopicRetry && c.opts != nil && c.opts.MaxTopicRetryCount > 0
}

func (c *eventRetryHandler) parseRetryMessages(ctx context.Context, errs []error, allMsgs []*messagex.Message) ([]*messagex.Message, []*messagex.Message, []error) {
	l := getContexLogger(ctx, c.l)
	retryableMessages := make([]*messagex.Message, 0)
	poisonQueueMessages := make([]*messagex.Message, 0)
	poisonQueueErrs := make([]error, 0)

	checkErrs := len(errs) == len(allMsgs)
	retryable := false
	var referErr error
	if !checkErrs {
		if len(errs) == 1 {
			l.Debugf("using first error as reference to if we should retry the batch")
			_, retryable = errorx.IsRetryableError(errs[0])
			referErr = errs[0]
		} else {
			l.Warnf("errors handler result mismatch messages length, can't identify which message failed, sending them all back")
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
			l.Warnf("message is missing %s header, setting it to '1'", messagex.RetryCountHeaderKey)
			copiedMsg.Metadata[messagex.RetryCountHeaderKey] = "1"
		} else {
			numericRetryCount, err := strconv.Atoi(retryCount)
			if !c.canTopicRetry() || err != nil || numericRetryCount >= int(c.opts.MaxTopicRetryCount)-1 {
				l.Errorf("not retrying, adding message to poison queue messages")
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
