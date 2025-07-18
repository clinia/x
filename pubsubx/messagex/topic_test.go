package messagex

import (
	"testing"

	"github.com/clinia/x/errorx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTopic(t *testing.T) {
	t.Run("should return new topic with valid topic name", func(t *testing.T) {
		topic, err := NewTopic("my-topic")
		assert.NoError(t, err)
		assert.Equal(t, "my-topic", string(topic))
	})

	t.Run("should not return error when topic contains dots", func(t *testing.T) {
		topic, err := NewTopic("my" + TopicSeparator + "topic")
		assert.NoError(t, err)
		assert.Equal(t, Topic("my.topic"), topic)
	})
}

func TestTopicName(t *testing.T) {
	t.Run("should return topic name with no scope", func(t *testing.T) {
		topic, err := NewTopic("my-topic")
		require.NoError(t, err)
		var scope string
		assert.Equal(t, "my-topic", topic.TopicName(scope))
	})

	t.Run("should return topic name with scope", func(t *testing.T) {
		topic, err := NewTopic("my-topic")
		require.NoError(t, err)
		assert.Equal(t, "scope"+TopicSeparator+"my-topic", topic.TopicName("scope"))
	})
}

func TestGenerateRetryTopic(t *testing.T) {
	t.Run("should return a retry topic for a specific consumer group", func(t *testing.T) {
		topic, err := NewTopic("my-topic")
		require.NoError(t, err)
		retryTopic := topic.GenerateRetryTopic(ConsumerGroup("group"))
		assert.Equal(t, "my-topic"+TopicSeparator+"group"+TopicRetrySuffix, string(retryTopic))
	})
}

func TestTopicFromName(t *testing.T) {
	t.Run("should return a topic extracted from a topic name without the scope", func(t *testing.T) {
		topic := TopicFromName("scope.my-topic")
		assert.Equal(t, Topic("my-topic"), topic)
	})

	t.Run("should return a topic extracted from a short topic name", func(t *testing.T) {
		topic := TopicFromName("my-topic")
		assert.Equal(t, Topic("my-topic"), topic)
	})

	t.Run("should return a topic extracted from a topic name without the scope and additional fields", func(t *testing.T) {
		topic := TopicFromName("scope.my-topic.interestingly.long")
		assert.Equal(t, Topic("my-topic.interestingly.long"), topic)
	})
}

func TestBaseTopicFromName(t *testing.T) {
	t.Run("should return a topic extracted from a topic name without the scope", func(t *testing.T) {
		topic := BaseTopicFromName("scope.my-topic")
		assert.Equal(t, Topic("my-topic"), topic)
	})

	t.Run("should return a topic extracted from a short topic name", func(t *testing.T) {
		topic := BaseTopicFromName("my-topic")
		assert.Equal(t, Topic("my-topic"), topic)
	})

	t.Run("should return a topic extracted from a topic name without the scope and additional fields", func(t *testing.T) {
		topic := BaseTopicFromName("scope.my-topic.interestingly.long")
		assert.Equal(t, Topic("my-topic.interestingly.long"), topic)
	})

	t.Run("should return a topic extracted from a retry topic name without scope", func(t *testing.T) {
		topic := BaseTopicFromName("scope.my-topic.interestingly.consumer-group" + TopicRetrySuffix)
		assert.Equal(t, Topic("my-topic.interestingly"), topic)
	})

	t.Run("should return a topic extracted from a short retry topic", func(t *testing.T) {
		topic := BaseTopicFromName("scope.my-topic.consumer-group" + TopicRetrySuffix)
		assert.Equal(t, Topic("my-topic"), topic)
	})

	t.Run("should return an empty topic extracted from a retry suffix only topic", func(t *testing.T) {
		topic := BaseTopicFromName("consumer-group" + TopicRetrySuffix)
		assert.Equal(t, Topic(""), topic)
	})

	t.Run("should return the original topic name when it contains dots", func(t *testing.T) {
		topic := BaseTopicFromName("scope.my-topic.interestingly.long")
		assert.Equal(t, Topic("my-topic.interestingly.long"), topic)
	})

	t.Run("should return the original topic name when it contains dots with retry suffix", func(t *testing.T) {
		topic := BaseTopicFromName("scope.my-topic.interestingly.long.consumer-group" + TopicRetrySuffix)
		assert.Equal(t, Topic("my-topic.interestingly.long"), topic)
	})
}

func TestTopicScopeUtils(t *testing.T) {
	t.Run("ExtractScopeFromTopic", func(t *testing.T) {
		for _, tc := range []struct {
			name          string
			topic         string
			expectedScope string
			expectedTopic Topic
			expectedErr   error
		}{
			{
				name:          "valid topic with scope",
				topic:         "scope.my-topic",
				expectedScope: "scope",
				expectedTopic: "my-topic",
			},
			{
				name:        "topic without scope",
				topic:       "my-topic",
				expectedErr: errorx.InvalidArgumentError("topic 'my-topic' does not have a valid format, expected at least scope and topic name"),
			},
			{
				name:          "valid topic with multiple segments",
				topic:         "scope.my-topic.interestingly.long",
				expectedScope: "scope",
				expectedTopic: "my-topic.interestingly.long",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				scope, topic, err := ExtractScopeFromTopic(tc.topic)
				if tc.expectedErr != nil {
					assert.EqualError(t, err, tc.expectedErr.Error())
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedScope, scope)
				assert.Equal(t, tc.expectedTopic, topic)
			})
		}
	})

	t.Run("RenameTopicWithScope", func(t *testing.T) {
		for _, tc := range []struct {
			name          string
			topic         string
			scope         string
			expectedTopic string
			expectedErr   error
		}{
			{
				name:          "valid topic with scope",
				topic:         "scope.my-topic",
				scope:         "new-scope",
				expectedTopic: "new-scope.my-topic",
			},
			{
				name:          "valid topic with multiple segments and scope",
				topic:         "oldScope.my-topic.interestingly.long",
				scope:         "my-new-scope",
				expectedTopic: "my-new-scope.my-topic.interestingly.long",
			},
			{
				name:        "topic without scope",
				topic:       "my-topic",
				scope:       "",
				expectedErr: errorx.InvalidArgumentError("topic 'my-topic' does not have a valid format, expected at least scope and topic name"),
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				newTopic, err := RenameTopicWithScope(tc.topic, tc.scope)
				if tc.expectedErr != nil {
					assert.EqualError(t, err, tc.expectedErr.Error())
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTopic, newTopic)
			})
		}
	})
}
