package messagex

import (
	"testing"

	"github.com/clinia/x/featureflagx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewMessageWithFeatureFlags(t *testing.T) {
	Flags := []featureflagx.FeatureFlag{
		featureflagx.FeatureFlag("feature1"),
		featureflagx.FeatureFlag("feature2"),
		featureflagx.FeatureFlag("feature3"),
		featureflagx.FeatureFlag("feature4"),
	}
	ffm := map[string]bool{
		"feature1": true,
		"feature2": false,
		"feature3": true,
	}

	t.Run("should not create a new message with feature flags if nil", func(t *testing.T) {
		mWithFF := NewMessage(
			[]byte("TestPayload"),
			WithFeatureFlags(nil),
		)
		assert.Equal(t, &Message{
			ID: mWithFF.ID,
			Metadata: MessageMetadata{
				"_clinia_retry_count": "0",
			},
			Payload: []byte("TestPayload"),
		}, mWithFF)
	})
	t.Run("should create a new message with feature flags", func(t *testing.T) {
		ff, err := featureflagx.New(ffm, Flags)
		require.NoError(t, err)
		require.NotNil(t, ff)
		mWithFF := NewMessage(
			[]byte("TestPayload"),
			WithFeatureFlags(ff),
		)
		assert.Equal(t, &Message{
			ID: mWithFF.ID,
			Metadata: MessageMetadata{
				"_clinia_retry_count":   "0",
				"_clinia_feature_flags": "{\"feature1\":true,\"feature2\":false,\"feature3\":true,\"feature4\":false}",
			},
			Payload: []byte("TestPayload"),
		}, mWithFF)
	})

	t.Run("should extract feature flags from message", func(t *testing.T) {
		m := &Message{
			ID: "1",
			Metadata: MessageMetadata{
				"_clinia_retry_count":   "0",
				"_clinia_feature_flags": "{\"feature1\":true,\"feature2\":false,\"feature3\":true,\"feature4\":false}",
			},
			Payload: []byte("TestPayload"),
		}
		ff, err := featureflagx.New(ffm, Flags)
		require.NoError(t, err)
		require.NotNil(t, ff)

		ffFromMessage, ok := m.ExtractFeatureFlags()
		assert.True(t, ok)
		assert.Equal(t, ff, ffFromMessage)
	})

	t.Run("should handle missing or empty feature flags gracefully", func(t *testing.T) {
		testCases := []struct {
			name     string
			message  *Message
			expected bool
		}{
			{
				name: "missing feature flags",
				message: &Message{
					ID: "1",
					Metadata: MessageMetadata{
						"_clinia_retry_count": "0",
					},
					Payload: []byte("TestPayload"),
				},
				expected: false,
			},
			{
				name: "empty feature flags",
				message: &Message{
					ID: "1",
					Metadata: MessageMetadata{
						"_clinia_retry_count":   "0",
						"_clinia_feature_flags": "",
					},
					Payload: []byte("TestPayload"),
				},
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ffFromMessage, ok := tc.message.ExtractFeatureFlags()
				assert.Equal(t, tc.expected, ok)
				assert.Nil(t, ffFromMessage)
			})
		}
	})
	// t.Run("should panic", func(t *testing.T) {
	// 	testCases := []struct {
	// 		name     string
	// 		message  *Message
	// 		expected bool
	// 	}{
	// 		{
	// 			name: "value not boolean",
	// 			message: &Message{
	// 				ID: "1",
	// 				Metadata: MessageMetadata{
	// 					"_clinia_retry_count":   "0",
	// 					"_clinia_feature_flags": "{\"feature1\":1,\"feature2\":false,\"feature3\":true}",
	// 				},
	// 				Payload: []byte("TestPayload"),
	// 			},
	// 			expected: false,
	// 		},
	// 	}

	// 	for _, tc := range testCases {
	// 		t.Run(tc.name, func(t *testing.T) {
	// 			ctx := context.Background()
	// 			assert.PanicsWithError(t, "failed to unmarshal feature flags: json: cannot unmarshal number into Go value of type bool", func() {
	// 				tc.message.NewContextWithFeatureFlags(ctx)
	// 			})
	// 		})
	// 	}
	// })

}
