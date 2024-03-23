package pubsubx

import "context"

type Admin interface {
	// CreateTopic creates a new topic.
	CreateTopic(ctx context.Context, topic string, detail *TopicDetail) error

	// ListTopics returns a map of topics to their details.
	ListTopics(ctx context.Context) (map[string]TopicDetail, error)

	// DeleteTopic deletes a topic.
	DeleteTopic(ctx context.Context, topic string) error

	// ListSubscribers returns a map of subscribers to their topics.
	ListSubscribers(ctx context.Context, topic string) (map[string]string, error)

	// DeleteSubscriber deletes a subscriber.
	DeleteSubscriber(ctx context.Context, subscriber string) error
}

type TopicDetail struct {
	NumPartitions     int32
	ReplicationFactor int16
	ReplicaAssignment map[int32][]int32
	ConfigEntries     map[string]*string
}
