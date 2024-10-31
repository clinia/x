package tracexx

import (
	internaltracex "github.com/clinia/x/internal/tracex"
	"github.com/clinia/x/logrusx"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// RecoverWithStackTrace recovers from a panic and logs the message with a stack trace.
// It should only be used as a defer statement at the beginning of a function.
// i.e. defer tracex.RecoverWithStackTrace(l, "panic while handling messages")
func RecoverWithStackTrace(l *logrusx.Logger, msg string) {
	// We don't want the recoverer itself to panic - that would be a shame.
	defer recover()

	if r := recover(); r != nil {
		if l == nil {
			return
		}
		// We want to omit the getStackTrace but preserve RecoverWithStackTrace
		stackTrace := internaltracex.GetStackTrace(2)
		l = l.WithFields(logrusx.NewLogFields(semconv.ExceptionStacktrace(stackTrace)))
		switch v := r.(type) {
		case string:
			l = l.WithField(string(semconv.ExceptionMessageKey), v)
		case error:
			l = l.WithField(string(semconv.ExceptionMessageKey), v.Error())
		default:
			l = l.WithField(string(semconv.ExceptionMessageKey), "unknown panic")
		}

		l.Errorln(msg)
	}
}

// GetStackTrace returns the stack trace of the caller.
func GetStackTrace() string {
	return internaltracex.GetStackTrace(3)
}
