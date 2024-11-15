package inmemorypubsub

import (
	"context"

	"github.com/clinia/x/pubsubx"
	"github.com/twmb/franz-go/pkg/kadm"
)

type NoopAdminClient struct {
	topics kadm.TopicDetails
}

var _ pubsubx.PubSubAdminClient = (*NoopAdminClient)(nil)

func NewNoopAdminClient() pubsubx.PubSubAdminClient {
	return &NoopAdminClient{
		topics: make(kadm.TopicDetails),
	}
}

// Close implements pubsubx.PubSubAdminClient.
func (n *NoopAdminClient) Close() {
	// noop
}

// CreateTopic implements pubsubx.PubSubAdminClient.
func (n *NoopAdminClient) CreateTopic(ctx context.Context, partitions int32, replicationFactor int16, topic string, configs ...map[string]*string) (kadm.CreateTopicResponse, error) {
	pDetails := kadm.PartitionDetails{}
	for i := range partitions {
		pDetails[i] = kadm.PartitionDetail{
			Topic:     topic,
			Partition: i,
			Replicas:  make([]int32, replicationFactor),
		}
	}
	n.topics[topic] = kadm.TopicDetail{
		Topic:      topic,
		Partitions: pDetails,
	}

	return kadm.CreateTopicResponse{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}, nil
}

// CreateTopic implements pubsubx.PubSubAdminClient.
func (n *NoopAdminClient) CreateTopics(ctx context.Context, partitions int32, replicationFactor int16, topics []string, configs ...map[string]*string) (kadm.CreateTopicResponses, error) {
	pDetails := kadm.PartitionDetails{}
	res := make(kadm.CreateTopicResponses)
	for _, t := range topics {
		for i := range partitions {
			pDetails[i] = kadm.PartitionDetail{
				Topic:     t,
				Partition: i,
				Replicas:  make([]int32, replicationFactor),
			}
		}
		n.topics[t] = kadm.TopicDetail{
			Topic:      t,
			Partitions: pDetails,
		}
		res[t] = kadm.CreateTopicResponse{
			Topic:             t,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		}
	}

	return res, nil
}

// DeleteTopic implements pubsubx.PubSubAdminClient.
func (n *NoopAdminClient) DeleteTopic(ctx context.Context, topic string) (kadm.DeleteTopicResponse, error) {
	delete(n.topics, topic)
	return kadm.DeleteTopicResponse{
		Topic: topic,
	}, nil
}

// HealthCheck implements pubsubx.PubSubAdminClient.
func (n *NoopAdminClient) HealthCheck(ctx context.Context) error {
	return nil
}

// ListTopics implements pubsubx.PubSubAdminClient.
func (n *NoopAdminClient) ListTopics(ctx context.Context, topics ...string) (kadm.TopicDetails, error) {
	return n.topics, nil
}
