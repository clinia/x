package internaltracex

import (
	"fmt"
	"runtime"
)

func GetStackTrace(skipLevels int) string {
	// Skip the first two callers (CallerStackTrace and ExampleFunction)
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
