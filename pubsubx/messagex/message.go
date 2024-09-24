package messagex

import "github.com/segmentio/ksuid"

const IDHeaderKey = "_clinia_message_id"

// Message intentionally has no json marshalling fields as we want to pass by our own kgox.DefaultMarshaler
type Message struct {
	ID       string
	Metadata MessageMetadata
	Payload  []byte
}

type MessageMetadata map[string]string

// NewMessage creates a new Message with the given payload and options.
//
// Parameters:
//   - payload: The payload of the message as a byte slice.
//   - opts: Optional parameters to customize the creation of the message. By default, a new UUID is generated for the message.
//
// Returns:
//   - *Message: A pointer to the created Message.
func NewMessage(payload []byte, opts ...newMessageOption) *Message {
	o := newMessageOptions{}
	for _, opt := range opts {
		opt(&o)
	}

	if o.id == "" {
		o.id = ksuid.New().String()
	}

	if o.m == nil {
		o.m = make(MessageMetadata)
	}

	return &Message{
		ID:       o.id,
		Metadata: o.m,
		Payload:  payload,
	}
}

type newMessageOptions struct {
	id string
	m  MessageMetadata
}

type newMessageOption func(*newMessageOptions)

// WithID sets the ID of the message.
// A ksuid will be generated if no ID is provided.
func WithID(id string) newMessageOption {
	return func(o *newMessageOptions) {
		o.id = id
	}
}

// WithMetadata sets the metadata of the message.
func WithMetadata(m MessageMetadata) newMessageOption {
	return func(o *newMessageOptions) {
		o.m = m
	}
}
