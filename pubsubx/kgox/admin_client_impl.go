package kgox

import (
	"context"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pubsubx"
	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kadm"
)

type KgoxAdminClient struct {
	*kadm.Client
	defaultCreateTopicConfigEntries map[string]*string
}

var _ pubsubx.PubSubAdminClient = (*KgoxAdminClient)(nil)

func NewPubSubAdminClient(cl *kadm.Client, defaultCreateTopicConfigEntries map[string]*string) *KgoxAdminClient {
	return &KgoxAdminClient{cl, defaultCreateTopicConfigEntries}
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

// HealthCheck implements pubsubx.PubSubAdminClient.
func (p *KgoxAdminClient) HealthCheck(ctx context.Context) error {
	_, err := p.ListBrokers(ctx)
	if err != nil {
		return errorx.InternalErrorf("failed to connect to pubsub: %v", err)
	}

	return nil
}
