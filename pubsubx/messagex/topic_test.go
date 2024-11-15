package messagex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTopic(t *testing.T) {
	t.Run("should return new topic with valid topic name", func(t *testing.T) {
		topic, err := NewTopic("my-topic")
		assert.NoError(t, err)
		assert.Equal(t, "my-topic", string(topic))
	})

	t.Run("should return error with invalid topic name", func(t *testing.T) {
		topic, err := NewTopic("my" + topicSeparator + "topic")
		assert.Error(t, err)
		assert.Equal(t, Topic(""), topic)
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
		assert.Equal(t, "scope"+topicSeparator+"my-topic", topic.TopicName("scope"))
	})
}

func TestGenerateRetryTopic(t *testing.T) {
	t.Run("should return a retry topic for a specific consumer group", func(t *testing.T) {
		topic, err := NewTopic("my-topic")
		require.NoError(t, err)
		retryTopic := topic.GenerateRetryTopic(ConsumerGroup("group"))
		assert.Equal(t, "my-topic"+topicSeparator+"group"+retrySuffix, string(retryTopic))
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
		topic := BaseTopicFromName("scope.my-topic.interestingly.consumer-group." + retrySuffix)
		assert.Equal(t, Topic("my-topic.interestingly"), topic)
	})

	t.Run("should return a topic extracted from a short retry topic", func(t *testing.T) {
		topic := BaseTopicFromName("scope.my-topic.consumer-group." + retrySuffix)
		assert.Equal(t, Topic("my-topic"), topic)
	})

	t.Run("should return an empty topic extracted from a retry suffix only topic", func(t *testing.T) {
		topic := BaseTopicFromName("consumer-group." + retrySuffix)
		assert.Equal(t, Topic(""), topic)
	})
}
