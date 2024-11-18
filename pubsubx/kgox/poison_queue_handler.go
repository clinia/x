package kgox

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/twmb/franz-go/pkg/kgo"
)

type poisonQueueHandler PubSub

type PoisonQueueHandler interface {
	PublishMessagesToPoisonQueue(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msgErrs []error, msgs []*messagex.Message) error
	PublishMessagesToPoisonQueueWithGenericError(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msgErr error, msgs ...*messagex.Message) error
	CanUsePoisonQueue() bool
}

var _ PoisonQueueHandler = (*poisonQueueHandler)(nil)

const (
	defaultMissingErrorString    = "important - error is missing"
	originConsumerGroupHeaderKey = "_clinia_origin_consumer_group"
	originErrorHeaderKey         = "_clinia_origin_error"
	originTopicHeaderKey         = "_clinia_origin_topic"
)

func (pqh *poisonQueueHandler) PublishMessagesToPoisonQueue(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msgErrs []error, msgs []*messagex.Message) error {
	if !pqh.conf.PoisonQueue.Enabled && len(msgs) == 0 {
		return nil
	}
	poisonQueueRecords := make([]*kgo.Record, len(msgs))
	checkErrs := len(msgs) == len(msgErrs)
	errs := make([]error, 0, len(msgs))
	for i, msg := range msgs {
		var msgErr error
		if checkErrs && msgErrs[i] != nil {
			msgErr = msgErrs[i]
		}
		pqr, err := pqh.generatePoisonQueueRecord(ctx, topic, consumerGroup, msg, msgErr)
		if err != nil {
			errs = append(errs, err)
			pqh.l.WithError(err).Errorf("failed to generate poison queue record for message id '%s'", msg.ID)
			continue
		}
		poisonQueueRecords[i] = pqr
	}
	err := errors.Join(errs...)
	if err != nil {
		pqh.l.Warnf("failed to generate some poison queue records")
	}
	prs := pqh.writeClient.ProduceSync(ctx, poisonQueueRecords...)
	prErrs := make([]error, 0, len(prs))
	for _, pr := range prs {
		if pr.Err != nil {
			pqh.l.WithError(err).Errorf("failed to publish poison queue record")
			prErrs = append(prErrs, pr.Err)
		}
	}
	return errors.Join(err, errors.Join(prErrs...))
}

func (pqh *poisonQueueHandler) PublishMessagesToPoisonQueueWithGenericError(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msgErr error, msgs ...*messagex.Message) error {
	if !pqh.conf.PoisonQueue.Enabled && len(msgs) == 0 {
		return nil
	}
	poisonQueueRecords := make([]*kgo.Record, len(msgs))
	errs := make([]error, 0, len(msgs))
	for i, msg := range msgs {
		pqr, err := pqh.generatePoisonQueueRecord(ctx, topic, consumerGroup, msg, msgErr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		poisonQueueRecords[i] = pqr
	}
	err := errors.Join(errs...)
	if err != nil {
		pqh.l.WithError(err).Errorf("failed to generate some poison queue records")
	}
	prs := pqh.writeClient.ProduceSync(ctx, poisonQueueRecords...)
	prErrs := make([]error, 0, len(prs))
	for _, pr := range prs {
		if pr.Err != nil {
			prErrs = append(prErrs, pr.Err)
		}
	}
	return errors.Join(err, errors.Join(prErrs...))
}

func (pqh *poisonQueueHandler) generatePoisonQueueRecord(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msg *messagex.Message, msgErr error) (*kgo.Record, error) {
	if msg == nil {
		return nil, errorx.InvalidArgumentErrorf("message can't be nil")
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	errStr := defaultMissingErrorString
	if msgErr != nil {
		errStr = msgErr.Error()
	}
	pqmsg := messagex.NewMessage(payload,
		messagex.WithMetadata(messagex.MessageMetadata{
			originConsumerGroupHeaderKey: string(consumerGroup),
			originTopicHeaderKey:         topic,
			originErrorHeaderKey:         errStr,
		}))
	pqr, err := defaultMarshaler.Marshal(ctx, pqmsg, messagex.TopicFromName(pqh.conf.PoisonQueue.TopicName).TopicName(pqh.conf.Scope))
	if err != nil {
		return nil, err
	}
	return pqr, nil
}

func (pqh *poisonQueueHandler) CanUsePoisonQueue() bool {
	return pqh.conf != nil && pqh.conf.PoisonQueue.IsEnabled()
}

func (pqh *poisonQueueHandler) Consume(ctx context.Context, handlers pubsubx.Handlers, topicFilters ...string) ([]error, error) {
	kopts := []kgo.Opt{
		kgo.ConsumerGroup(messagex.ConsumerGroup(pqh.conf.PoisonQueue.TopicName + ".cg").ConsumerGroup(pqh.conf.Scope)),
		kgo.SeedBrokers(pqh.conf.Providers.Kafka.Brokers...),
		kgo.ConsumeTopics(messagex.Topic(pqh.conf.PoisonQueue.TopicName).TopicName(pqh.conf.Scope)),
	}

	if pqh.kotelService != nil {
		kopts = append(kopts, kgo.WithHooks(pqh.kotelService.Hooks()...))
	}

	cl, err := kgo.NewClient(kopts...)

	fetches := cl.PollFetches(ctx)

	fetches.EachTopic(func(ft kgo.FetchTopic) {
		ft.EachRecord(func(r *kgo.Record) {
			msg, err := defaultMarshaler.Unmarshal(r)
			if err != nil {
				return
			}
			originTopic, ok := msg.Metadata[originTopicHeaderKey]
			if !ok {
				return
			}
			allowed := false
			for _, f := range topicFilters {
				if originTopic == f {
					allowed = true
					break
				}
			}
			if !allowed {
				return
			}
			var eventMessage messagex.Message
			err = json.Unmarshal(msg.Payload, &eventMessage)
			_, err = handlers[messagex.BaseTopicFromName(originTopic)](ctx, []*messagex.Message{&eventMessage})
			if err != nil {
				return
			}
		})
	})

	return []error{}, err
}
