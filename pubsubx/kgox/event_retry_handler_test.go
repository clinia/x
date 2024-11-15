package kgox

import (
	"context"
	"testing"

	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRetryTopics(t *testing.T) {
	l := logrusx.New("test", "")
	config := getPubsubConfig(t, true)

	t.Run("should create retry topic for the topic list", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 4)
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		retryTopics, errs, err := erh.generateRetryTopics(context.Background(), topics...)
		t.Cleanup(func() {
			cl, err := erh.AdminClient()
			require.NoError(t, err)
			for _, rt := range retryTopics {
				cl.DeleteTopic(context.Background(), rt.TopicName(config.Scope))
			}
			cl.Close()
		})
		assert.NoError(t, err)
		assert.Empty(t, lo.Filter(errs, func(item error, _ int) bool { return item != nil }))
		assert.Equal(t, lo.Map(topics, func(t messagex.Topic, _ int) messagex.Topic {
			return t.GenerateRetryTopic(cg)
		}), retryTopics)
	})

	t.Run("should create non-existing retry topic for the topic list", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 4)
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		cl, err := erh.AdminClient()
		require.NoError(t, err)
		//nolint:all
		_, err = cl.CreateTopic(context.Background(), 1, int16(len(config.Providers.Kafka.Brokers)), string(topics[1].GenerateRetryTopic(cg)))
		assert.NoError(t, err)
		retryTopics, errs, err := erh.generateRetryTopics(context.Background(), topics...)
		t.Cleanup(func() {
			for _, rt := range retryTopics {
				cl.DeleteTopic(context.Background(), rt.TopicName(config.Scope))
			}
			cl.Close()
		})
		assert.NoError(t, err)
		assert.Empty(t, lo.Filter(errs, func(item error, _ int) bool { return item != nil }))
		assert.Equal(t, lo.Map(topics, func(t messagex.Topic, _ int) messagex.Topic {
			return t.GenerateRetryTopic(cg)
		}), retryTopics)
	})

	t.Run("should return only valid retry topics", func(t *testing.T) {
		group, topics := getRandomGroupTopics(t, 4)
		topics = append(topics, messagex.Topic("wrong name ~ `"))
		cg := messagex.ConsumerGroup(group)
		erh := getEventRetryHandler(t, l, config, cg, nil)
		retryTopics, errs, err := erh.generateRetryTopics(context.Background(), topics...)
		t.Cleanup(func() {
			cl, err := erh.AdminClient()
			require.NoError(t, err)
			for _, rt := range retryTopics {
				cl.DeleteTopic(context.Background(), rt.TopicName(config.Scope))
			}
			cl.Close()
		})
		assert.Error(t, err)
		remainingErrs := lo.Filter(errs, func(item error, _ int) bool { return item != nil })
		assert.Equal(t, 1, len(remainingErrs))
		assert.Equal(t, lo.Map(topics[:len(topics)-1], func(t messagex.Topic, _ int) messagex.Topic {
			return t.GenerateRetryTopic(cg)
		}), retryTopics)
		assert.Equal(t, len(topics)-1, len(retryTopics))
	})
}
