package kgox

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/clinia/x/pointerx"
	"github.com/clinia/x/pubsubx"
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
	ctx := context.Background()
	t.Run("should delete a topic with retry topics", func(t *testing.T) {
		kgoxAdmCl, kadmCl, _ := getKadmClient(t, config)
		group, topic := getRandomGroupTopics(t, 1)
		topics := []string{topic[0].TopicName(config.Scope), topic[0].GenerateRetryTopic(messagex.ConsumerGroup(group)).TopicName(config.Scope)}
		//nolint:all
		resCreate, err := kadmCl.CreateTopics(ctx, 1, int16(len(config.Providers.Kafka.Brokers)), make(map[string]*string), topics...)
		require.NoError(t, err)
		defer func() {
			defer kadmCl.Close()
			_, _ = kadmCl.DeleteTopics(context.Background(), topics...)
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
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			lt, err := kadmCl.ListTopics(ctx)
			require.NoError(c, err)
			for _, top := range topics {
				assert.Contains(c, lt.TopicsList().Topics(), top)
			}
		}, 2*time.Second, 250*time.Millisecond)
		res, err := kgoxAdmCl.DeleteTopicWithRetryTopics(ctx, topic[0].TopicName(config.Scope))
		assert.NoError(t, err)
		assert.Equal(t, len(topics), len(res))
		for _, top := range topics {
			r, ok := res[top]
			assert.True(t, ok)
			assert.NoError(t, r.Err)
		}
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			lt, err := kadmCl.ListTopics(ctx)
			require.NoError(c, err)
			for _, top := range topics {
				assert.NotContains(c, lt.TopicsList().Topics(), top)
			}
		}, 2*time.Second, 250*time.Millisecond)
	})

	t.Run("should delete topics with retry topics", func(t *testing.T) {
		kgoxAdmCl, kadmCl, _ := getKadmClient(t, config)
		group, topic := getRandomGroupTopics(t, 2)
		topics := []string{topic[0].TopicName(config.Scope), topic[0].GenerateRetryTopic(messagex.ConsumerGroup(group)).TopicName(config.Scope), topic[1].TopicName(config.Scope), topic[1].GenerateRetryTopic(messagex.ConsumerGroup(group)).TopicName(config.Scope)}
		//nolint:all
		resCreate, err := kadmCl.CreateTopics(ctx, 1, int16(len(config.Providers.Kafka.Brokers)), make(map[string]*string), topics...)
		require.NoError(t, err)
		defer func() {
			defer kadmCl.Close()
			_, _ = kadmCl.DeleteTopics(context.Background(), topics...)
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
			_, _ = kadmCl.DeleteTopics(context.Background(), topics...)
		}()
		require.NoError(t, err)
		for _, top := range topics {
			assert.Contains(t, createdTopics, top)
		}
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			lt, err := kadmCl.ListTopics(ctx)
			require.NoError(c, err)
			for _, top := range topics {
				assert.Contains(c, lt.TopicsList().Topics(), top)
			}
		}, 2*time.Second, 250*time.Millisecond)
		res, err := kgoxAdmCl.DeleteTopicsWithRetryTopics(ctx, topic[0].TopicName(config.Scope), topic[1].TopicName(config.Scope))
		assert.NoError(t, err)
		assert.Equal(t, len(topics), len(res))
		for _, top := range topics {
			r, ok := res[top]
			assert.True(t, ok)
			assert.NoError(t, r.Err)
		}
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			lt, err := kadmCl.ListTopics(ctx)
			require.NoError(c, err)
			for _, top := range topics {
				assert.NotContains(c, lt.TopicsList().Topics(), top)
			}
		}, 2*time.Second, 250*time.Millisecond)
	})
}

