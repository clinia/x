package kgox

import (
	"context"
	"fmt"
	"testing"

	"github.com/clinia/x/pointerx"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestKadminClient_CreateTopic(t *testing.T) {
	getKadmClient := func(t *testing.T, defaultCreateTopicConfigs ...map[string]*string) (*KgoxAdminClient, *kadm.Client) {
		config := getPubsubConfig(t)
		wc, err := kgo.NewClient(
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
		)
		require.NoError(t, err)
		t.Cleanup(wc.Close)
		var defaultCreateTopicConfigEntries map[string]*string
		if len(defaultCreateTopicConfigs) > 0 {
			defaultCreateTopicConfigEntries = lo.Assign(defaultCreateTopicConfigs...)
		}

		kadmCl := kadm.NewClient(wc)
		return NewPubSubAdminClient(kadmCl, defaultCreateTopicConfigEntries), kadmCl
	}

	ctx := context.Background()

	t.Run("should create a topic with no specific configs", func(t *testing.T) {
		kgoxAdmCl, kadmCl := getKadmClient(t)
		topic := fmt.Sprintf("test-topic-%s", ksuid.New().String())
		resp, err := kgoxAdmCl.CreateTopic(ctx, 1, 1, topic)
		require.NoError(t, err)
		require.Equal(t, topic, resp.Topic)
		t.Cleanup(func() {
			res, err := kadmCl.DeleteTopics(ctx, topic)
			require.NoError(t, err)
			require.NoError(t, res.Error())
		})

		// Check if the topic exists
		tDetails, err := kadmCl.ListTopics(ctx, topic)
		require.NoError(t, err)
		require.Len(t, tDetails, 1)
		require.Equal(t, topic, tDetails[topic].Topic)
	})

	t.Run("should create a topic with specific configs", func(t *testing.T) {
		maxBytes := "10485760" // 10MB
		kgoxAdmCl, kadmCl := getKadmClient(t, map[string]*string{
			"max.message.bytes": pointerx.Ptr(maxBytes),
		})
		topic := fmt.Sprintf("test-topic-%s", ksuid.New().String())
		resp, err := kgoxAdmCl.CreateTopic(ctx, 1, 1, topic)
		require.NoError(t, err)
		require.Equal(t, topic, resp.Topic)
		t.Cleanup(func() {
			res, err := kadmCl.DeleteTopics(ctx, topic)
			require.NoError(t, err)
			require.NoError(t, res.Error())
		})

		// Check if the topic exists
		rConfigs, err := kadmCl.DescribeTopicConfigs(ctx, topic)
		require.NoError(t, err)
		require.Len(t, rConfigs, 1)
		require.True(t, lo.ContainsBy(rConfigs[0].Configs, func(c kadm.Config) bool {
			return c.Key == "max.message.bytes" && c.Value != nil && *c.Value == maxBytes
		}), "expected max.message.bytes to be %s, but config entries are %v", maxBytes, rConfigs[0].Configs)
	})

	t.Run("should create a topic with specific configs and overrides", func(t *testing.T) {
		maxBytes := "10485760"     // 10MB
		maxBytesOneMB := "1048576" // 1MB
		kgoxAdmCl, kadmCl := getKadmClient(t, map[string]*string{
			"max.message.bytes": pointerx.Ptr(maxBytes),
		})
		topic := fmt.Sprintf("test-topic-%s", ksuid.New().String())
		resp, err := kgoxAdmCl.CreateTopic(ctx, 1, 1, topic, map[string]*string{
			"max.message.bytes": pointerx.Ptr(maxBytesOneMB), // 1MB
		})
		require.NoError(t, err)
		require.Equal(t, topic, resp.Topic)
		t.Cleanup(func() {
			res, err := kadmCl.DeleteTopics(ctx, topic)
			require.NoError(t, err)
			require.NoError(t, res.Error())
		})

		// Check if the topic exists
		rConfigs, err := kadmCl.DescribeTopicConfigs(ctx, topic)
		require.NoError(t, err)
		require.Len(t, rConfigs, 1)
		require.True(t, lo.ContainsBy(rConfigs[0].Configs, func(c kadm.Config) bool {
			return c.Key == "max.message.bytes" && c.Value != nil && *c.Value == maxBytesOneMB
		}))
	})

	t.Run("should DescribeTopicConfigs", func(t *testing.T) {
		kgoxAdmCl, _ := getKadmClient(t)
		_, err := kgoxAdmCl.DescribeTopicConfigs(ctx)
		require.NoError(t, err)
	})
}
