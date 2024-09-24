package kgox

import (
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/segmentio/ksuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Marshaler interface {
	// Marshal marshals a message into a Kafka record.
	Marshal(m *messagex.Message, topic string) (*kgo.Record, error)
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

func (m *DefaultMarshaler) Marshal(msg *messagex.Message, topic string) (*kgo.Record, error) {
	headers := make([]kgo.RecordHeader, len(msg.Metadata))

	setIDHeader := false
	i := 0
	for k, v := range msg.Metadata {
		if k == messagex.IDHeaderKey {
			if msg.ID == "" {
				setIDHeader = true
				headers = append(headers, kgo.RecordHeader{
					Key:   messagex.IDHeaderKey,
					Value: []byte(v),
				})
				continue
			}
			// In the else case, we will set the ID header below.

			continue
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

	if !setIDHeader {
		headers = append(headers, kgo.RecordHeader{
			Key:   messagex.IDHeaderKey,
			Value: []byte(msg.ID),
		})
	}

	return &kgo.Record{
		Topic:   topic,
		Headers: headers,
		Value:   msg.Payload,
	}, nil
}

// Unmarshal implements Unmarshaler.
func (m *DefaultMarshaler) Unmarshal(r *kgo.Record) (*messagex.Message, error) {
	msg := &messagex.Message{
		Metadata: make(messagex.MessageMetadata, len(r.Headers)),
		Payload:  r.Value,
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
