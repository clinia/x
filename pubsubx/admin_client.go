package pubsubx

import (
	"context"

	"github.com/clinia/x/pubsubx/messagex"
	"github.com/twmb/franz-go/pkg/kadm"
)

type PubSubAdminClient interface {
	// CreateTopic creates a topic with the given configuration.
	// The default configuration entries are set by default, but they can be overridden (see `pubsub.NewCreateTopicConfigEntries()`).
	CreateTopic(ctx context.Context, partitions int32, replicationFactor int16, topic string, configs ...map[string]*string) (kadm.CreateTopicResponse, error)
	// CreateTopics creates a topics with the given configuration.
	// The default configuration entries are set by default, but they can be overridden (see `pubsub.NewCreateTopicConfigEntries()`).
	CreateTopics(ctx context.Context, partitions int32, replicationFactor int16, topics []string, configs ...map[string]*string) (kadm.CreateTopicResponses, error)
	// DeleteTopic deletes a topic.
	DeleteTopic(ctx context.Context, topic string) (kadm.DeleteTopicResponse, error)
	// DeleteTopicsWithRetryTopics deletes the topics with their related retry topics.
	DeleteTopicsWithRetryTopics(ctx context.Context, topics ...string) (kadm.DeleteTopicResponses, error)
	// DeleteTopicWithRetryTopics deletes a topic with it's related retry topics.
	DeleteTopicWithRetryTopics(ctx context.Context, topic string) (kadm.DeleteTopicResponses, error)
	// DeleteGroup deletes a group and related resources.
	DeleteGroup(ctx context.Context, group messagex.ConsumerGroup) (kadm.DeleteGroupResponse, error)
	// DeleteGroups deletes groups and related resources.
	DeleteGroups(ctx context.Context, groups ...messagex.ConsumerGroup) (kadm.DeleteGroupResponses, error)
	// HealthCheck checks the health of the underlying pubsub. It returns an error if the pubsub is unhealthy or we cannot connect to it.
	HealthCheck(ctx context.Context) error

	// ListTopics returns the details of the given topics.
	// If no topics are provided, it returns the details of all topics.
	ListTopics(ctx context.Context, topics ...string) (kadm.TopicDetails, error)

	// TruncateTopicsWithRetryTopics deleted the records of the provided topics, as well as the corresponding retry topics,
	// by setting the offsets to the the end offsets
	TruncateTopicsWithRetryTopics(ctx context.Context, topics ...string) (kadm.DeleteRecordsResponses, error)

	Close()
}
