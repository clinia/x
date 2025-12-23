package loggerx

import (
	"context"
	"log/slog"

	"github.com/clinia/x/slogx"
	"github.com/clinia/x/tracex"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Logger struct {
	*slog.Logger
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.Logger.With(slogx.ErrorAttr(err))}
}

func (l *Logger) Panic(ctx context.Context, msg string, kvs ...attribute.KeyValue) *Logger {
	l.Error(ctx, msg, kvs...)
	panic(msg)
}

func (l *Logger) WithStackTrace() *Logger {
	stackTrace := tracex.GetStackTrace()
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
