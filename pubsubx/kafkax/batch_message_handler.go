package kafkax

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/clinia/x/timerx"
)

// batchedMessageHandler works by fetching up to N messages from the provided channel
// waiting until maxWaitTime.
// It then takes the collected KafkaMessages and pushes them in order to outputChannel.
// When all have been ACKed or NACKed, it updates the offsets with the highest ACKed
// for each involved partition.
type batchedMessageHandler struct {
	outputChannel chan<- *message.Message

	maxBatchSize int16
	maxWaitTime  time.Duration

	nackResendSleep time.Duration

	logger        watermill.LoggerAdapter
	closing       chan struct{}
	messageParser messageParser
	messages      chan *messageHolder
	wg            sync.WaitGroup
	cancel        context.CancelFunc
}

func NewBatchedMessageHandler(
	outputChannel chan<- *message.Message,
	unmarshaler Unmarshaler,
	logger watermill.LoggerAdapter,
	closing chan struct{},
	maxBatchSize int16,
	maxWaitTime time.Duration,
	nackResendSleep time.Duration,
) MessageHandler {
	handler := &batchedMessageHandler{
		outputChannel:   outputChannel,
		maxBatchSize:    maxBatchSize,
		maxWaitTime:     maxWaitTime,
		nackResendSleep: nackResendSleep,
		closing:         closing,
		logger:          logger,
		messageParser: messageParser{
			unmarshaler: unmarshaler,
		},
		messages: make(chan *messageHolder),
		wg:       sync.WaitGroup{},
	}
	handler.wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	handler.cancel = cancel
	go handler.startProcessing(ctx)
	return handler
}

func (h *batchedMessageHandler) startProcessing(ctx context.Context) {
	defer h.wg.Done()
	buffer := make([]*messageHolder, 0, h.maxBatchSize)
	mustSleep := h.nackResendSleep != NoSleep
	logFields := watermill.LogFields{}
	timer := time.NewTimer(h.maxWaitTime)
	batch := 0

	for {
		if len(buffer) == 0 {
			select {

			case firstMessage, ok := <-h.messages:
				if !ok {
					h.logger.Debug("Messages channel is closed", logFields)
				} else {
					buffer = append(buffer, firstMessage)
				}
			case <-ctx.Done():
				return
			}

			timerx.StopTimer(timer)
			timer.Reset(h.maxWaitTime)
		}

		timerExpired := false
		select {
		case message, ok := <-h.messages:
			if !ok {
				h.logger.Debug("Messages channel is closed", logFields)
			}
			buffer = append(buffer, message)
		case <-timer.C:
			if len(buffer) > 0 {
				h.logger.Trace("Timer expired, sending already fetched messages.", logFields)
			}
			timerExpired = true
		case <-ctx.Done():
			h.logger.Debug("Context done, terminating startProcessing", logFields)
			return
		}
		size := len(buffer)
		if (timerExpired && size > 0) || size == int(h.maxBatchSize) {
			timerExpired = false
			newBuffer, err := h.processBatch(buffer)
			batch++
			logFields["batch"] = batch
			if err != nil {
				return
			}
			if newBuffer == nil {
				return
			}
			buffer = newBuffer
			// if there are messages in the buffer, it means there was NACKs, so we wait
			if len(buffer) > 0 && mustSleep {
				timerx.StopTimer(timer)
				timer.Reset(h.nackResendSleep)
				<-timer.C
			}
		}
		timerx.StopTimer(timer)
		timer.Reset(h.maxWaitTime)
	}
}

