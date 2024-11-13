package messagex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	t.Run("should deep copy message", func(t *testing.T) {
		originalMessage := &Message{
			ID: "myId",
			Metadata: MessageMetadata{
				"test1": "value1",
				"test2": "value2",
			},
			Payload: []byte("TestPayload"),
		}

		copyMessage := originalMessage.Copy()
		assert.Equal(t, originalMessage, copyMessage)
		assert.False(t, originalMessage == copyMessage)
		assert.Equal(t, originalMessage.ID, copyMessage.ID)
		assert.Equal(t, originalMessage.Metadata, copyMessage.Metadata)
		assert.Equal(t, originalMessage.Payload, copyMessage.Payload)

		copyMessage.ID = "newId"
		copyMessage.Metadata["test1"] = "test2"
		copyMessage.Payload[2] = byte(5)

		assert.NotEqual(t, originalMessage, copyMessage)
		assert.NotEqual(t, originalMessage.ID, copyMessage.ID)
		assert.NotEqual(t, originalMessage.Metadata, copyMessage.Metadata)
		assert.NotEqual(t, originalMessage.Payload, copyMessage.Payload)
	})
}
