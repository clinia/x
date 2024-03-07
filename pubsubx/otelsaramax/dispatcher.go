package otelsaramax

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Shopify/sarama"

	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type consumerMessagesDispatcher interface {
	Messages() <-chan *sarama.ConsumerMessage
}

type consumerMessagesDispatcherWrapper struct {
	d            consumerMessagesDispatcher
	consumerInfo ConsumerInfo
	messages     chan *sarama.ConsumerMessage

	cfg config
}

func newConsumerMessagesDispatcherWrapper(d consumerMessagesDispatcher, consumerInfo ConsumerInfo, cfg config) *consumerMessagesDispatcherWrapper {
	return &consumerMessagesDispatcherWrapper{
		d:            d,
		consumerInfo: consumerInfo,
		messages:     make(chan *sarama.ConsumerMessage),
		cfg:          cfg,
	}
}

// Messages returns the read channel for the messages that are returned by
// the broker.
func (w *consumerMessagesDispatcherWrapper) Messages() <-chan *sarama.ConsumerMessage {
	return w.messages
}

func (w *consumerMessagesDispatcherWrapper) Run() {
	msgs := w.d.Messages()

	for msg := range msgs {
		// Extract a span context from message to link.
		carrier := NewConsumerMessageCarrier(msg)
		parentSpanContext := w.cfg.Propagators.Extract(context.Background(), carrier)

		// Create a span.
		attrs := append(w.consumerInfo.Attributes,
			semconv.MessagingSystem("kafka"),
			semconv.MessagingDestinationKindTopic,
			semconv.MessagingDestinationName(msg.Topic),
			semconv.MessagingOperationReceive,
			semconv.MessagingMessageID(strconv.FormatInt(msg.Offset, 10)),
			semconv.MessagingKafkaSourcePartition(int(msg.Partition)),
		)
		opts := []trace.SpanStartOption{
			trace.WithAttributes(attrs...),
			trace.WithSpanKind(trace.SpanKindConsumer),
		}
		newCtx, span := w.cfg.Tracer.Start(parentSpanContext, fmt.Sprintf("%s consumerGroup receive", w.consumerInfo.ConsumerGroup), opts...)

		// Inject current span context, so consumers can use it to propagate span.
		w.cfg.Propagators.Inject(newCtx, carrier)

		// Send messages back to user.
		w.messages <- msg

		span.End()
	}
	close(w.messages)
}