func TestKadminClient_CreateTopic(t *testing.T) {
	config := getPubsubConfig(t, true)
	ctx := context.Background()

	t.Run("should create a topic with no specific configs", func(t *testing.T) {
		kgoxAdmCl, kadmCl, _ := getKadmClient(t, config)
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
		kgoxAdmCl, kadmCl, _ := getKadmClient(t, config, map[string]*string{
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
		kgoxAdmCl, kadmCl, _ := getKadmClient(t, config, map[string]*string{
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
		kgoxAdmCl, _, _ := getKadmClient(t, config)
		_, err := kgoxAdmCl.DescribeTopicConfigs(ctx)
		require.NoError(t, err)
	})
}

func TestKadminClient_DeleteGroup(t *testing.T) {
	config := getPubsubConfig(t, true)

	t.Run("test delete consumer group with retry topics", func(t *testing.T) {
		ctx := context.Background()
		kgoxAdmCl, kadmCl, _ := getKadmClient(t, config)
		group, topics := getRandomGroupTopics(t, 1)
		cGroup := messagex.ConsumerGroup(group)
		retryTopic := topics[0].GenerateRetryTopic(cGroup)
		_, err := kadmCl.CreateTopics(ctx, 1, 1, map[string]*string{}, topics[0].TopicName(config.Scope))
		assert.NoError(t, err)
		t.Cleanup(func() {
			_, _ = kadmCl.DeleteTopic(context.Background(), topics[0].TopicName(config.Scope))
		})
		_, err = kadmCl.CreateTopics(ctx, 1, 1, map[string]*string{}, retryTopic.TopicName(config.Scope))
		assert.NoError(t, err)
		t.Cleanup(func() {
			_, _ = kadmCl.DeleteTopic(context.Background(), retryTopic.TopicName(config.Scope))
		})
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
		// TODO: test with a eventually (polling)
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

	t.Run("test delete consumer groups with retry topics", func(t *testing.T) {
		ctx := context.Background()
		kgoxAdmCl, kadmCl, _ := getKadmClient(t, config)
		group, topics := getRandomGroupTopics(t, 1)
		cGroup := messagex.ConsumerGroup(group)
		group2, _ := getRandomGroupTopics(t, 0)
		cGroup2 := messagex.ConsumerGroup(group2)
		retryTopic := topics[0].GenerateRetryTopic(cGroup)
		retryTopic2 := topics[0].GenerateRetryTopic(cGroup2)
		_, err := kadmCl.CreateTopics(ctx, 1, 1, map[string]*string{}, topics[0].TopicName(config.Scope))
		assert.NoError(t, err)
		defer func() {
			_, _ = kadmCl.DeleteTopic(context.Background(), topics[0].TopicName(config.Scope))
		}()
		_, err = kadmCl.CreateTopics(ctx, 1, 1, map[string]*string{}, retryTopic.TopicName(config.Scope))
		assert.NoError(t, err)
		defer func() {
			_, _ = kadmCl.DeleteTopic(context.Background(), retryTopic.TopicName(config.Scope))
		}()
		_, err = kadmCl.CreateTopics(ctx, 1, 1, map[string]*string{}, retryTopic2.TopicName(config.Scope))
		assert.NoError(t, err)
		defer func() {
			_, _ = kadmCl.DeleteTopic(context.Background(), retryTopic2.TopicName(config.Scope))
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
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			startTopics, err := kadmCl.ListTopics(ctx)
			assert.Contains(c, startTopics.TopicsList().Topics(), retryTopic.TopicName(config.Scope))
			assert.Contains(c, startTopics.TopicsList().Topics(), retryTopic2.TopicName(config.Scope))
			assert.NoError(c, err)
			startGroups, err := kadmCl.ListGroups(ctx)
			assert.Contains(c, startGroups.Groups(), cGroup.ConsumerGroup(config.Scope))
			assert.Contains(c, startGroups.Groups(), cGroup2.ConsumerGroup(config.Scope))
			assert.NoError(c, err)
		}, 2*time.Second, 250*time.Millisecond)
		groupClient.Close()
		groupClient2.Close()
		assert.NoError(t, err)

		res, err := kgoxAdmCl.DeleteGroups(ctx, cGroup, cGroup2)

		assert.NoError(t, err)
		for _, dgr := range res {
			assert.NoError(t, dgr.Err)
		}
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			endTopics, err := kadmCl.ListTopics(ctx)
			assert.NoError(c, err)
			assert.NotContains(c, endTopics.TopicsList().Topics(), retryTopic.TopicName(config.Scope))
			assert.NotContains(c, endTopics.TopicsList().Topics(), retryTopic2.TopicName(config.Scope))
			endGroups, err := kadmCl.ListGroups(ctx)
			assert.NoError(c, err)
			assert.NotContains(c, endGroups.Groups(), cGroup.ConsumerGroup(config.Scope))
			assert.NotContains(c, endGroups.Groups(), cGroup2.ConsumerGroup(config.Scope))
		}, 2*time.Second, 250*time.Millisecond)
	})
}

func TestKadminClient_Healthcheck(t *testing.T) {
	config := getPubsubConfig(t, true)

	t.Run("healthcheck should not return error", func(t *testing.T) {
		c, _, _ := getKadmClient(t, config)
		err := c.HealthCheck(context.Background())
		assert.NoError(t, err)
	})
}

func TestTruncateTopicsWithRetryTopics(t *testing.T) {
	l := getLogger()
	config := getPubsubConfig(t, false)
	ctx := context.Background()
	p := getPublisher(t, l, config)
	kgoxAdmCl, kadmCl, _ := getKadmClient(t, config)

	msgBatch := []*messagex.Message{}
	msgBatchCount := 100
	for i := range msgBatchCount {
		msgBatch = append(msgBatch, messagex.NewMessage([]byte(fmt.Sprintf("test-%d", i))))
	}

	_, topics := getRandomGroupTopics(t, 2)

	var retryTopics []messagex.Topic
	for _, topic := range topics {
		// create topic and publish messages to it
		createTopic(t, config, topic)
		errs, err := p.PublishSync(context.Background(), topic, msgBatch...)
		require.NoError(t, err)
		require.NoError(t, errs.FirstNonNil())

		// craete corresponding retry topic and publish messages to it
		retryTopic := topic.GenerateRetryTopic(messagex.ConsumerGroup("test"))
		retryTopics = append(retryTopics, retryTopic)
		createTopic(t, config, retryTopic)

		errs, err = p.PublishSync(context.Background(), retryTopic, msgBatch...)
		require.NoError(t, err)
		require.NoError(t, errs.FirstNonNil())

	}

	topicsAsStrings := lo.Map(topics, func(t messagex.Topic, _ int) string {
		return t.TopicName(config.Scope)
	})

	retryTopicsAsStrings := lo.Map(retryTopics, func(t messagex.Topic, _ int) string {
		return t.TopicName(config.Scope)
	})

	offsetsBefore, err := kadmCl.ListStartOffsets(ctx, topicsAsStrings...)
	require.NoError(t, err)

	offsetsBefore.Offsets().Each(func(o kadm.Offset) {
		assert.Equal(t, o.At, int64(0), "offset should be 0")
	})

	retryOffsetsBefore, err := kadmCl.ListStartOffsets(ctx, retryTopicsAsStrings...)
	require.NoError(t, err)

	retryOffsetsBefore.Offsets().Each(func(o kadm.Offset) {
		assert.Equal(t, o.At, int64(0), "offset should be 0")
	})

	// truncate only the first topic
	resp, err := kgoxAdmCl.TruncateTopicsWithRetryTopics(ctx, topicsAsStrings[0])
	require.NoError(t, err)

	assert.Equal(t, resp, []pubsubx.TruncateReponse{
		{
			Topic:        topicsAsStrings[0],
			OffsetBefore: 0,
			OffsetAfter:  100,
			Err:          nil,
		},
		{
			Topic:        retryTopicsAsStrings[0],
			OffsetBefore: 0,
			OffsetAfter:  100,
			Err:          nil,
		},
	})

	t.Run("should truncate topic", func(t *testing.T) {
		offsetsAfter, err := kadmCl.ListStartOffsets(ctx, topicsAsStrings[0])
		require.NoError(t, err)

		offsetsAfter.Offsets().Each(func(o kadm.Offset) {
			assert.Equal(t, o.At, int64(100), "offset should be 100 for truncated topic %s", o.Topic)
		})
	})

	t.Run("should truncate corresponding retry topic", func(t *testing.T) {
		offsetsAfter, err := kadmCl.ListStartOffsets(ctx, retryTopicsAsStrings[0])
		require.NoError(t, err)

		offsetsAfter.Offsets().Each(func(o kadm.Offset) {
			assert.Equal(t, o.At, int64(100), "offset should be 100 for truncated retry topic %s", o.Topic)
		})
	})

	t.Run("should keep untruncated topic and corresponding retry topic untouched", func(t *testing.T) {
		offsetsAfter, err := kadmCl.ListStartOffsets(ctx, topicsAsStrings[1])
		require.NoError(t, err)

		offsetsAfter.Offsets().Each(func(o kadm.Offset) {
			assert.Equal(t, o.At, int64(0), "offset should be 0 for untruncated topic %s", o.Topic)
		})

		offsetsAfter, err = kadmCl.ListStartOffsets(ctx, retryTopicsAsStrings[1])
		require.NoError(t, err)

		offsetsAfter.Offsets().Each(func(o kadm.Offset) {
			assert.Equal(t, o.At, int64(0), "offset should be 0 for untruncated topic %s", o.Topic)
		})
	})
}
