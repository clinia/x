package pubsubx

import (
	"context"

	"github.com/twmb/franz-go/pkg/kadm"
)

type PubSubAdminClient interface {
	// CreateTopic creates a topic with the given configuration.
	// The default configuration entries are set by default, but they can be overridden (see `pubsub.NewCreateTopicConfigEntries()`).
	CreateTopic(ctx context.Context, partitions int32, replicationFactor int16, topic string, configs ...map[string]*string) (kadm.CreateTopicResponse, error)
	// DeleteTopic deletes a topic.
	DeleteTopic(ctx context.Context, topic string) (kadm.DeleteTopicResponse, error)
	// HealthCheck checks the health of the underlying pubsub. It returns an error if the pubsub is unhealthy or we cannot connect to it.
	HealthCheck(ctx context.Context) error

	// ListTopics returns the details of the given topics.
	// If no topics are provided, it returns the details of all topics.
	ListTopics(ctx context.Context, topics ...string) (kadm.TopicDetails, error)

	Close()
}
