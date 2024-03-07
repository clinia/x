package otelsaramax

import "go.opentelemetry.io/otel/attribute"

// ConsumerInfo holds information about the consumer group.
type ConsumerInfo struct {
	ConsumerGroup string
	Attributes    []attribute.KeyValue
}
