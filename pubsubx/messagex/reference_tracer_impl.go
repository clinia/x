package messagex

import (
	"context"

	"github.com/clinia/x/errorx"
	"go.opentelemetry.io/otel/trace"
)

type messageReferenceTracer struct {
	msg                  *Message
	referenceMsgSpanName string
	tr                   trace.Tracer
}

var _ MessageReferenceTracer = (*messageReferenceTracer)(nil)

func NewMessageReferenceTracer(msg *Message, referenceMsgSpanName string, tr trace.Tracer) (MessageReferenceTracer, error) {
	if msg == nil {
		return nil, errorx.InternalErrorf("message cannot be nil")
	}

	if tr == nil {
		return nil, errorx.InternalErrorf("tracer cannot be nil")
	}

	return messageReferenceTracer{
		msg:                  msg,
		referenceMsgSpanName: referenceMsgSpanName,
		tr:                   tr,
	}, nil
}

// AttachMessageProcessingSpan implements MessageReferenceTracer.
func (m messageReferenceTracer) AttachMessageProcessingSpan(ctx context.Context, span trace.Span) {
	originalMsgCtx := context.Background()
	originalMsgCtx = m.Message().ExtractTraceContext(originalMsgCtx)

	// Check if context is empty
	// If it is, we don't want to add a link to the span because it will be a new span
	if originalMsgCtx == context.Background() {
		return
	}

	span.AddLink(trace.LinkFromContext(originalMsgCtx))
	m.recordReferenceSpan(originalMsgCtx, span)
}

// StartMessageProcessingSpan implements MessageReferenceTracer.
func (m messageReferenceTracer) StartMessageProcessingSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	originalMsgCtx := context.Background()
	originalMsgCtx = m.Message().ExtractTraceContext(originalMsgCtx)
	opts = append(opts, trace.WithLinks(
		trace.LinkFromContext(originalMsgCtx),
	))

	ctx, currentSpan := m.tr.Start(ctx, name, opts...)

	m.recordReferenceSpan(originalMsgCtx, currentSpan)

	return ctx, currentSpan
}

// recordReferenceSpan implements MessageReferenceTracer.
func (m messageReferenceTracer) recordReferenceSpan(originalMsgCtx context.Context, currentSpan trace.Span) {
	refSpanName := m.ReferenceMsgSpanName()
	if refSpanName == "" {
		refSpanName = "async processing reference"
	}

	_, refSpan := m.tr.Start(originalMsgCtx, refSpanName, trace.WithLinks(trace.Link{
		SpanContext: currentSpan.SpanContext(),
	}))
	refSpan.End()
}

// Message implements MessageReferenceTracerInfo.
func (m messageReferenceTracer) Message() *Message {
	return m.msg
}

// ReferenceMsgSpanName implements MessageReferenceTracerInfo.
func (m messageReferenceTracer) ReferenceMsgSpanName() string {
	return m.referenceMsgSpanName
}
