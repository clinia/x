package kgox

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/clinia/x/pointerx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestKadminClient_CreateTopic(t *testing.T) {
	config := getPubsubConfig(t, true)
	getKadmClient := func(t *testing.T, defaultCreateTopicConfigs ...map[string]*string) (*KgoxAdminClient, *kadm.Client) {
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
		require.EventuallyWithT(t, func(t *assert.CollectT) {
			rConfigs, err := kadmCl.DescribeTopicConfigs(ctx, topic)
			assert.NoError(t, err)
			assert.Len(t, rConfigs, 1)
			assert.NoError(t, rConfigs[0].Err)
			assert.True(t, lo.ContainsBy(rConfigs[0].Configs, func(c kadm.Config) bool {
				return c.Key == "max.message.bytes" && c.Value != nil && *c.Value == maxBytes
			}), "expected max.message.bytes to be %s, but config entries are %v", maxBytes, rConfigs[0].Configs)
		}, 3*time.Second, 100*time.Millisecond)
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
		require.EventuallyWithT(t, func(t *assert.CollectT) {
			rConfigs, err := kadmCl.DescribeTopicConfigs(ctx, topic)
			assert.NoError(t, err)
			assert.Len(t, rConfigs, 1)
			assert.NoError(t, rConfigs[0].Err)
			assert.True(t, lo.ContainsBy(rConfigs[0].Configs, func(c kadm.Config) bool {
				return c.Key == "max.message.bytes" && c.Value != nil && *c.Value == maxBytesOneMB
			}), "expected max.message.bytes to be %s, but config entries are %v", maxBytes, rConfigs[0].Configs)
		}, 3*time.Second, 100*time.Millisecond)
	})

	t.Run("should DescribeTopicConfigs", func(t *testing.T) {
		kgoxAdmCl, _ := getKadmClient(t)
		_, err := kgoxAdmCl.DescribeTopicConfigs(ctx)
		require.NoError(t, err)
	})
}

