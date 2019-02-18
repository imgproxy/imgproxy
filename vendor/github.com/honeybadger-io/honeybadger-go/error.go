package honeybadger

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
)

const maxFrames = 20

// Frame represent a stack frame inside of a Honeybadger backtrace.
type Frame struct {
	Number string `json:"number"`
	File   string `json:"file"`
	Method string `json:"method"`
}

// Error provides more structured information about a Go error.
type Error struct {
	err     interface{}
	Message string
	Class   string
	Stack   []*Frame
}

func (e Error) Error() string {
	return e.Message
}

func NewError(msg interface{}) Error {
	return newError(msg, 2)
}

func newError(thing interface{}, stackOffset int) Error {
	var err error

	switch t := thing.(type) {
	case Error:
		return t
	case error:
		err = t
	default:
		err = fmt.Errorf("%v", t)
	}

	return Error{
		err:     err,
		Message: err.Error(),
		Class:   reflect.TypeOf(err).String(),
		Stack:   generateStack(stackOffset),
	}
}

func generateStack(offset int) (frames []*Frame) {
	stack := make([]uintptr, maxFrames)
	length := runtime.Callers(2+offset, stack[:])
	for _, pc := range stack[:length] {
		f := runtime.FuncForPC(pc)
		if f == nil {
			continue
		}
		file, line := f.FileLine(pc)
		frame := &Frame{
			File:   file,
			Number: strconv.Itoa(line),
			Method: f.Name(),
		}
		frames = append(frames, frame)
	}

	return
}
