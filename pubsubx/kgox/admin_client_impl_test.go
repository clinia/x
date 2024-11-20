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

func TestKadminClient_DeleteTopicsWithRetryTopics(t *testing.T) {
	config := getPubsubConfig(t, true)
	getKadmClient := func(t *testing.T, defaultCreateTopicConfigs ...map[string]*string) (*KgoxAdminClient, *kadm.Client, *kgo.Client) {
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
		return NewPubSubAdminClient(wc, config, defaultCreateTopicConfigEntries), kadmCl, wc
	}

	ctx := context.Background()

	t.Run("should delete a topic with retry topics", func(t *testing.T) {
		kgoxAdmCl, kadmCl, _ := getKadmClient(t)
		group, topic := getRandomGroupTopics(t, 1)
		topics := []string{topic[0].TopicName(config.Scope), topic[0].GenerateRetryTopic(messagex.ConsumerGroup(group)).TopicName(config.Scope)}
		//nolint:all
		resCreate, err := kadmCl.CreateTopics(ctx, 1, int16(len(config.Providers.Kafka.Brokers)), make(map[string]*string), topics...)
		require.NoError(t, err)
		defer func() {
			defer kadmCl.Close()
			kadmCl.DeleteTopics(context.Background(), topics...)
		}()
		for _, v := range resCreate {
			require.NoError(t, v.Err)
		}
		createdTopics := make([]string, len(resCreate))
		i := 0
		for k := range resCreate {
			createdTopics[i] = k
			i++
		}
		assert.NoError(t, err)
		require.NoError(t, err)
		for _, top := range topics {
			assert.Contains(t, createdTopics, top)
		}
		time.Sleep(time.Second)
		lt, err := kadmCl.ListTopics(ctx)
		require.NoError(t, err)
		for _, top := range topics {
			assert.Contains(t, lt.TopicsList().Topics(), top)
		}
		res, err := kgoxAdmCl.DeleteTopicWithRetryTopics(ctx, topic[0].TopicName(config.Scope))
		assert.NoError(t, err)
		assert.Equal(t, len(topics), len(res))
		for _, top := range topics {
			r, ok := res[top]
			assert.True(t, ok)
			assert.NoError(t, r.Err)
		}
		time.Sleep(time.Second)
		lt, err = kadmCl.ListTopics(ctx)
		require.NoError(t, err)
		for _, top := range topics {
			assert.NotContains(t, lt.TopicsList().Topics(), top)
		}
	})

	t.Run("should delete topics with retry topics", func(t *testing.T) {
		kgoxAdmCl, kadmCl, _ := getKadmClient(t)
		group, topic := getRandomGroupTopics(t, 2)
		topics := []string{topic[0].TopicName(config.Scope), topic[0].GenerateRetryTopic(messagex.ConsumerGroup(group)).TopicName(config.Scope), topic[1].TopicName(config.Scope), topic[1].GenerateRetryTopic(messagex.ConsumerGroup(group)).TopicName(config.Scope)}
		//nolint:all
		resCreate, err := kadmCl.CreateTopics(ctx, 1, int16(len(config.Providers.Kafka.Brokers)), make(map[string]*string), topics...)
		require.NoError(t, err)
		defer func() {
			defer kadmCl.Close()
			kadmCl.DeleteTopics(context.Background(), topics...)
		}()
		for _, v := range resCreate {
			require.NoError(t, v.Err)
		}
		createdTopics := make([]string, len(resCreate))
		i := 0
		for k := range resCreate {
			createdTopics[i] = k
			i++
		}
		assert.NoError(t, err)
		defer func() {
			defer kadmCl.Close()
			kadmCl.DeleteTopics(context.Background(), topics...)
		}()
		require.NoError(t, err)
		for _, top := range topics {
			assert.Contains(t, createdTopics, top)
		}
		time.Sleep(time.Second)
		lt, err := kadmCl.ListTopics(ctx)
		require.NoError(t, err)
		for _, top := range topics {
			assert.Contains(t, lt.TopicsList().Topics(), top)
		}
		res, err := kgoxAdmCl.DeleteTopicsWithRetryTopics(ctx, topic[0].TopicName(config.Scope), topic[1].TopicName(config.Scope))
		assert.NoError(t, err)
		assert.Equal(t, len(topics), len(res))
		for _, top := range topics {
			r, ok := res[top]
			assert.True(t, ok)
			assert.NoError(t, r.Err)
		}
		time.Sleep(time.Second)
		lt, err = kadmCl.ListTopics(ctx)
		require.NoError(t, err)
		for _, top := range topics {
			assert.NotContains(t, lt.TopicsList().Topics(), top)
		}
	})
}

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
		return NewPubSubAdminClient(wc, config, defaultCreateTopicConfigEntries), kadmCl
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
		return NewPubSubAdminClient(wc, config, defaultCreateTopicConfigEntries), kadmCl
	}

	t.Run("test delete consumer group with retry topics", func(t *testing.T) {
		ctx := context.Background()
		kgoxAdmCl, kadmCl := getKadmClient(t)
		group, topics := getRandomGroupTopics(t, 1)
		cGroup := messagex.ConsumerGroup(group)
		retryTopic := topics[0].GenerateRetryTopic(cGroup)
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
			kgo.ConsumerGroup(cGroup.ConsumerGroup(config.Scope)),
			kgo.ConsumeTopics(topics[0].TopicName(config.Scope), retryTopic.TopicName(config.Scope)),
		)
		assert.NoError(t, err)
		time.Sleep(500 * time.Millisecond)
		groupClient.ForceMetadataRefresh()
		time.Sleep(500 * time.Millisecond)
		startTopics, err := kadmCl.ListTopics(ctx)
		assert.NoError(t, err)
		assert.Contains(t, startTopics.TopicsList().Topics(), retryTopic.TopicName(config.Scope))
		assert.NoError(t, err)
		startGroups, err := kadmCl.ListGroups(ctx)
		assert.NoError(t, err)
		assert.Contains(t, startGroups.Groups(), cGroup.ConsumerGroup(config.Scope))
		assert.NoError(t, err)
		groupClient.Close()
		assert.NoError(t, err)

		res, err := kgoxAdmCl.DeleteGroup(ctx, cGroup)

		assert.NoError(t, err)
		assert.NoError(t, res.Err)
		endTopics, err := kadmCl.ListTopics(ctx)
		assert.NoError(t, err)
		assert.NotContains(t, endTopics.TopicsList().Topics(), retryTopic.TopicName(config.Scope))
		endGroups, err := kadmCl.ListGroups(ctx)
		assert.NoError(t, err)
		assert.NotContains(t, endGroups.Groups(), cGroup.ConsumerGroup(config.Scope))
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
		return NewPubSubAdminClient(wc, config, defaultCreateTopicConfigEntries), kadmCl
	}

	t.Run("test delete consumer groups with retry topics", func(t *testing.T) {
		ctx := context.Background()
		kgoxAdmCl, kadmCl := getKadmClient(t)
		group, topics := getRandomGroupTopics(t, 1)
		cGroup := messagex.ConsumerGroup(group)
		group2, _ := getRandomGroupTopics(t, 0)
		cGroup2 := messagex.ConsumerGroup(group2)
		retryTopic := topics[0].GenerateRetryTopic(cGroup)
		retryTopic2 := topics[0].GenerateRetryTopic(cGroup2)
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
			kgo.ConsumerGroup(cGroup.ConsumerGroup(config.Scope)),
			kgo.ConsumeTopics(topics[0].TopicName(config.Scope), retryTopic.TopicName(config.Scope)),
		)
		assert.NoError(t, err)
		groupClient2, err := kgo.NewClient(
			kgo.SeedBrokers(config.Providers.Kafka.Brokers...),
			kgo.ConsumerGroup(cGroup2.ConsumerGroup(config.Scope)),
			kgo.ConsumeTopics(topics[0].TopicName(config.Scope), retryTopic.TopicName(config.Scope)),
		)
		assert.NoError(t, err)
		time.Sleep(500 * time.Millisecond)
		groupClient.ForceMetadataRefresh()
		groupClient2.ForceMetadataRefresh()
		time.Sleep(500 * time.Millisecond)
		startTopics, err := kadmCl.ListTopics(ctx)
		assert.Contains(t, startTopics.TopicsList().Topics(), retryTopic.TopicName(config.Scope))
		assert.Contains(t, startTopics.TopicsList().Topics(), retryTopic2.TopicName(config.Scope))
		assert.NoError(t, err)
		startGroups, err := kadmCl.ListGroups(ctx)
		assert.Contains(t, startGroups.Groups(), cGroup.ConsumerGroup(config.Scope))
		assert.Contains(t, startGroups.Groups(), cGroup2.ConsumerGroup(config.Scope))
		assert.NoError(t, err)
		groupClient.Close()
		groupClient2.Close()
		assert.NoError(t, err)

		res, err := kgoxAdmCl.DeleteGroups(ctx, cGroup, cGroup2)

		assert.NoError(t, err)
		for _, dgr := range res {
			assert.NoError(t, dgr.Err)
		}
		endTopics, err := kadmCl.ListTopics(ctx)
		assert.NoError(t, err)
		assert.NotContains(t, endTopics.TopicsList().Topics(), retryTopic.TopicName(config.Scope))
		assert.NotContains(t, endTopics.TopicsList().Topics(), retryTopic2.TopicName(config.Scope))
		endGroups, err := kadmCl.ListGroups(ctx)
		assert.NoError(t, err)
		assert.NotContains(t, endGroups.Groups(), cGroup.ConsumerGroup(config.Scope))
		assert.NotContains(t, endGroups.Groups(), cGroup2.ConsumerGroup(config.Scope))
	})
}

func TestKadminClient_Healthcheck(t *testing.T) {
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
		return NewPubSubAdminClient(wc, config, defaultCreateTopicConfigEntries), kadmCl
	}
	t.Run("healthcheck should not return error", func(t *testing.T) {
		c, _ := getKadmClient(t)
		err := c.HealthCheck(context.Background())
		assert.NoError(t, err)
	})
}
