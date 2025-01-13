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
	}
	return logrus.InfoLevel
}

func (cl *pubsubLogger) Level() kgo.LogLevel {
	switch cl.l.Entry.Level {
	case logrus.TraceLevel, logrus.DebugLevel:
		return kgo.LogLevelDebug
	case logrus.InfoLevel:
		return kgo.LogLevelInfo
	case logrus.WarnLevel:
		return kgo.LogLevelWarn
	case logrus.ErrorLevel:
		return kgo.LogLevelError
	case logrus.FatalLevel:
		return kgo.LogLevelNone
	}
	return kgo.LogLevelInfo
}

func (cl *pubsubLogger) Log(level kgo.LogLevel, msg string, keyvals ...any) {
	fields := make(logrus.Fields)
	if len(keyvals)%2 != 0 {
		cl.l.Errorf("failed to evaluate keyvals : %v", keyvals)
	} else {
		for i := range keyvals {
			if i%2 == 0 {
				fields[keyvals[i].(string)] = keyvals[i+1]
			}
		}
	}
	cl.l.WithFields(fields).Logf(kgoLogLevelToLogrusLogLevel(level), msg)
}
