package kgox

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
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
		expected := []logrus.Level{
			logrus.TraceLevel,
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.InfoLevel,
		}
		require.Equal(t, len(value), len(expected))
		for i := range value {
			assert.Equal(t, expected[i], kgoLogLevelToLogrusLogLevel(value[i]))
		}
	})

	t.Run("should return the proper log level from logrus", func(t *testing.T) {
		value := []logrus.Level{
			logrus.FatalLevel,
			logrus.PanicLevel,
			logrus.TraceLevel,
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.Level(10),
		}
		expected := []kgo.LogLevel{
			kgo.LogLevelError,
			kgo.LogLevelError,
			kgo.LogLevelDebug,
			kgo.LogLevelDebug,
			kgo.LogLevelInfo,
			kgo.LogLevelWarn,
			kgo.LogLevelError,
			kgo.LogLevelInfo,
		}
		require.Equal(t, len(value), len(expected))
		for i := range value {
			assert.Equal(t, expected[i], logrusLogLevelToKgoLogLevel(value[i]))
		}
	})
}

func TestPubSubLoggerLog(t *testing.T) {
	t.Run("should return the proper log level", func(t *testing.T) {
		value := []logrus.Level{
			logrus.PanicLevel,
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
		}
		expected := []kgo.LogLevel{
			kgo.LogLevelError,
			kgo.LogLevelDebug,
			kgo.LogLevelInfo,
			kgo.LogLevelWarn,
			kgo.LogLevelError,
		}
		require.Equal(t, len(value), len(expected))
		for i := range value {
			l := getLogger()
			l.Entry.Level = value[i]
			l.Entry.Logger.Level = value[i]
			psl := &pubsubLogger{l: l}
			assert.Equal(t, expected[i].String(), psl.Level().String())
		}
	})

	t.Run("should log the proper log level", func(t *testing.T) {
		expected := []logrus.Level{
			logrus.ErrorLevel,
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.TraceLevel,
		}
		value := []kgo.LogLevel{
			kgo.LogLevelError,
			kgo.LogLevelDebug,
			kgo.LogLevelInfo,
			kgo.LogLevelWarn,
			kgo.LogLevelNone,
		}
		require.Equal(t, len(value), len(expected))
		for i := range value {
			l := getLogger()
			l.Entry.Level = expected[i]
			l.Entry.Logger.Level = expected[i]
			h := test.NewLocal(l.Entry.Logger)
			psl := &pubsubLogger{l: l}
			psl.Log(value[i], "test_msg")
			assert.Equal(t, 1, len(h.AllEntries()), "log level %s didn't log the right amount of time", value[i].String())
			assert.Equal(t, expected[i].String(), h.AllEntries()[0].Level.String())
		}
	})

	t.Run("should add the log fields when they are passed in", func(t *testing.T) {
		l := getLogger()
		h := test.NewLocal(l.Entry.Logger)
		psl := &pubsubLogger{l: l}
		psl.Log(kgo.LogLevelInfo, "test_msg", "test_key", "value", "test_key_with_interface", struct{}{})
		le := h.LastEntry()
		assert.Equal(t, "value", le.Data["pubsub.test_key"])
		assert.Equal(t, struct{}{}, le.Data["pubsub.test_key_with_interface"])
		assert.Equal(t, "test_msg", le.Message)
		assert.Equal(t, logrus.InfoLevel, le.Level)
	})

	t.Run("should failed to add a field if the keyvals format is wrong (keyvals length)", func(t *testing.T) {
		l := getLogger()
		h := test.NewLocal(l.Entry.Logger)
		psl := &pubsubLogger{l: l}
		psl.Log(kgo.LogLevelInfo, "test_msg", "test_key", "value", "test_key_with_interface")
		require.Equal(t, 2, len(h.AllEntries()))
		assert.Contains(t, h.AllEntries()[0].Message, "failed to evaluate keyvals as the vals count is odd (3)")
		assert.Equal(t, logrus.ErrorLevel, h.AllEntries()[0].Level)
		assert.Equal(t, "test_msg", h.AllEntries()[1].Message)
		assert.Equal(t, logrus.InfoLevel, h.AllEntries()[1].Level)
	})

	t.Run("should failed to add a field if the keyvals format is wrong (key type wrong)", func(t *testing.T) {
		l := getLogger()
		h := test.NewLocal(l.Entry.Logger)
		psl := &pubsubLogger{l: l}
		psl.Log(kgo.LogLevelInfo, "test_msg", struct{}{}, "value")
		require.Equal(t, 2, len(h.AllEntries()))
		assert.Contains(t, h.AllEntries()[0].Message, "key is not a string type for kgo keyval pair, unable to add field key")
		assert.Equal(t, logrus.ErrorLevel, h.AllEntries()[0].Level)
		assert.Equal(t, "test_msg", h.AllEntries()[1].Message)
		assert.Equal(t, logrus.InfoLevel, h.AllEntries()[1].Level)
	})
}
