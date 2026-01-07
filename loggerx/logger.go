package loggerx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

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
	return &Logger{Logger: l.With(slogx.ErrorAttr(err))}
}

func (l *Logger) WithErrors(errs ...error) *Logger {
	return &Logger{Logger: l.With(slogx.ErrorsAttr(errs...))}
}

func (l *Logger) Panic(ctx context.Context, msg string, kvs ...attribute.KeyValue) *Logger {
	l.Error(ctx, msg, kvs...)
	panic(msg)
}

func (l *Logger) WithSpanStartOptions(opts ...trace.SpanStartOption) *Logger {
	sc := trace.NewSpanStartConfig(opts...)
	return l.WithFields(sc.Attributes()...)
}

func (l *Logger) WithStackTrace() *Logger {
	stackTrace := internaltracex.GetStackTrace(3)
	return l.WithFields(semconv.ExceptionStacktrace(stackTrace))
}

func (l *Logger) Error(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	l.log(ctx, slog.LevelError, msg, kvs...)
}

func (l *Logger) Warn(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	l.log(ctx, slog.LevelWarn, msg, kvs...)
}

func (l *Logger) Info(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	l.log(ctx, slog.LevelInfo, msg, kvs...)
}

func (l *Logger) Debug(ctx context.Context, msg string, kvs ...attribute.KeyValue) {
	l.log(ctx, slog.LevelDebug, msg, kvs...)
}

func (l *Logger) WithFields(kvs ...attribute.KeyValue) *Logger {
	lfs := slogx.NewLogFields(kvs...)
	// This is a workaround until we get a nice slog.WithAttrs method - See https://github.com/golang/go/issues/66937#issuecomment-2730350514
	return &Logger{l.With("", slog.GroupValue(lfs...))}
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, kvs ...attribute.KeyValue) {
	if !l.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	// Skip 3: runtime.Callers + l.log + l.Info/Error
	runtime.Callers(3, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(slogx.NewLogFields(kvs...)...)

	if err := l.Logger.Handler().Handle(ctx, r); err != nil {
		// In case logging itself fails, write the error to stderr
		fmt.Fprintf(os.Stderr, "logging error: %v\n", err)
	}
}
