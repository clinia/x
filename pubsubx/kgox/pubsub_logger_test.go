package kgox

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/clinia/x/loggerx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestPubSubLoggerHelper(t *testing.T) {
	t.Run("should return the proper log level from kgo", func(t *testing.T) {
		value := []kgo.LogLevel{
			kgo.LogLevelNone,
			kgo.LogLevelDebug,
			kgo.LogLevelInfo,
			kgo.LogLevelWarn,
			kgo.LogLevelError,
			kgo.LogLevel(7), // Unknown Level
		}
		expected := []slog.Level{
			slog.LevelDebug - 4,
			slog.LevelDebug,
			slog.LevelInfo,
			slog.LevelWarn,
			slog.LevelError,
			slog.LevelInfo,
		}
		require.Equal(t, len(value), len(expected))
		for i := range value {
			assert.Equal(t, expected[i], kgoLogLevelToSlogLevel(value[i]))
		}
	})
}

func TestPubSubLoggerLog(t *testing.T) {
	t.Run("should return the proper log level", func(t *testing.T) {
		value := []slog.Level{
			slog.LevelDebug - 4,
			slog.LevelDebug,
			slog.LevelInfo,
			slog.LevelWarn,
			slog.LevelError,
		}
		expected := []kgo.LogLevel{
			kgo.LogLevelNone,
			kgo.LogLevelDebug,
			kgo.LogLevelInfo,
			kgo.LogLevelWarn,
			kgo.LogLevelError,
		}
		require.Equal(t, len(value), len(expected))
		for i := range value {
			buf := new(bytes.Buffer)
			l := &loggerx.Logger{Logger: slog.New(
				slog.NewTextHandler(buf, &slog.HandlerOptions{Level: value[i]}),
			)}
			psl := &pubsubLogger{l: l}
			assert.Equal(t, expected[i].String(), psl.Level().String())
		}
	})

	t.Run("should log the proper log level", func(t *testing.T) {
		expected := []slog.Level{
			slog.LevelDebug - 4,
			slog.LevelDebug,
			slog.LevelInfo,
			slog.LevelWarn,
			slog.LevelError,
		}
		value := []kgo.LogLevel{
			kgo.LogLevelNone,
			kgo.LogLevelDebug,
			kgo.LogLevelInfo,
			kgo.LogLevelWarn,
			kgo.LogLevelError,
		}
		require.Equal(t, len(value), len(expected))
		for i := range value {
			buf := new(bytes.Buffer)
			l := &loggerx.Logger{Logger: slog.New(
				slog.NewTextHandler(buf, &slog.HandlerOptions{Level: expected[i]}),
			)}
			psl := &pubsubLogger{l: l}
			psl.Log(value[i], "test_msg")
			assert.Contains(t, buf.String(), "test_msg")
			assert.Contains(t, buf.String(), expected[i].String())
		}
	})

	// TODO: [ENG-1953] Either fix or remove these tests
	// t.Run("should add the log fields when they are passed in", func(t *testing.T) {
	// 	l := getLogger()
	// 	h := test.NewLocal(l.Entry.Logger)
	// 	psl := &pubsubLogger{l: l}
	// 	psl.Log(kgo.LogLevelInfo, "test_msg", "test_key", "value", "test_key_with_interface", struct{}{})
	// 	le := h.LastEntry()
	// 	assert.Equal(t, "value", le.Data["pubsub.test_key"])
	// 	assert.Equal(t, struct{}{}, le.Data["pubsub.test_key_with_interface"])
	// 	assert.Equal(t, "test_msg", le.Message)
	// 	assert.Equal(t, logrus.InfoLevel, le.Level)
	// })

	// t.Run("should failed to add a field if the keyvals format is wrong (keyvals length)", func(t *testing.T) {
	// 	l := getLogger()
	// 	h := test.NewLocal(l.Entry.Logger)
	// 	psl := &pubsubLogger{l: l}
	// 	psl.Log(kgo.LogLevelInfo, "test_msg", "test_key", "value", "test_key_with_interface")
	// 	require.Equal(t, 2, len(h.AllEntries()))
	// 	assert.Contains(t, h.AllEntries()[0].Message, "failed to evaluate keyvals as the vals count is odd (3)")
	// 	assert.Equal(t, logrus.ErrorLevel, h.AllEntries()[0].Level)
	// 	assert.Equal(t, "test_msg", h.AllEntries()[1].Message)
	// 	assert.Equal(t, logrus.InfoLevel, h.AllEntries()[1].Level)
	// })

	// t.Run("should failed to add a field if the keyvals format is wrong (key type wrong)", func(t *testing.T) {
	// 	l := getLogger()
	// 	h := test.NewLocal(l.Entry.Logger)
	// 	psl := &pubsubLogger{l: l}
	// 	psl.Log(kgo.LogLevelInfo, "test_msg", struct{}{}, "value")
	// 	require.Equal(t, 2, len(h.AllEntries()))
	// 	assert.Contains(t, h.AllEntries()[0].Message, "key is not a string type for kgo keyval pair, unable to add field key")
	// 	assert.Equal(t, logrus.ErrorLevel, h.AllEntries()[0].Level)
	// 	assert.Equal(t, "test_msg", h.AllEntries()[1].Message)
	// 	assert.Equal(t, logrus.InfoLevel, h.AllEntries()[1].Level)
	// })
}
