package messagex

import (
	"context"
	"fmt"

	"github.com/clinia/x/featureflagx"
	"github.com/segmentio/ksuid"
	"go.opentelemetry.io/otel/trace"
)

const (
	IDHeaderKey           = "_clinia_message_id"
	RetryCountHeaderKey   = "_clinia_retry_count"
	FeatureFlagsHeaderKey = "_clinia_feature_flags"
)

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

	if _, ok := o.m[RetryCountHeaderKey]; !ok {
		o.m[RetryCountHeaderKey] = "0"
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

func WithFeatureFlags(ffs *featureflagx.FeatureFlags) newMessageOption {
	return func(o *newMessageOptions) {
		if ffs == nil {
			return
		}
		if o.m == nil {
			o.m = make(MessageMetadata)
		}
		// Serialize the feature flags to JSON
		s, err := ffs.MarshalJSON()
		if err == nil {
			o.m["_clinia_feature_flags"] = string(s)
		}
		// } else {
		// 	panic(fmt.Errorf("failed to marshal feature flags: %w", err))
		// }
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

func (m *Message) ExtractFeatureFlags() (*featureflagx.FeatureFlags, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	if featureFlags, ok := m.Metadata[FeatureFlagsHeaderKey]; !ok {
		return nil, false
	} else {
		// Unmarshal the feature flags from JSON
		ff, err := featureflagx.New(map[string]bool{}, []featureflagx.FeatureFlag{})
		if err != nil {
			// panic(fmt.Errorf("failed to create feature flags: %w", err))
			return nil, false
		}
		err = ff.UnmarshalJSON([]byte(featureFlags))
		if err != nil {
			// panic(fmt.Errorf("failed to unmarshal feature flags: %w", err))
			return nil, false
		}

		if len(ff.GetFlags()) == 0 {
			return nil, false
		}
		// Create a new context with the feature flags
		return ff, true
	}
}
