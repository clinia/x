package kgox

import (
	"context"
	"errors"
	"strings"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pubsubx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kadm"
)

type KgoxAdminClient struct {
	*kadm.Client
	*pubsubx.Config
	defaultCreateTopicConfigEntries map[string]*string
}

var _ pubsubx.PubSubAdminClient = (*KgoxAdminClient)(nil)

func NewPubSubAdminClient(cl *kadm.Client, config *pubsubx.Config, defaultCreateTopicConfigEntries map[string]*string) *KgoxAdminClient {
	return &KgoxAdminClient{cl, config, defaultCreateTopicConfigEntries}
}

// CreateTopic implements PubSubAdminClient.
// Subtle: this method shadows the method (*Client).CreateTopic of pubsubAdminClient.Client.
func (p *KgoxAdminClient) CreateTopic(ctx context.Context, partitions int32, replicationFactor int16, topic string, configs ...map[string]*string) (kadm.CreateTopicResponse, error) {
	configMaps := append([]map[string]*string{p.defaultCreateTopicConfigEntries}, configs...)
	configMaps = lo.Filter(configMaps, func(m map[string]*string, i int) bool {
		return m != nil
	})
	conf := lo.Assign(configMaps...)
	return p.Client.CreateTopic(ctx, partitions, replicationFactor, conf, topic)
}

// CreateTopics implements PubSubAdminClient.
// Subtle: this method shadows the method (*Client).CreateTopics of pubsubAdminClient.Client.
func (p *KgoxAdminClient) CreateTopics(ctx context.Context, partitions int32, replicationFactor int16, topics []string, configs ...map[string]*string) (kadm.CreateTopicResponses, error) {
	configMaps := append([]map[string]*string{p.defaultCreateTopicConfigEntries}, configs...)
	configMaps = lo.Filter(configMaps, func(m map[string]*string, i int) bool {
		return m != nil
	})
	conf := lo.Assign(configMaps...)
	return p.Client.CreateTopics(ctx, partitions, replicationFactor, conf, topics...)
}

// HealthCheck implements pubsubx.PubSubAdminClient.
func (p *KgoxAdminClient) HealthCheck(ctx context.Context) error {
	_, err := p.ListBrokers(ctx)
	if err != nil {
		return errorx.InternalErrorf("failed to connect to pubsub: %v", err)
	}

	return nil
}

// DeleteGroup implements PubSubAdminClient.
// Subtle: this method shadows the method (*Client).DeleteGroup of pubsubAdminClient.Client.
func (p *KgoxAdminClient) DeleteGroup(ctx context.Context, group messagex.ConsumerGroup) (kadm.DeleteGroupResponse, error) {
	r, err := p.Client.DeleteGroup(ctx, group.ConsumerGroup(p.Scope))
	if err != nil {
		return r, err
	}
	rt, err := p.ListTopics(ctx)
	if err != nil {
		return r, err
	}
	retryTopics := lo.Filter(rt.TopicsList().Topics(), func(topic string, _ int) bool {
		return strings.HasSuffix(topic, messagex.TopicSeparator+string(group)+messagex.TopicRetrySuffix)
	})
	deleteResponses, err := p.DeleteTopics(ctx, retryTopics...)
	if err != nil {
		return r, err
	}
	deleteErrs := make([]error, 0, len(deleteResponses))
	for _, v := range deleteResponses {
		if v.Err != nil {
			deleteErrs = append(deleteErrs, v.Err)
		}
	}
	err = errors.Join(deleteErrs...)
	if err != nil {
		return r, err
	}

	return r, nil
}

func (p *KgoxAdminClient) DeleteTopicsWithRetryTopics(ctx context.Context, topics ...string) (kadm.DeleteTopicResponses, error) {
	rt, err := p.ListTopics(ctx)
	if err != nil {
		return kadm.DeleteTopicResponses{}, err
	}
	topicsToDelete := lo.Filter(rt.TopicsList().Topics(), func(t string, _ int) bool {
		for _, topic := range topics {
			if strings.HasPrefix(t, topic) && strings.HasSuffix(t, messagex.TopicRetrySuffix) {
				return true
			}
		}
		return false
	})
	topicsToDelete = append(topicsToDelete, topics...)
	return p.DeleteTopics(ctx, topicsToDelete...)
}

func (p *KgoxAdminClient) DeleteTopicWithRetryTopics(ctx context.Context, topic string) (kadm.DeleteTopicResponses, error) {
	return p.DeleteTopicsWithRetryTopics(ctx, topic)
}

// DeleteGroups implements PubSubAdminClient.
// Subtle: this method shadows the method (*Client).DeleteGroups of pubsubAdminClient.Client.
func (p *KgoxAdminClient) DeleteGroups(ctx context.Context, groups ...messagex.ConsumerGroup) (kadm.DeleteGroupResponses, error) {
	r, err := p.Client.DeleteGroups(ctx, lo.Map(groups, func(g messagex.ConsumerGroup, _ int) string { return g.ConsumerGroup(p.Scope) })...)
	if err != nil {
		return r, err
	}
	rt, err := p.ListTopics(ctx)
	if err != nil {
		// AdminClient doesn't allow you to create a group back, this is by design
		return r, err
	}
	retryTopics := lo.Filter(rt.TopicsList().Topics(), func(topic string, _ int) bool {
		for _, group := range groups {
			if strings.HasSuffix(topic, messagex.TopicSeparator+string(group)+messagex.TopicRetrySuffix) {
				return true
			}
		}
		return false
	})
	_, err = p.DeleteTopics(ctx, retryTopics...)
	if err != nil {
		return r, err
	}

	return r, nil
}
