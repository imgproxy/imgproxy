package internal

import (
	"bytes"
	"path"
	"runtime"
)

// StackTrace is a stack trace.
type StackTrace []uintptr

// GetStackTrace returns a new StackTrace.
func GetStackTrace(skipFrames int) StackTrace {
	skip := 2 // skips runtime.Callers and this function
	skip += skipFrames

	callers := make([]uintptr, maxStackTraceFrames)
	written := runtime.Callers(skip, callers)
	return StackTrace(callers[0:written])
}

func pcToFunc(pc uintptr) (*runtime.Func, uintptr) {
	// The Golang runtime package documentation says "To look up the file
	// and line number of the call itself, use pc[i]-1. As an exception to
	// this rule, if pc[i-1] corresponds to the function runtime.sigpanic,
	// then pc[i] is the program counter of a faulting instruction and
	// should be used without any subtraction."
	//
	// TODO: Fully understand when this subtraction is necessary.
	place := pc - 1
	return runtime.FuncForPC(place), place
}

func topCallerNameBase(st StackTrace) string {
	f, _ := pcToFunc(st[0])
	if nil == f {
		return ""
	}
	return path.Base(f.Name())
}

// WriteJSON adds the stack trace to the buffer in the JSON form expected by the
// collector.
func (st StackTrace) WriteJSON(buf *bytes.Buffer) {
	buf.WriteByte('[')
	for i, pc := range st {
		// Stack traces may be provided by the customer, and therefore
		// may be excessively long.  The truncation is done here to
		// facilitate testing.
		if i >= maxStackTraceFrames {
			break
		}
		if i > 0 {
			buf.WriteByte(',')
		}
		// Implements the format documented here:
		// https://source.datanerd.us/agents/agent-specs/blob/master/Stack-Traces.md
		buf.WriteByte('{')
		if f, place := pcToFunc(pc); nil != f {
			name := path.Base(f.Name())
			file, line := f.FileLine(place)

			w := jsonFieldsWriter{buf: buf}
			w.stringField("filepath", file)
			w.stringField("name", name)
			w.intField("line", int64(line))
		}
		buf.WriteByte('}')
	}
	buf.WriteByte(']')
}

// MarshalJSON prepares JSON in the format expected by the collector.
func (st StackTrace) MarshalJSON() ([]byte, error) {
	estimate := 256 * len(st)
	buf := bytes.NewBuffer(make([]byte, 0, estimate))

	st.WriteJSON(buf)

	return buf.Bytes(), nil
}
