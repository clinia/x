package kgox

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/clinia/x/loggerx"
	"github.com/twmb/franz-go/pkg/kgo"
)

type pubsubLogger struct {
	l *loggerx.Logger
}

func kgoLogLevelToSlogLevel(l kgo.LogLevel) slog.Level {
	switch l {
	case kgo.LogLevelNone:
		return slog.LevelDebug - 4 // Below debug, similar to trace
	case kgo.LogLevelDebug:
		return slog.LevelDebug
	case kgo.LogLevelInfo:
		return slog.LevelInfo
	case kgo.LogLevelWarn:
		return slog.LevelWarn
	case kgo.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (cl *pubsubLogger) Level() kgo.LogLevel {
	ctx := context.Background()
	// Check from most verbose to least verbose
	switch {
	case cl.l.Logger.Enabled(ctx, slog.LevelDebug-4):
		return kgo.LogLevelNone // Trace-like level
	case cl.l.Logger.Enabled(ctx, slog.LevelDebug):
		return kgo.LogLevelDebug
	case cl.l.Logger.Enabled(ctx, slog.LevelInfo):
		return kgo.LogLevelInfo
	case cl.l.Logger.Enabled(ctx, slog.LevelWarn):
		return kgo.LogLevelWarn
	case cl.l.Logger.Enabled(ctx, slog.LevelError):
		return kgo.LogLevelError
	default:
		return kgo.LogLevelNone
	}
}

func (cl *pubsubLogger) Log(level kgo.LogLevel, msg string, keyvals ...any) {
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			cl.l.Error(ctx, fmt.Sprintf("pubsub logger panic : %v", r))
		}
	}()
	fields := []slog.Attr{}
	// By design, the keyvals slice should always be a pair length as each key comes
	// with a subsequent entry that is a val
	if len(keyvals)%2 != 0 {
		cl.l.Error(ctx, fmt.Sprintf("failed to evaluate keyvals as the vals count is odd (%d)", len(keyvals)))
	} else if len(keyvals) > 0 {
		for i := 0; i < len(keyvals)-1; i += 2 {
			if key, ok := keyvals[i].(string); ok {
				fk := fmt.Sprintf("pubsub.%s", key)
				slog.Any(fk, keyvals[i+1])
			} else {
				cl.l.Error(ctx, fmt.Sprintf("key is not a string type for kgo keyval pair, unable to add field key : %v", keyvals[i]))
			}
		}
	}
	lev := kgoLogLevelToSlogLevel(level)
	cl.l.Logger.LogAttrs(ctx, lev, msg, fields...)
}
