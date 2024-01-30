package pubsubx

type ConsumerModel string

var (

	// Default is a model when only one message is sent to the customer and customer needs to ACK the message
	// to receive the next.
	ConsumerModelDefault ConsumerModel = ""
	// Batch works by sending multiple messages in a batch
	// You can ack all of them at once, or one by one.
	ConsumerModelBatch ConsumerModel = "batch"
)
