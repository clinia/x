package otelsaramax

import (
	"github.com/IBM/sarama"
)

type consumerGroupHandler struct {
	sarama.ConsumerGroupHandler
	consumerInfo ConsumerInfo

	cfg config
}

// ConsumeClaim wraps the session and claim to add instruments for messages.
// It implements parts of `ConsumerGroupHandler`.
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Wrap claim
	dispatcher := newConsumerMessagesDispatcherWrapper(claim, h.consumerInfo, h.cfg)
	go dispatcher.Run()
	claim = &consumerGroupClaim{
		ConsumerGroupClaim: claim,
		dispatcher:         dispatcher,
	}

	return h.ConsumerGroupHandler.ConsumeClaim(session, claim)
}

// WrapConsumerGroupHandler wraps a sarama.ConsumerGroupHandler causing each received
// message to be traced.
func WrapConsumerGroupHandler(handler sarama.ConsumerGroupHandler, consumerInfo ConsumerInfo, opts ...Option) sarama.ConsumerGroupHandler {
	cfg := newConfig(opts...)

	return &consumerGroupHandler{
		ConsumerGroupHandler: handler,
		consumerInfo:         consumerInfo,
		cfg:                  cfg,
	}
}

type consumerGroupClaim struct {
	sarama.ConsumerGroupClaim
	dispatcher consumerMessagesDispatcher
}

func (c *consumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	return c.dispatcher.Messages()
}
