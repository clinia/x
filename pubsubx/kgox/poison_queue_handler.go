package kgox

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/twmb/franz-go/pkg/kgo"
)

type poisonQueueHandler PubSub

type PoisonQueueHandler interface {
	PublishMessagesToPoisonQueue(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msgErrs []error, msgs []*messagex.Message)
	PublishMessagesToPoisonQueueWithGenericError(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msgErr error, msgs ...*messagex.Message)
}

var _ PoisonQueueHandler = (*poisonQueueHandler)(nil)

func (pqh *poisonQueueHandler) PublishMessagesToPoisonQueue(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msgErrs []error, msgs []*messagex.Message) {
	poisonQueueRecords := make([]*kgo.Record, len(msgs))
	checkErrs := len(msgs) == len(msgErrs)
	errs := make([]error, 0, len(msgs))
	for i, msg := range msgs {
		var msgErr error
		if checkErrs && msgErrs[i] != nil {
			msgErr = msgErrs[i]
		}
		pqr, err := pqh.generatePoisonQueueRecord(topic, consumerGroup, msg, msgErr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		poisonQueueRecords[i] = pqr
	}
	if len(errs) > 0 {
		pqh.l.WithError(errors.Join(errs...)).Errorf("failed to generate some poison queue records")
	}
	_ = pqh.writeClient.ProduceSync(ctx, poisonQueueRecords...)
}

func (pqh *poisonQueueHandler) PublishMessagesToPoisonQueueWithGenericError(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msgErr error, msgs ...*messagex.Message) {
	poisonQueueRecords := make([]*kgo.Record, len(msgs))
	errs := make([]error, 0, len(msgs))
	for i, msg := range msgs {
		pqr, err := pqh.generatePoisonQueueRecord(topic, consumerGroup, msg, msgErr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		poisonQueueRecords[i] = pqr
	}
	_ = pqh.writeClient.ProduceSync(ctx, poisonQueueRecords...)
}

func (pqh *poisonQueueHandler) generatePoisonQueueRecord(topic string, consumerGroup messagex.ConsumerGroup, msg *messagex.Message, msgErr error) (*kgo.Record, error) {
	if msg == nil {
		return nil, errorx.InvalidArgumentErrorf("message can't be nil")
	}
	eventMsg, err := defaultMarshaler.Marshal(msg, topic)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(eventMsg)
	if err != nil {
		return nil, err
	}
	errStr := "important - error is missing"
	if msgErr != nil {
		errStr = msgErr.Error()
	}
	pqmsg := messagex.NewMessage(payload,
		messagex.WithMetadata(messagex.MessageMetadata{
			"_clinia_origin_consumer_group": string(consumerGroup),
			"_clinia_origin_topic":          topic,
			"_clinia_origin_err":            errStr,
		}))
	pqr, err := defaultMarshaler.Marshal(pqmsg, messagex.TopicFromName(pqh.conf.PoisonQueueTopic).TopicName(pqh.conf.Scope))
	if err != nil {
		return nil, err
	}
	return pqr, nil
}
