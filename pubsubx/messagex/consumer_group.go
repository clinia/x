package messagex

import "fmt"

type ConsumerGroup string

// ConsumerGroup returns the consumer group name with the given scope.
// If the scope is empty, it returns the consumer group name as is.
// This should be used when interacting with the concrete pubsubs (e.g. Kafka).
func (group ConsumerGroup) ConsumerGroup(scope string) string {
	if scope != "" {
		return fmt.Sprintf("%s.%s", scope, group)
	}

	return string(group)
}
