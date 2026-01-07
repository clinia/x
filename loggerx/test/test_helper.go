package loggerxtest

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/clinia/x/loggerx"
)

func NewTestLogger(t testing.TB) *loggerx.Logger {
	t.Helper()
	return &loggerx.Logger{Logger: slog.New(slog.DiscardHandler)}
}

func NewTestLoggerWithJSONBuffer(t testing.TB) (*loggerx.Logger, *bytes.Buffer) {
	t.Helper()
	buf := new(bytes.Buffer)
	l := slog.New(slog.NewJSONHandler(buf, nil))
	return &loggerx.Logger{Logger: l}, buf
}

func NewTestLoggerWithTextBuffer(t testing.TB) (*loggerx.Logger, *bytes.Buffer) {
	t.Helper()
	buf := new(bytes.Buffer)
	l := slog.New(slog.NewTextHandler(buf, nil))
	return &loggerx.Logger{Logger: l}, buf
}
