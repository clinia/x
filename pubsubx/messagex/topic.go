package messagex

import (
	"fmt"
	"strings"
)

type Topic string

const topicSeparator = "."

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

func TopicFromName(topicName string) Topic {
	splits := strings.Split(topicName, topicSeparator)
	if len(splits) > 1 {
		// Should never happen, but just in case.
		return Topic(strings.Join(splits[1:], topicSeparator))
	}

	return Topic(splits[1])
}
