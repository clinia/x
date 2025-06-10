package tracex

import (
	"context"
	"testing"

	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/otelx"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestComponentName(t *testing.T) {
	t.Run("should return component name", func(t *testing.T) {
		assert.Equal(t, "testComponent.testStructName", ComponentName("testComponent", "testStructName"))
	})
}

func TestInstrument(t *testing.T) {
	l := logrusx.New("test", "1.0.0").WithField("testBefore", "value")
	lp := func(ctx context.Context) *logrusx.Logger {
		return l
	}
	ot := otelx.NewNoopTracer("test")
	tp := func(ctx context.Context) *otelx.Tracer {
		return ot
	}
	t.Run("should return instrumentation outputs", func(t *testing.T) {
		ctx, span, logger := Instrument(context.Background(), lp, tp, "testComponent.testStruct", "testInstrument", trace.WithAttributes(attribute.Bool("test", true)))
		assert.Equal(t, span, trace.SpanFromContext(ctx))
		assert.Equal(t, logger.Entry.Data["testBefore"], "value")
		assert.Equal(t, logger.Entry.Data["test"], true)
		assert.NotSame(t, l, logger)
	})
}