func TestKadminClient_DeleteGroup(t *testing.T) {
	config := getPubsubConfig(t, true)
	getKadmClient := func(t *testing.T, defaultCreateTopicConfigs ...map[string]*string) (*KgoxAdminClient, *kadm.Client) {
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

	t.Run("test delete consumer group with retry topics", func(t *testing.T) {
		ctx := context.Background()
		kgoxAdmCl, kadmCl := getKadmClient(t)
		group, topics := getRandomGroupTopics(t, 1)
		retryTopic := topics[0].GenerateRetryTopic(messagex.ConsumerGroup(group))
		_, err := kadmCl.CreateTopics(ctx, 1, 1, map[string]*string{}, topics[0].TopicName(config.Scope))
		assert.NoError(t, err)
		defer func() {
			kadmCl.DeleteTopic(context.Background(), topics[0].TopicName(config.Scope))
		}()
		_, err = kadmCl.CreateTopics(ctx, 1, 1, map[string]*string{}, retryTopic.TopicName(config.Scope))
		assert.NoError(t, err)
		defer func() {
			kadmCl.DeleteTopic(context.Background(), retryTopic.TopicName(config.Scope))
		}()
		groupClient, err := kgo.NewClient(
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
			kgo.ConsumerGroup(group),
			kgo.ConsumeTopics(topics[0].TopicName(config.Scope), retryTopic.TopicName(config.Scope)),
		)
		time.Sleep(500 * time.Millisecond)
		groupClient.PollFetches(nil)
		time.Sleep(500 * time.Millisecond)
		startTopics, err := kadmCl.ListTopics(ctx)
		assert.Contains(t, startTopics.TopicsList().Topics(), retryTopic.TopicName(config.Scope))
		assert.NoError(t, err)
		startGroups, err := kadmCl.ListGroups(ctx)
		assert.Contains(t, startGroups.Groups(), group)
		assert.NoError(t, err)
		groupClient.Close()
		assert.NoError(t, err)

		_, err = kgoxAdmCl.DeleteGroup(ctx, group)

		assert.NoError(t, err)
		endTopics, err := kadmCl.ListTopics(ctx)
		assert.NoError(t, err)
		assert.NotContains(t, endTopics.TopicsList().Topics(), retryTopic.TopicName(config.Scope))
		endGroups, err := kadmCl.ListGroups(ctx)
		assert.NoError(t, err)
		assert.NotContains(t, endGroups.Groups(), group)
	})
}

func TestKadminClient_DeleteGroups(t *testing.T) {
	config := getPubsubConfig(t, true)
	getKadmClient := func(t *testing.T, defaultCreateTopicConfigs ...map[string]*string) (*KgoxAdminClient, *kadm.Client) {
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

	t.Run("test delete consumer groups with retry topics", func(t *testing.T) {
		ctx := context.Background()
		kgoxAdmCl, kadmCl := getKadmClient(t)
		group, topics := getRandomGroupTopics(t, 1)
		group2, _ := getRandomGroupTopics(t, 0)
		retryTopic := topics[0].GenerateRetryTopic(messagex.ConsumerGroup(group))
		retryTopic2 := topics[0].GenerateRetryTopic(messagex.ConsumerGroup(group2))
		_, err := kadmCl.CreateTopics(ctx, 1, 1, map[string]*string{}, topics[0].TopicName(config.Scope))
		assert.NoError(t, err)
		defer func() {
			kadmCl.DeleteTopic(context.Background(), topics[0].TopicName(config.Scope))
		}()
		_, err = kadmCl.CreateTopics(ctx, 1, 1, map[string]*string{}, retryTopic.TopicName(config.Scope))
		assert.NoError(t, err)
		defer func() {
			kadmCl.DeleteTopic(context.Background(), retryTopic.TopicName(config.Scope))
		}()
		_, err = kadmCl.CreateTopics(ctx, 1, 1, map[string]*string{}, retryTopic2.TopicName(config.Scope))
		assert.NoError(t, err)
		defer func() {
			kadmCl.DeleteTopic(context.Background(), retryTopic2.TopicName(config.Scope))
		}()
		groupClient, err := kgo.NewClient(
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
			kgo.ConsumerGroup(group),
			kgo.ConsumeTopics(topics[0].TopicName(config.Scope), retryTopic.TopicName(config.Scope)),
		)
		groupClient2, err := kgo.NewClient(
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
			kgo.ConsumerGroup(group2),
			kgo.ConsumeTopics(topics[0].TopicName(config.Scope), retryTopic.TopicName(config.Scope)),
		)
		time.Sleep(500 * time.Millisecond)
		groupClient.PollFetches(nil)
		groupClient2.PollFetches(nil)
		time.Sleep(500 * time.Millisecond)
		startTopics, err := kadmCl.ListTopics(ctx)
		assert.Contains(t, startTopics.TopicsList().Topics(), retryTopic.TopicName(config.Scope))
		assert.Contains(t, startTopics.TopicsList().Topics(), retryTopic2.TopicName(config.Scope))
		assert.NoError(t, err)
		startGroups, err := kadmCl.ListGroups(ctx)
		assert.Contains(t, startGroups.Groups(), group)
		assert.Contains(t, startGroups.Groups(), group2)
		assert.NoError(t, err)
		groupClient.Close()
		groupClient2.Close()
		assert.NoError(t, err)

		_, err = kgoxAdmCl.DeleteGroups(ctx, group, group2)

		assert.NoError(t, err)
		endTopics, err := kadmCl.ListTopics(ctx)
		assert.NoError(t, err)
		assert.NotContains(t, endTopics.TopicsList().Topics(), retryTopic.TopicName(config.Scope))
		assert.NotContains(t, endTopics.TopicsList().Topics(), retryTopic2.TopicName(config.Scope))
		endGroups, err := kadmCl.ListGroups(ctx)
		assert.NoError(t, err)
		assert.NotContains(t, endGroups.Groups(), group)
		assert.NotContains(t, endGroups.Groups(), group2)
	})
}
