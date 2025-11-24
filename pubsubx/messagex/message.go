package messagex

import (
	"context"
	"fmt"

	"github.com/segmentio/ksuid"
	"go.opentelemetry.io/otel/trace"
)

const (
	IDHeaderKey         = "_clinia_message_id"
	RetryCountHeaderKey = "_clinia_retry_count"
)

// Message intentionally has no json marshalling fields as we want to pass by our own kgox.DefaultMarshaler
type Message struct {
	ID string
	// When specified, Key is used as the Kafka message key.
	Key      []byte
	Metadata MessageMetadata
	Payload  []byte
	// The offset of the message. This is readonly and only set when consuming.
	Offset int64
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
func NewMessage(payload []byte, opts ...NewMessageOption) *Message {
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

	if _, ok := o.m[RetryCountHeaderKey]; !ok {
		o.m[RetryCountHeaderKey] = "0"
	}

	return &Message{
		ID:       o.id,
		Key:      o.key,
		Metadata: o.m,
		Payload:  payload,
	}
}

type newMessageOptions struct {
	key []byte
	id  string
	m   MessageMetadata
}

type NewMessageOption func(*newMessageOptions)

// WithID sets the ID of the message.
// A ksuid will be generated if no ID is provided.
func WithID(id string) NewMessageOption {
	return func(o *newMessageOptions) {
		o.id = id
	}
}

// WithStringKey sets the key of the message using a string.
func WithStringKey(key string) NewMessageOption {
	return WithKey([]byte(key))
}

// WithKey sets the key of the message.
func WithKey(key []byte) NewMessageOption {
	return func(o *newMessageOptions) {
		o.key = key
	}
}

// WithMetadata sets the metadata of the message.
func WithMetadata(m MessageMetadata) NewMessageOption {
	return func(o *newMessageOptions) {
		o.m = m
	}
}

func (m *Message) WithSpan(ctx context.Context, tracer trace.Tracer, spanPrefix string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// Create a new span for the message
	msgctx, span := tracer.Start(ctx, fmt.Sprintf("%s.message", spanPrefix), opts...)

	// Inject TraceContext to the message metadata
	prop := NewTraceContextPropagator()
	prop.Inject(msgctx, m)

	return msgctx, span
}

func (m *Message) Copy() *Message {
	newMessage := Message{
		ID:       m.ID,
		Metadata: MessageMetadata{},
		Payload:  make([]byte, len(m.Payload)),
	}

	copy(newMessage.Payload, m.Payload)

	for key, value := range m.Metadata {
		newMessage.Metadata[key] = value
	}

	return &newMessage
}

func (m *Message) ExtractTraceContext(ctx context.Context) context.Context {
	prop := NewTraceContextPropagator()
	return prop.Extract(ctx, m.Metadata)
}

func (m *Message) InjectTraceContext(ctx context.Context) {
	prop := NewTraceContextPropagator()
	prop.Inject(ctx, m)
}
