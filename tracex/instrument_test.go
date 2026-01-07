package tracex

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/clinia/x/loggerx"
	loggerxtest "github.com/clinia/x/loggerx/test"
	"github.com/clinia/x/otelx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestComponentName(t *testing.T) {
	t.Run("should return component name", func(t *testing.T) {
		assert.Equal(t, "testComponent.testStructName", ComponentName("testComponent", "testStructName"))
	})
}

func TestInstrumentNext(t *testing.T) {
	l, buf := loggerxtest.NewTestLoggerWithJSONBuffer(t)
	lp := func() *loggerx.Logger {
		return l
	}
	ot := otelx.NewNoopTracer("test")
	tp := func(ctx context.Context) *otelx.Tracer {
		return ot
	}
	t.Run("should return instrumentation outputs", func(t *testing.T) {
		ctx, span, logger := InstrumentNext(context.Background(), lp, tp, "testComponent.testStruct", "testInstrument", trace.WithAttributes(attribute.Bool("test", true)))
		assert.Equal(t, span, trace.SpanFromContext(ctx))
		assert.NotSame(t, l, logger)

		// Log a message to verify fields are present
		logger.Info(ctx, "test message")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "test message", logEntry["msg"])
		assert.Equal(t, true, logEntry["test"])
		assert.Equal(t, "testComponent.testStruct.testInstrument", logEntry["component"])
	})
}
