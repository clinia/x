package kafkax

import (
	"github.com/IBM/sarama"
	"github.com/clinia/x/otelx/instrumentation/otelsaramax"
)

type SaramaTracer interface {
	WrapConsumer(sarama.Consumer, otelsaramax.ConsumerInfo) sarama.Consumer
	// WrapPartitionConsumer(sarama.PartitionConsumer) sarama.PartitionConsumer
	WrapConsumerGroupHandler(sarama.ConsumerGroupHandler, otelsaramax.ConsumerInfo) sarama.ConsumerGroupHandler
	WrapSyncProducer(*sarama.Config, sarama.SyncProducer) sarama.SyncProducer
}

type OTELSaramaTracer struct {
	opts []otelsaramax.Option
}

func NewOTELSaramaTracer(option ...otelsaramax.Option) SaramaTracer {
	return &OTELSaramaTracer{
		opts: option,
	}
}

func (t OTELSaramaTracer) WrapConsumer(c sarama.Consumer, consumerInfo otelsaramax.ConsumerInfo) sarama.Consumer {
	return otelsaramax.WrapConsumer(c, consumerInfo, t.opts...)
}

func (t OTELSaramaTracer) WrapConsumerGroupHandler(h sarama.ConsumerGroupHandler, consumerInfo otelsaramax.ConsumerInfo) sarama.ConsumerGroupHandler {
	return otelsaramax.WrapConsumerGroupHandler(h, consumerInfo, t.opts...)
}

// func (t OTELSaramaTracer) WrapPartitionConsumer(pc sarama.PartitionConsumer) sarama.PartitionConsumer {
// 	return otelsaramax.WrapPartitionConsumer(pc)
// }

func (t OTELSaramaTracer) WrapSyncProducer(cfg *sarama.Config, p sarama.SyncProducer) sarama.SyncProducer {
	return otelsaramax.WrapSyncProducer(cfg, p, t.opts...)
}
