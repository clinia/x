package errorx

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
)

const stackTraceDepth = 32

type Frame struct {
	File     string
	Line     int
	Function string
}

// String implements the fmt.Stringer interface.
func (f Frame) String() string {
	s := &strings.Builder{}
	f.writeFrame(s)
	return s.String()
}

// Format implements the fmt.Formatter interface.
//
// The verbs:
//
//	%s	function, file and line number in a single line
//	%f	filename
//	%d	line number
//	%n	function name, the plus flag adds a package name
//	%v	same as %s, the plus or hash flags print struct details
//	%q	a double-quoted Go string with same contents as %s
func (f Frame) Format(s fmt.State, verb rune) {
	type _Frame Frame
	switch verb {
	case 's':
		f.writeFrame(s)
	case 'f':
		io.WriteString(s, f.File)
	case 'd':
		io.WriteString(s, strconv.Itoa(f.Line))
	case 'n':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.Function)
		default:
			io.WriteString(s, shortname(f.Function))
		}
	case 'v':
		switch {
		case s.Flag('+') || s.Flag('#'):
			format(s, verb, _Frame(f))
		default:
			f.Format(s, 's')
		}
	case 'q':
		io.WriteString(s, strconv.Quote(f.String()))
	default:
		format(s, verb, _Frame(f))
	}
}

func (f Frame) writeFrame(w io.Writer) {
	io.WriteString(w, "\tat ")
	io.WriteString(w, shortname(f.Function))
	io.WriteString(w, " (")
	io.WriteString(w, f.File)
	io.WriteString(w, ":")
	io.WriteString(w, strconv.Itoa(f.Line))
	io.WriteString(w, ")")
}

// Callers is a list of program counters returned by the runtime.Callers.
type Callers []uintptr

// Frames returns a slice of structures with a function/file/line information.
func (c Callers) Frames() []Frame {
	r := make([]Frame, len(c))
	f := runtime.CallersFrames(c)
	n := 0
	for {
		frame, more := f.Next()
		r[n] = Frame{
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
		}
		if !more {
			break
		}
		n++
	}
	return r
}

// String implements the fmt.Stringer interface.
func (c Callers) String() string {
	s := &strings.Builder{}
	c.writeTrace(s)
	return s.String()
}

// Format implements the fmt.Formatter interface.
//
// The verbs:
//
//	%s	a stack trace
//	%v	same as %s, the plus or hash flags print struct details
//	%q	a double-quoted Go string with same contents as %s
func (c Callers) Format(s fmt.State, verb rune) {
	type _Callers Callers
	switch verb {
	case 's':
		c.writeTrace(s)
	case 'v':
		switch {
		case s.Flag('+') || s.Flag('#'):
			format(s, verb, _Callers(c))
		default:
			c.Format(s, 's')
		}
	case 'q':
		io.WriteString(s, strconv.Quote(c.String()))
	default:
		format(s, verb, _Callers(c))
	}
}

func (c Callers) writeTrace(w io.Writer) {
	frames := c.Frames()
	for _, frame := range frames {
		frame.writeFrame(w)
		io.WriteString(w, "\n")
	}
}

func callers(skip int) Callers {
	b := make([]uintptr, stackTraceDepth)
	l := runtime.Callers(skip+2, b[:])
	return b[:l]
}

func shortname(name string) string {
	i := strings.LastIndex(name, "/")
	return name[i+1:]
}
