package kgox

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestPubSubLogger(t *testing.T) {
	t.Run("should return the proper log level", func(t *testing.T) {
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
}
