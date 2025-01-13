package kgox

import (
	"github.com/clinia/x/logrusx"
	"github.com/sirupsen/logrus"
	"github.com/twmb/franz-go/pkg/kgo"
)

type pubsubLogger struct {
	l *logrusx.Logger
}

func kgoLogLevelToLogrusLogLevel(l kgo.LogLevel) logrus.Level {
	switch l {
	case kgo.LogLevelDebug:
		return logrus.DebugLevel
	case kgo.LogLevelInfo:
		return logrus.InfoLevel
	case kgo.LogLevelWarn:
		return logrus.WarnLevel
	case kgo.LogLevelError:
		return logrus.ErrorLevel
	case kgo.LogLevelNone:
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

func logrusLogLevelToKgoLogLevel(l logrus.Level) kgo.LogLevel {
	switch l {
	case logrus.TraceLevel, logrus.DebugLevel:
		return kgo.LogLevelDebug
	case logrus.InfoLevel:
		return kgo.LogLevelInfo
	case logrus.WarnLevel:
		return kgo.LogLevelWarn
	case logrus.ErrorLevel:
		return kgo.LogLevelError
	case logrus.FatalLevel, logrus.PanicLevel:
		return kgo.LogLevelNone
	default:
		return kgo.LogLevelInfo
	}
}

func (cl *pubsubLogger) Level() kgo.LogLevel {
	return logrusLogLevelToKgoLogLevel(cl.l.Logrus().GetLevel())
}

func (cl *pubsubLogger) Log(level kgo.LogLevel, msg string, keyvals ...any) {
	fields := make(logrus.Fields)
	// By design, the keyvals slice should always be a pair length as each key comes
	// with a subsequent entry that is a val
	if len(keyvals)%2 != 0 {
		cl.l.Errorf("failed to evaluate keyvals : %v", keyvals)
	} else {
		for i := range keyvals {
			if i%2 == 0 {
				fields["pubsub."+keyvals[i].(string)] = keyvals[i+1]
			}
		}
	}
	cl.l.WithFields(fields).Logf(kgoLogLevelToLogrusLogLevel(level), "%s", msg)
}
