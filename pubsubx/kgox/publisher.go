package kgox

import (
	"context"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/twmb/franz-go/pkg/kgo"
)

type publisher PubSub

var _ pubsubx.Publisher = (*publisher)(nil)

// BulkPublish implements pubsubx.Publisher.
func (p *publisher) PublishSync(ctx context.Context, topic messagex.Topic, messages ...*messagex.Message) (pubsubx.Errors, error) {
	records, errs, err := p.marshalMessages(topic, messages...)
	if err != nil {
		return errs, err
	}

	produceResults := p.writeClient.ProduceSync(ctx, records...)
	if len(errs) != len(produceResults) {
		// Should not happen as marshalMessages return a slice of errors with the same size as the message
		errs = make(pubsubx.Errors, len(produceResults))
	}

	for i, result := range produceResults {
		if result.Err == nil {
			errs[i] = nil
			continue
		}
		if result.Record == nil {
			errs[i] = errorx.InternalErrorf("failed to produce message: %v", result.Err)
			continue
		}

		msg, err := defaultMarshaler.Unmarshal(result.Record)
		if err != nil {
			errs[i] = errorx.InternalErrorf("failed to unmarshal message: %v", err)
			continue
		}
		errs[i] = errorx.InternalErrorf("failed to produce message '%v': %v", msg.Metadata, result.Err)
	}

	return errs, errs.FirstNonNil()
}

// Close implements pubsubx.Publisher.
func (p *publisher) Close() error {
	p.writeClient.Close()
	return nil
}

// Publish implements pubsubx.Publisher.
func (p *publisher) PublishAsync(ctx context.Context, topic messagex.Topic, messages ...*messagex.Message) error {
	records, errs, err := p.marshalMessages(topic, messages...)
	if err != nil {
		return errs.Join()
	}

	for _, record := range records {
		p.writeClient.Produce(ctx, record, nil)
	}

	return nil
}

func (p *publisher) marshalMessages(topic messagex.Topic, messages ...*messagex.Message) ([]*kgo.Record, pubsubx.Errors, error) {
	scopedTopic := topic.TopicName(p.conf.Scope)
	out := make([]*kgo.Record, len(messages))
	errs := make(pubsubx.Errors, len(messages))
	for i, m := range messages {
		out[i], errs[i] = defaultMarshaler.Marshal(m, scopedTopic)
	}

	return out, errs, errs.FirstNonNil()
}
