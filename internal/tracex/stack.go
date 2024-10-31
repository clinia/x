package internaltracex

import (
	"fmt"
	"runtime"
)

// GetStackTrace returns the stack trace of the caller.
// We can specify skipLevels to skip the number of stack frames.
// i.e. GetStackTrace(2) will skip the first 2 calls (including GetStackTrace itself), so this will return the stack trace of the caller of the function that calls GetStackTrace.
func GetStackTrace(skipLevels int) string {
	pc := make([]uintptr, 10)
	n := runtime.Callers(skipLevels, pc)
	frames := runtime.CallersFrames(pc[:n])

	var stackBuf string
	for {
		frame, more := frames.Next()
		stackBuf += fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more || len(stackBuf) > 1024 {
			break
		}

	}

	return stackBuf
}
