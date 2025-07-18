package messagex

import (
	"strings"

	"github.com/clinia/x/errorx"
)

const ConsumerGroupSeparator = "."

type ConsumerGroup string

// ConsumerGroup returns the consumer group name with the given scope.
// If the scope is empty, it returns the consumer group name as is.
// This should be used when interacting with the concrete pubsubs (e.g. Kafka).
func (group ConsumerGroup) ConsumerGroup(scope string) string {
	if scope != "" {
		return scope + ConsumerGroupSeparator + string(group)
	}

	return string(group)
}

// ExtractScopeFromConsumerGroup extracts the scope and the consumer group name from the given consumer group string.
// It expects the format to be `{scope}.{group}`.
// If the scope is missing, it returns an error.
func ExtractScopeFromConsumerGroup(group string) (scope string, groupWithoutScope ConsumerGroup, err error) {
	splits := strings.Split(group, ConsumerGroupSeparator)
	if len(splits) < 2 {
		return "", "", errorx.InvalidArgumentErrorf("group '%s' does not have a valid format, expected at least scope and group name", group)
	}
	return splits[0], ConsumerGroup(strings.Join(splits[1:], ConsumerGroupSeparator)), nil
}

// RenameConsumerGroupWithScope renames the consumer group with the given scope.
// If the group does not have a valid format, it returns an error.
func RenameConsumerGroupWithScope(group string, scope string) (string, error) {
	_, withoutScope, err := ExtractScopeFromConsumerGroup(group)
	if err != nil {
		return "", err
	}
	return withoutScope.ConsumerGroup(scope), nil
}
