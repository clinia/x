package loggerx

import (
	"context"
	"log/slog"

	internaltracex "github.com/clinia/x/internal/tracex"
	"github.com/clinia/x/slogx"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	*slog.Logger
}

func NewDefaultLogger() *Logger {
	return &Logger{Logger: slog.Default()}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.Logger.With(slogx.ErrorAttr(err))}
}

func (l *Logger) Panic(ctx context.Context, msg string, kvs ...attribute.KeyValue) *Logger {
	l.Error(ctx, msg, kvs...)
	panic(msg)
}

func (l *Logger) WithSpanStartOptions(opts ...trace.SpanStartOption) *Logger {
	sc := trace.NewSpanStartConfig(opts...)
	ll := l.WithFields(sc.Attributes()...)
	return ll
}

func (l *Logger) WithStackTrace() *Logger {
	stackTrace := internaltracex.GetStackTrace(3)
	return l.WithFields(semconv.ExceptionStacktrace(stackTrace))
}

func (l *Logger) Error(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	l.Logger.LogAttrs(ctx, slog.LevelError, msg, slogx.NewLogFields(kvs...)...)
}

func (l *Logger) Warn(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	l.Logger.LogAttrs(ctx, slog.LevelWarn, msg, slogx.NewLogFields(kvs...)...)
}

func (l *Logger) Info(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	l.Logger.LogAttrs(ctx, slog.LevelInfo, msg, slogx.NewLogFields(kvs...)...)
}

func (l *Logger) Debug(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	l.Logger.LogAttrs(ctx, slog.LevelDebug, msg, slogx.NewLogFields(kvs...)...)
}

func (l *Logger) WithFields(kvs ...attribute.KeyValue) *Logger {
	lfs := slogx.NewLogFields(kvs...)
	// This is a workaround until we get a nice slog.WithAttrs method - See https://github.com/golang/go/issues/66937#issuecomment-2730350514
	return &Logger{l.Logger.With("", slog.GroupValue(lfs...))}
}
