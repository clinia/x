package messagex

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type MessageReferenceTracer interface {
	Message() *Message
	ReferenceMsgSpanName() string

	// StartMessageProcessingSpan starts a new span with a bidirectional link with
	// the original span associated with the creation of the provided Message.
	//
	// For the created span, a link will be added pointing to the original message's span.
	// For the original message's span, a subspan will be created with `${msgReferenceSpanName}`
	// which links back to the current span. If `msgReferenceSpanName` is empty, the default
	// name "async processing reference" will be used.
	//
	// This function should be called for any asynchronous processing of the message to enable
	// advanced tracing and provide a clear view of the message's lifecycle and its associated
	// operations.
	StartMessageProcessingSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
	// AttachMessageProcessingSpan creates a bidirectional link between the current span
	// and the original span associated with the creation of the provided Message.
	// This is useful when the original message's span is already created and you want to
	// attach the current span to it.
	//
	// This function should be called for any asynchronous processing of the message to enable
	// advanced tracing and provide a clear view of the message's lifecycle and its associated
	// operations.
	AttachMessageProcessingSpan(ctx context.Context, span trace.Span)

	recordReferenceSpan(originalMsgContext context.Context, span trace.Span)
}
