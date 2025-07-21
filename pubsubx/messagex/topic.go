package messagex

import (
	"strings"

	"github.com/clinia/x/errorx"
)

type Topic string

const (
	TopicSeparator   = "."
	TopicRetrySuffix = TopicSeparator + "retry"
)

func NewTopic(topic string) (Topic, error) {
	return Topic(topic), nil
}

// TopicName returns the topic name with the given scope.
// If the scope is empty, it returns the topic name as is.
// This should be used when interacting with the concrete pubsubs (e.g. Kafka).
func (t Topic) TopicName(scope string) string {
	if scope != "" {
		return scope + TopicSeparator + string(t)
	}

	return string(t)
}

func (t Topic) GenerateRetryTopic(consumerGroup ConsumerGroup) Topic {
	return Topic(string(t) + TopicSeparator + string(consumerGroup) + TopicRetrySuffix)
}

func TopicFromName(topicName string) Topic {
	splits := strings.Split(topicName, TopicSeparator)
	if len(splits) > 1 {
		return Topic(strings.Join(splits[1:], TopicSeparator))
	}

	return Topic(splits[0])
}

// Expect topic to be format `{scope}.{topic}.{consumer-group}.retry` or `{scope}.{topic}`
// If the scope is missing, this function will return a wrong result
func BaseTopicFromName(topicName string) Topic {
	splits := strings.Split(topicName, TopicSeparator)
	if len(splits) > 1 {
		if strings.HasSuffix(topicName, TopicRetrySuffix) {
			if len(splits) > 2 {
				return Topic(strings.Join(splits[1:len(splits)-2], TopicSeparator))
			}
			// Should not happen, this is the use case where the topic and the scope or not included
			return Topic(strings.Join(splits[1:len(splits)-1], TopicSeparator))
		}
		return Topic(strings.Join(splits[1:], TopicSeparator))
	}

	return Topic(splits[0])
}

// ExtractScopeFromTopic extracts the scope and topic from a topic string.
// It expects the topic to be in the format `{scope}.{topic}`.
// If the topic does not have a valid format, it returns an error.
func ExtractScopeFromTopic(topic string) (scope string, topicWithoutScope Topic, err error) {
	splits := strings.Split(topic, TopicSeparator)
	if len(splits) < 2 {
		return "", "", errorx.InvalidArgumentErrorf("topic '%s' does not have a valid format, expected at least scope and topic name", topic)
	}
	return splits[0], Topic(strings.Join(splits[1:], TopicSeparator)), nil
}

// RenameTopicWithScope renames a topic by replacing the existing scope with the given scope.
func RenameTopicWithScope(topic string, scope string) (string, error) {
	_, withoutScope, err := ExtractScopeFromTopic(topic)
	if err != nil {
		return "", err
	}
	return withoutScope.TopicName(scope), nil
}
