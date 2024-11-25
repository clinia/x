package kgox

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

type poisonQueueHandler PubSub

type PoisonQueueHandler interface {
	PublishMessagesToPoisonQueue(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msgErrs []error, msgs []*messagex.Message) error
	PublishMessagesToPoisonQueueWithGenericError(ctx context.Context, topic string, consumerGroup messagex.ConsumerGroup, msgErr error, msgs ...*messagex.Message) error
	CanUsePoisonQueue() bool
	ConsumeQueue(ctx context.Context, handler pubsubx.Handler) ([]error, error)
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
	return pqh.conf != nil && pqh.conf.PoisonQueue.IsEnabled() && pqh.conf.PoisonQueue.TopicName != ""
}

// ConsumeQueue consumes all the event in the poison queue up to the moment this function is called,
// only one execution of the handler is done to prevent pushing back any event on the poison queue while consuming
func (pqh *poisonQueueHandler) ConsumeQueue(ctx context.Context, handler pubsubx.Handler) ([]error, error) {
	if !pqh.CanUsePoisonQueue() {
		return []error{}, errorx.InternalErrorf("poison queue is disabled")
	}
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	pqTopic := messagex.Topic(pqh.conf.PoisonQueue.TopicName).TopicName(pqh.conf.Scope)
	kopts := []kgo.Opt{
		kgo.SeedBrokers(pqh.conf.Providers.Kafka.Brokers...),
		kgo.ConsumeTopics(pqTopic),
	}
	if pqh.kotelService != nil {
		kopts = append(kopts, kgo.WithHooks(pqh.kotelService.Hooks()...))
	}
	client, err := kgo.NewClient(kopts...)
	if err != nil {
		return []error{}, err
	}
	defer client.Close()
	endOffsets, err := kadm.NewClient(client).ListEndOffsets(cctx, pqTopic)
	if err != nil {
		return []error{}, err
	}

	msgs := make([]*messagex.Message, 0)
	errs := make([]error, 0)
FLOOP:
	for {
		select {
		case <-cctx.Done():
			break FLOOP
		default:
		}
		tcctx, tcancel := context.WithTimeout(cctx, 5*time.Second)
		fetches := client.PollRecords(tcctx, 0)
		// Check if it's a FetchErr, by default PollRecords return a single fetch with a single topic and single partition holding the err
		if len(fetches) == 1 &&
			len(fetches[0].Topics) == 1 &&
			len(fetches[0].Topics[0].Partitions) == 1 &&
			fetches[0].Topics[0].Partitions[0].Err != nil {
			if fetches[0].Topics[0].Partitions[0].Err != tcctx.Err() {
				err = fetches[0].Topics[0].Partitions[0].Err
				pqh.l.WithError(err).Errorf("error fetches returned")
			} else {
				pqh.l.WithError(fetches[0].Topics[0].Partitions[0].Err).Infof("expected fetch error trigger consumption termination")
			}
			tcancel()
			break FLOOP
		}
		if len(fetches) == 0 {
			pqh.l.Infof("no fetches were found, assuming poison queue is empty")
			tcancel()
			break FLOOP
		}

		for _, f := range fetches {
			for _, t := range f.Topics {
				for _, p := range t.Partitions {
					pEndOffset := endOffsets[t.Topic][p.Partition].Offset
					if p.LogStartOffset >= pEndOffset {
						pqh.l.Infof("reached the end offset identified on consumption start '%v' >= '%v'", p.LogStartOffset, pEndOffset)
						tcancel()
						break FLOOP
					}
				}
			}
		}
		fetches.EachRecord(func(r *kgo.Record) {
			select {
			case <-cctx.Done():
				return
			default:
			}
			msg, localErr := defaultMarshaler.Unmarshal(r)
			msgs = append(msgs, msg)
			errs = append(errs, localErr)
		})
		tcancel()
	}
	if err != nil {
		return []error{}, err
	}
	if cctx.Err() != nil {
		return []error{}, cctx.Err()
	}
	handlerErrs, err := handler(ctx, msgs)
	if len(handlerErrs) != len(errs) {
		pqh.l.Errorf("errors output length mismatch, dismissing Unmarshal errs")
		return handlerErrs, err
	}
	for i := range handlerErrs {
		if errs[i] != nil {
			handlerErrs[i] = errs[i]
		}
	}
	return handlerErrs, err
}
