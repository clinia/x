package kgox

import (
	"context"
	"testing"

	"github.com/clinia/x/pubsubx/messagex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshal(t *testing.T) {
	t.Run("should marshal message into record with root ID overwrite metadata", func(t *testing.T) {
		topic := "topic_name"
		msg := &messagex.Message{
			ID: "messageId",
			Metadata: messagex.MessageMetadata{
				messagex.IDHeaderKey:         "messageIdHeader",
				messagex.RetryCountHeaderKey: "0",
				"random_header":              "random",
			},
			Payload: []byte("PAYLOAD"),
		}
		record, err := defaultMarshaler.Marshal(context.Background(), msg, topic)
		assert.NoError(t, err)
		require.NotNil(t, record)
		assert.Equal(t, topic, record.Topic)
		assert.Equal(t, msg.Payload, record.Value)
		assert.Equal(t, len(msg.Metadata), len(record.Headers))
		for _, h := range record.Headers {
			if h.Key == "" {
				continue
			}
			if h.Key == messagex.IDHeaderKey {
				assert.Equal(t, msg.ID, string(h.Value))
				continue
			}
			msgMetadata, ok := msg.Metadata[h.Key]
			assert.True(t, ok)
			if ok {
				assert.Equal(t, msgMetadata, string(h.Value))
			}
		}
	})

	t.Run("should marshal message into record without root ID", func(t *testing.T) {
		topic := "topic_name"
		msg := &messagex.Message{
			Metadata: messagex.MessageMetadata{
				messagex.IDHeaderKey:         "messageIdHeader",
				messagex.RetryCountHeaderKey: "0",
				"random_header":              "random",
			},
			Payload: []byte("PAYLOAD"),
		}
		record, err := defaultMarshaler.Marshal(context.Background(), msg, topic)
		assert.NoError(t, err)
		require.NotNil(t, record)
		assert.Equal(t, topic, record.Topic)
		assert.Equal(t, msg.Payload, record.Value)
		assert.Equal(t, len(msg.Metadata), len(record.Headers))
		for _, h := range record.Headers {
			if h.Key == "" {
				continue
			}
			msgMetadata, ok := msg.Metadata[h.Key]
			assert.True(t, ok)
			if ok {
				assert.Equal(t, msgMetadata, string(h.Value))
			}
		}
	})

	t.Run("should marshal message into record with the root ID", func(t *testing.T) {
		topic := "topic_name"
		msg := &messagex.Message{
			ID: "messageId",
			Metadata: messagex.MessageMetadata{
				messagex.RetryCountHeaderKey: "0",
				"random_header":              "random",
			},
			Payload: []byte("PAYLOAD"),
		}
		record, err := defaultMarshaler.Marshal(context.Background(), msg, topic)
		assert.NoError(t, err)
		require.NotNil(t, record)
		assert.Equal(t, topic, record.Topic)
		assert.Equal(t, msg.Payload, record.Value)
		assert.Equal(t, len(msg.Metadata)+1, len(record.Headers))
		for _, h := range record.Headers {
			if h.Key == messagex.IDHeaderKey {
				assert.Equal(t, msg.ID, string(h.Value))
				continue
			}
			msgMetadata, ok := msg.Metadata[h.Key]
			assert.True(t, ok)
			if ok {
				assert.Equal(t, msgMetadata, string(h.Value))
			}
		}
	})

	t.Run("should marshal message into record with generated id", func(t *testing.T) {
		topic := "topic_name"
		msg := &messagex.Message{
			Metadata: messagex.MessageMetadata{
				messagex.RetryCountHeaderKey: "0",
				"random_header":              "random",
			},
			Payload: []byte("PAYLOAD"),
		}
		record, err := defaultMarshaler.Marshal(context.Background(), msg, topic)
		assert.NoError(t, err)
		require.NotNil(t, record)
		assert.Equal(t, topic, record.Topic)
		assert.Equal(t, msg.Payload, record.Value)
		assert.Equal(t, len(msg.Metadata)+1, len(record.Headers))
		exist := false
		for _, h := range record.Headers {
			if h.Key == messagex.IDHeaderKey {
				exist = true
				continue
			}
			msgMetadata, ok := msg.Metadata[h.Key]
			assert.True(t, ok)
			if ok {
				assert.Equal(t, msgMetadata, string(h.Value))
			}
		}
		assert.True(t, exist)
	})
}
