package inmemorypubsub

import (
	"context"
	"strings"

	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/twmb/franz-go/pkg/kadm"
)

type NoopAdminClient struct {
	topics kadm.TopicDetails
	*pubsubx.Config
}

var _ pubsubx.PubSubAdminClient = (*NoopAdminClient)(nil)

func NewNoopAdminClient(c *pubsubx.Config) pubsubx.PubSubAdminClient {
	return &NoopAdminClient{
		topics: make(kadm.TopicDetails),
		Config: c,
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

func (n *NoopAdminClient) DeleteTopicsWithRetryTopics(ctx context.Context, topics ...string) (kadm.DeleteTopicResponses, error) {
	res := make(kadm.DeleteTopicResponses)
	for _, t := range topics {
		for lt := range n.topics {
			if strings.HasPrefix(lt, t) && strings.HasSuffix(lt, messagex.TopicRetrySuffix) {
				delete(n.topics, lt)
				res[lt] = kadm.DeleteTopicResponse{
					Topic: t,
				}

			}
		}
		delete(n.topics, t)
		res[t] = kadm.DeleteTopicResponse{
			Topic: t,
		}
	}
	return res, nil
}

func (n *NoopAdminClient) DeleteTopicWithRetryTopics(ctx context.Context, topic string) (kadm.DeleteTopicResponses, error) {
	return n.DeleteTopicsWithRetryTopics(ctx, topic)
}

func (n *NoopAdminClient) DeleteGroup(ctx context.Context, group messagex.ConsumerGroup) (kadm.DeleteGroupResponse, error) {
	for t := range n.topics {
		if strings.HasSuffix(t, string(group)+messagex.TopicSeparator+messagex.TopicRetrySuffix) {
			delete(n.topics, t)
		}
	}
	return kadm.DeleteGroupResponse{
		Group: group.ConsumerGroup(n.Scope),
		Err:   nil,
	}, nil
}

// DeleteGroups deletes groups and related resources.
func (n *NoopAdminClient) DeleteGroups(ctx context.Context, groups ...messagex.ConsumerGroup) (kadm.DeleteGroupResponses, error) {
	res := make(kadm.DeleteGroupResponses)
	for _, g := range groups {
		gScoped := g.ConsumerGroup(n.Scope)
		for t := range n.topics {
			if strings.HasSuffix(t, string(g)+messagex.TopicSeparator+messagex.TopicRetrySuffix) {
				delete(n.topics, t)
			}
		}
		res[gScoped] = kadm.DeleteGroupResponse{
			Group: gScoped,
			Err:   nil,
		}
	}
	return res, nil
}

// HealthCheck implements pubsubx.PubSubAdminClient.
func (n *NoopAdminClient) HealthCheck(ctx context.Context) error {
	return nil
}

// ListTopics implements pubsubx.PubSubAdminClient.
func (n *NoopAdminClient) ListTopics(ctx context.Context, topics ...string) (kadm.TopicDetails, error) {
	return n.topics, nil
}

func (n *NoopAdminClient) TruncateTopicsWithRetryTopics(ctx context.Context, topics ...string) (kadm.DeleteRecordsResponses, error) {
	res := make(kadm.DeleteRecordsResponses)
	return res, nil
}
