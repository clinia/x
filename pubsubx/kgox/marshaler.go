package kgox

import (
	"context"

	"github.com/clinia/x/pubsubx/messagex"
	"github.com/segmentio/ksuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Marshaler interface {
	// Marshal marshals a message into a Kafka record.
	Marshal(ctx context.Context, m *messagex.Message, topic string) (*kgo.Record, error)
}

type Unmarshaler interface {
	// Unmarshal unmarshals a Kafka record into a message.
	Unmarshal(r *kgo.Record) (*messagex.Message, error)
}

type DefaultMarshaler struct{}

var (
	_                Marshaler   = (*DefaultMarshaler)(nil)
	_                Unmarshaler = (*DefaultMarshaler)(nil)
	defaultMarshaler             = &DefaultMarshaler{}
)

func (m *DefaultMarshaler) Marshal(ctx context.Context, msg *messagex.Message, topic string) (*kgo.Record, error) {
	headers := make([]kgo.RecordHeader, len(msg.Metadata))

	setIDHeader := true
	setDefaultRetryCountHeader := true
	i := 0
	for k, v := range msg.Metadata {
		if k == messagex.IDHeaderKey {
			setIDHeader = false
			if msg.ID != "" {
				headers[i] = kgo.RecordHeader{
					Key:   messagex.IDHeaderKey,
					Value: []byte(msg.ID),
				}
				i++
				continue
			}
		}
		if k == messagex.RetryCountHeaderKey {
			setDefaultRetryCountHeader = false
		}
		headers[i] = kgo.RecordHeader{
			Key:   k,
			Value: []byte(v),
		}
		i++
	}

	if msg.ID == "" {
		msg.ID = ksuid.New().String()
	}

	if setIDHeader {
		headers = append(headers, kgo.RecordHeader{
			Key:   messagex.IDHeaderKey,
			Value: []byte(msg.ID),
		})
	}

	// Extract TraceContext from the message metadata
	ctx = msg.ExtractTraceContext(ctx)

	if setDefaultRetryCountHeader {
		headers = append(headers, kgo.RecordHeader{
			Key:   messagex.RetryCountHeaderKey,
			Value: []byte("0"),
		})
	}

	return &kgo.Record{
		Context: ctx,
		Topic:   topic,
		Headers: headers,
		Value:   msg.Payload,
		Key:     msg.Key,
	}, nil
}

// Unmarshal implements Unmarshaler.
func (m *DefaultMarshaler) Unmarshal(r *kgo.Record) (*messagex.Message, error) {
	msg := &messagex.Message{
		Metadata: make(messagex.MessageMetadata, len(r.Headers)),
		Payload:  r.Value,
		Offset:   r.Offset,
	}

	for _, header := range r.Headers {
		if header.Key == messagex.IDHeaderKey {
			msg.ID = string(header.Value)
			continue
		}

		msg.Metadata[header.Key] = string(header.Value)
	}

	return msg, nil
}
