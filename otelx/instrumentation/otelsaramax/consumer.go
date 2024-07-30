package otelsaramax

import (
	"github.com/IBM/sarama"
)

type partitionConsumer struct {
	sarama.PartitionConsumer
	dispatcher consumerMessagesDispatcher
}

// Messages returns the read channel for the messages that are returned by
// the broker.
func (pc *partitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return pc.dispatcher.Messages()
}

// WrapPartitionConsumer wraps a sarama.PartitionConsumer causing each received
// message to be traced.
func WrapPartitionConsumer(pc sarama.PartitionConsumer, consumerInfo ConsumerInfo, opts ...Option) sarama.PartitionConsumer {
	cfg := newConfig(opts...)

	dispatcher := newConsumerMessagesDispatcherWrapper(pc, consumerInfo, cfg)
	go dispatcher.Run()
	wrapped := &partitionConsumer{
		PartitionConsumer: pc,
		dispatcher:        dispatcher,
	}
	return wrapped
}

type consumer struct {
	sarama.Consumer
	consumerInfo ConsumerInfo

	opts []Option
}

// ConsumePartition invokes Consumer.ConsumePartition and wraps the resulting
// PartitionConsumer.
func (c *consumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	pc, err := c.Consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return nil, err
	}
	return WrapPartitionConsumer(pc, c.consumerInfo, c.opts...), nil
}

// WrapConsumer wraps a sarama.Consumer wrapping any PartitionConsumer created
// via Consumer.ConsumePartition.
func WrapConsumer(c sarama.Consumer, consumerInfo ConsumerInfo, opts ...Option) sarama.Consumer {
	return &consumer{
		Consumer:     c,
		consumerInfo: consumerInfo,
		opts:         opts,
	}
}
