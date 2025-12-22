package slogx

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
)

func Error(ctx context.Context, msg string, err error, kvs ...attribute.KeyValue) {
	slog.LogAttrs(ctx, slog.LevelError, msg, append(NewLogFields(kvs...), ErrorAttr(err))...)
}

func Warn(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	slog.LogAttrs(ctx, slog.LevelWarn, msg, NewLogFields(kvs...)...)
}

func Info(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	slog.LogAttrs(ctx, slog.LevelInfo, msg, NewLogFields(kvs...)...)
}

func Debug(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	slog.LogAttrs(ctx, slog.LevelDebug, msg, NewLogFields(kvs...)...)
}