func (h *batchedMessageHandler) ProcessMessages(
	ctx context.Context,
	kafkaMessages <-chan *sarama.ConsumerMessage,
	sess sarama.ConsumerGroupSession,
	logFields watermill.LogFields,
) error {
	defer func() {
		h.logger.Debug("BatchedMessageHandler.ProcessMessage is closing, stopping messageHandler...", logFields)
		if h.cancel != nil {
			h.cancel()
		} else {
			h.logger.Debug("Cancel is nil, not calling it", logFields)
			return
		}
		h.wg.Wait()
		h.logger.Debug("MessageHandler stopped successfully, returning", logFields)
	}()

	for {
		select {
		case kafkaMsg := <-kafkaMessages:
			if kafkaMsg == nil {
				h.logger.Debug("KafkaMsg is closed, stopping ProcessMessages", logFields)
				return nil
			}
			msg, err := h.messageParser.prepareAndProcessMessage(ctx, kafkaMsg, h.logger, logFields, sess)
			if err != nil {
				return err
			}
			h.messages <- msg
		case <-h.closing:
			h.logger.Debug("Subscriber is closing", logFields)
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}

func (h *batchedMessageHandler) processBatch(
	buffer []*messageHolder,
) ([]*messageHolder, error) {
	waitChannels := make([]<-chan bool, 0, len(buffer))
	for _, msgHolder := range buffer {
		ctx, cancelCtx := context.WithCancel(msgHolder.message.Context())
		msgHolder.message.SetContext(ctx)
		select {
		case h.outputChannel <- msgHolder.message:
			h.logger.Trace("Message sent to consumer", msgHolder.logFields)
			waitChannels = append(waitChannels, waitForMessage(ctx, h.logger, msgHolder, cancelCtx))
		case <-h.closing:
			h.logger.Trace("Closing, message discarded", msgHolder.logFields)
			defer cancelCtx()
			return nil, nil
		case <-ctx.Done():
			h.logger.Trace("Closing, ctx cancelled before message was sent to consumer", msgHolder.logFields)
			defer cancelCtx()
			return nil, nil
		}
	}

	// we wait for all the messages to be ACKed or NACKed
	// and we store for each partition the last message that was ACKed so we
	// can mark the latest complete offset
	lastComittableMessages := make(map[string]*messageHolder, 0)
	nackedPartitions := make(map[string]struct{})
	newBuffer := make([]*messageHolder, 0, h.maxBatchSize)

	defer func() {
		// If a session is provided, we mark the latest committable message for
		// each partition as done. This is required, because if we did not mark anything we might re-process
		// messages unnecessarily. If we marked the latest in the bulk, we could lose NACKed messages.
		//
		// The reason we defer the call to MarkMessage is because we want to mark the messages as complete
		// even when the context is cancelled, since we might have already processed (acked/nacked) some offsets that would not need to be reprocessed by other consumers.
		for _, lastComittable := range lastComittableMessages {
			if lastComittable.sess != nil {
				h.logger.Trace("Marking offset as complete for", lastComittable.logFields)
				lastComittable.sess.MarkMessage(lastComittable.kafkaMessage, "")
			}
		}
	}()

	for idx, waitChannel := range waitChannels {
		msgHolder := buffer[idx]
		h.logger.Trace("Waiting for message to be acked", msgHolder.logFields)
		//lint:ignore S1000 We keep the select to be able to break the select when a nack is received
		select {
		case ack, ok := <-waitChannel:
			h.logger.Debug("Received ACK / NACK response or closed", msgHolder.logFields)
			// it was aborted
			if !ok {
				h.logger.Info("Returning as messages were closed", msgHolder.logFields)
				return nil, nil
			}
			topicAndPartition := fmt.Sprintf("%s-%d", msgHolder.kafkaMessage.Topic, msgHolder.kafkaMessage.Partition)
			_, partitionNacked := nackedPartitions[topicAndPartition]
			if !ack || partitionNacked {
				newBuffer = append(newBuffer, msgHolder.Copy())
				nackedPartitions[topicAndPartition] = struct{}{}
				break
			}
			if !partitionNacked && ack {
				lastComittableMessages[topicAndPartition] = msgHolder
			}
		}
	}

	return newBuffer, nil
}
