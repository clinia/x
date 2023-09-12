package pubsubx

import (
	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
)

type OTELSaramaTracer struct {
	opts []otelsarama.Option
}

func NewOTELSaramaTracer(option ...otelsarama.Option) kafka.SaramaTracer {
	return &OTELSaramaTracer{
		opts: option,
	}
}

func (t OTELSaramaTracer) WrapConsumer(c sarama.Consumer) sarama.Consumer {
	return otelsarama.WrapConsumer(c, t.opts...)
}

func (t OTELSaramaTracer) WrapConsumerGroupHandler(h sarama.ConsumerGroupHandler) sarama.ConsumerGroupHandler {
	return otelsarama.WrapConsumerGroupHandler(h, t.opts...)
}

func (t OTELSaramaTracer) WrapPartitionConsumer(pc sarama.PartitionConsumer) sarama.PartitionConsumer {
	return otelsarama.WrapPartitionConsumer(pc, t.opts...)
}

func (t OTELSaramaTracer) WrapSyncProducer(cfg *sarama.Config, p sarama.SyncProducer) sarama.SyncProducer {
	return otelsarama.WrapSyncProducer(cfg, p, t.opts...)
}
