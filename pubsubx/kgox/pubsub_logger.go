package kgox

import (
	"fmt"

	"github.com/clinia/x/logrusx"
	"github.com/sirupsen/logrus"
	"github.com/twmb/franz-go/pkg/kgo"
)

type pubsubLogger struct {
	l *logrusx.Logger
}

func kgoLogLevelToLogrusLogLevel(l kgo.LogLevel) logrus.Level {
	switch l {
	case kgo.LogLevelNone:
		return logrus.TraceLevel
	case kgo.LogLevelDebug:
		return logrus.DebugLevel
	case kgo.LogLevelInfo:
		return logrus.InfoLevel
	case kgo.LogLevelWarn:
		return logrus.WarnLevel
	case kgo.LogLevelError:
		return logrus.ErrorLevel
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
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return kgo.LogLevelError
	default:
		return kgo.LogLevelInfo
	}
}

func (cl *pubsubLogger) Level() kgo.LogLevel {
	return logrusLogLevelToKgoLogLevel(cl.l.Logrus().GetLevel())
}

func (cl *pubsubLogger) Log(level kgo.LogLevel, msg string, keyvals ...any) {
	defer func() {
		if r := recover(); r != nil {
			cl.l.Errorf("pubsub logger panic : %v", r)
		}
	}()
	fields := make(logrus.Fields, len(keyvals)/2)
	// By design, the keyvals slice should always be a pair length as each key comes
	// with a subsequent entry that is a val
	if len(keyvals)%2 != 0 {
		cl.l.Errorf("failed to evaluate keyvals as the vals count is odd (%d)", len(keyvals))
	} else if len(keyvals) > 0 {
		for i := 0; i < len(keyvals)-1; i += 2 {
			if key, ok := keyvals[i].(string); ok {
				fk := fmt.Sprintf("pubsub.%s", key)
				fields[fk] = keyvals[i+1]
			} else {
				cl.l.Errorf("key is not a string type for kgo keyval pair, unable to add field key : %v", keyvals[i])
			}
		}
	}
	lev := kgoLogLevelToLogrusLogLevel(level)
	cl.l.WithFields(fields).Logf(lev, "%s", msg)
}
