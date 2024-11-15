package messagex

import (
	"fmt"
	"strings"
)

type Topic string

const (
	topicSeparator = "."
	retrySuffix    = "retry"
)

func NewTopic(topic string) (Topic, error) {
	if strings.Contains(string(topic), topicSeparator) {
		return "", fmt.Errorf("topic name cannot contain '.'")
	}

	return Topic(topic), nil
}

// TopicName returns the topic name with the given scope.
// If the scope is empty, it returns the topic name as is.
// This should be used when interacting with the concrete pubsubs (e.g. Kafka).
func (t Topic) TopicName(scope string) string {
	if scope != "" {
		return scope + topicSeparator + string(t)
	}

	return string(t)
}

func (t Topic) GenerateRetryTopic(consumerGroup ConsumerGroup) Topic {
	return Topic(string(t) + topicSeparator + string(consumerGroup) + topicSeparator + retrySuffix)
}

func TopicFromName(topicName string) Topic {
	splits := strings.Split(topicName, topicSeparator)
	if len(splits) > 1 {
		return Topic(strings.Join(splits[1:], topicSeparator))
	}

	return Topic(splits[0])
}

// Expect topic to be format `{scope}.{topic}.{consumer-group}.retry` or `{scope}.{topic}`
// If the scope is missing, this function will return a wrong result
func BaseTopicFromName(topicName string) Topic {
	splits := strings.Split(topicName, topicSeparator)
	if len(splits) > 1 {
		if strings.HasSuffix(splits[len(splits)-1], retrySuffix) {
			if len(splits) > 2 {
				return Topic(strings.Join(splits[1:len(splits)-2], topicSeparator))
			}
			// Should not happen, this is the use case where the topic and the scope or not included
			return Topic(strings.Join(splits[1:len(splits)-1], topicSeparator))
		}
		return Topic(strings.Join(splits[1:], topicSeparator))
	}

	return Topic(splits[0])
}
