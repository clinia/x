package kafkax

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
)

// MessageHandler an message processor that is able to receive a ConsumerMessage
// and perform some task with it. Once consumed, if there is a session, it will the offset
// will be marked as processed.
type MessageHandler interface {
	// Setup is called at the beginning of a new session, before ConsumeClaim (and thus ProcessMessages) is called.
	// It may have a consumer group session in the case of a consumer group.
	Setup(*sarama.ConsumerGroupSession) error
	Cleanup(*sarama.ConsumerGroupSession) error
	ProcessMessages(
		ctx context.Context,
		kafkaMessages <-chan *sarama.ConsumerMessage,
		sess sarama.ConsumerGroupSession,
		messageLogFields watermill.LogFields,
	) error
}

type messageHandler struct {
	outputChannel chan<- *message.Message

	nackResendSleep time.Duration

	logger        watermill.LoggerAdapter
	closing       chan struct{}
	messageParser messageParser
}

// Cleanup implements MessageHandler.
func (h messageHandler) Cleanup(*sarama.ConsumerGroupSession) error {
	return nil
}

// Setup implements MessageHandler.
func (h messageHandler) Setup(*sarama.ConsumerGroupSession) error {
	return nil
}

func NewMessageHandler(
	outputChannel chan<- *message.Message,
	unmarshaler Unmarshaler,
	logger watermill.LoggerAdapter,
	closing chan struct{},
	nackResendSleep time.Duration,
) MessageHandler {
	return messageHandler{
		outputChannel: outputChannel,
		messageParser: messageParser{
			unmarshaler: unmarshaler,
		},
		nackResendSleep: nackResendSleep,
		logger:          logger,
		closing:         closing,
	}
}

func (h messageHandler) ProcessMessages(
	ctx context.Context,
	kafkaMessages <-chan *sarama.ConsumerMessage,
	sess sarama.ConsumerGroupSession,
	logFields watermill.LogFields,
) error {
	for {
		select {
		case kafkaMsg := <-kafkaMessages:
			if kafkaMsg == nil {
				h.logger.Debug("kafkaMsg is closed, stopping ProcessMessages", logFields)
				return nil
			}
			if err := h.processMessage(ctx, kafkaMsg, sess, logFields); err != nil {
				return err
			}
		case <-h.closing:
			h.logger.Debug("Subscriber is closing, ", logFields)
			return nil
		case <-ctx.Done():
			h.logger.Debug("Ctx was cancelled, ", logFields)
			return nil
		}
	}
}

func (h messageHandler) processMessage(
	ctx context.Context,
	kafkaMsg *sarama.ConsumerMessage,
	sess sarama.ConsumerGroupSession,
	messageLogFields watermill.LogFields,
) error {
	msgHolder, err := h.messageParser.prepareAndProcessMessage(ctx, kafkaMsg, h.logger, messageLogFields, sess)
	if err != nil {
		return err
	}

	msg := msgHolder.message
	receivedMsgLogFields := msgHolder.logFields
	ctx, cancelCtx := context.WithCancel(msg.Context())
	msg.SetContext(ctx)
	defer cancelCtx()

ResendLoop:
	for {
		select {
		case h.outputChannel <- msg:
			h.logger.Trace("Message sent to consumer", receivedMsgLogFields)
		case <-h.closing:
			h.logger.Trace("Closing, message discarded", receivedMsgLogFields)
			return nil
		case <-ctx.Done():
			h.logger.Trace("Closing, ctx cancelled before message was sent to consumer", receivedMsgLogFields)
			return nil
		}

		select {
		case <-msg.Acked():
			if sess != nil {
				sess.MarkMessage(kafkaMsg, "")
			}
			h.logger.Trace("Message Acked", receivedMsgLogFields)
			break ResendLoop
		case <-msg.Nacked():
			h.logger.Trace("Message Nacked", receivedMsgLogFields)

			// reset acks, etc.
			msg = msg.Copy()
			msg.SetContext(ctx)
			if h.nackResendSleep != NoSleep {
				time.Sleep(h.nackResendSleep)
			}

			continue ResendLoop
		case <-h.closing:
			h.logger.Trace("Closing, message discarded before ack", receivedMsgLogFields)
			return nil
		case <-ctx.Done():
			h.logger.Trace("Closing, ctx cancelled before ack", receivedMsgLogFields)
			return nil
		}
	}

	return nil
}

type messageHolder struct {
	kafkaMessage *sarama.ConsumerMessage
	message      *message.Message
	logFields    watermill.LogFields
	sess         sarama.ConsumerGroupSession
}

func (mh messageHolder) Copy() *messageHolder {
	msg := mh.message.Copy()
	msg.SetContext(addMessageContextFields(msg.Context(), mh.kafkaMessage))
	return &messageHolder{
		kafkaMessage: mh.kafkaMessage,
		message:      msg,
		logFields:    mh.logFields,
		sess:         mh.sess,
	}
}

func addMessageContextFields(ctx context.Context, kafkaMsg *sarama.ConsumerMessage) context.Context {
	result := setPartitionToCtx(ctx, kafkaMsg.Partition)
	result = setPartitionOffsetToCtx(result, kafkaMsg.Offset)
	result = setMessageTimestampToCtx(result, kafkaMsg.Timestamp)
	result = setMessageKeyToCtx(result, kafkaMsg.Key)
	return result
}

type messageParser struct {
	unmarshaler Unmarshaler
}

func (mp messageParser) prepareAndProcessMessage(
	ctx context.Context,
	kafkaMsg *sarama.ConsumerMessage,
	logger watermill.LoggerAdapter,
	messageLogFields watermill.LogFields,
	sess sarama.ConsumerGroupSession,
) (*messageHolder, error) {
	receivedMsgLogFields := messageLogFields.Add(watermill.LogFields{
		"kafka_partition_offset": kafkaMsg.Offset,
		"kafka_partition":        kafkaMsg.Partition,
	})
	logger.Trace("Received message from Kafka", receivedMsgLogFields)

	msg, err := mp.unmarshaler.Unmarshal(kafkaMsg)
	if err != nil {
		// resend will make no sense, stopping consumerGroupHandler
		return nil, errors.Wrap(err, "message unmarshal failed")
	}
	ctx = addMessageContextFields(ctx, kafkaMsg)
	msg.SetContext(ctx)
	receivedMsgLogFields = receivedMsgLogFields.Add(watermill.LogFields{
		"message_uuid": msg.UUID,
	})
	return &messageHolder{
		kafkaMessage: kafkaMsg,
		message:      msg,
		logFields:    receivedMsgLogFields,
		sess:         sess,
	}, nil
}

func waitForMessage(ctx context.Context, logger watermill.LoggerAdapter, holder *messageHolder, cancelctx context.CancelFunc) <-chan bool {
	waitChan := make(chan bool, 1)
	go func() {
		defer close(waitChan)
		defer cancelctx()
		select {
		case <-holder.message.Acked():
			logger.Trace("Message was ACKed", holder.logFields)
			waitChan <- true
			break
		case <-holder.message.Nacked():
			logger.Trace("Message was NACKed", holder.logFields)
			waitChan <- false
			break
		case <-ctx.Done():
			logger.Trace("Context has finished", holder.logFields)
			break
		}
	}()
	return waitChan
}
