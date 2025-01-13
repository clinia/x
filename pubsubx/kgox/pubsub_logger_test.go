package kgox

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestPubSubLogger(t *testing.T) {
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
			logrus.PanicLevel,
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
			kgo.LogLevelNone,
			kgo.LogLevelNone,
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
