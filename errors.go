// Package errors provides errors that have stack-traces.
// This package is a copy from github.com/go-errors/errors

package errors

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
)

// The maximum number of stackframes on any error.
var MaxStackDepth = 50

// Error is an error with an attached stacktrace. It can be used
// wherever the builtin error interface is expected.
type CommonError struct {
	Err    error
	stack  []uintptr
	frames []StackFrame
	prefix string
}

type Error interface {
	error
	Stack() []byte
	ErrorStack() string
	StackFrames() []StackFrame
	TypeName() string
}

// New makes an Error from the given value. If that value is already an
// error then it will be used directly, if not, it will be passed to
// fmt.Errorf("%v"). The stacktrace will point to the line of code that
// called New.
func New(e interface{}) *CommonError {
	var err error
	switch e := e.(type) {
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}
	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2, stack[:])
	return &CommonError{
		Err:   err,
		stack: stack[:length],
	}
}

// Wrap makes an Error from the given value. If that value is already an
// error then it will be used directly, if not, it will be passed to
// fmt.Errorf("%v"). The skip parameter indicates how far up the stack
// to start the stacktrace. 0 is from the current call, 1 from its caller, etc.
func Wrap(e interface{}, skip int) *CommonError {
	var err error
	switch e := e.(type) {
	case *CommonError:
		return e
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}
	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2+skip, stack[:])
	return &CommonError{
		Err:   err,
		stack: stack[:length],
	}
}

func Unwrap(e error) error {
	var err error
	switch e := e.(type) {
	case *CommonError:
		err = e.Err
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}
	return err
}

// WrapPrefix makes an Error from the given value. If that value is already an
// error then it will be used directly, if not, it will be passed to
// fmt.Errorf("%v"). The prefix parameter is used to add a prefix to the
// error message when calling Error(). The skip parameter indicates how far
// up the stack to start the stacktrace. 0 is from the current call,
// 1 from its caller, etc.
func WrapPrefix(e interface{}, prefix string, skip int) *CommonError {
	err := Wrap(e, skip)
	if err.prefix != "" {
		err.prefix = fmt.Sprintf("%s: %s", prefix, err.prefix)
	} else {
		err.prefix = prefix
	}
	return err

}

// Is detects whether the error is equal to a given error. Errors
// are considered equal by this function if they are the same object,
// or if they both contain the same error inside an errors.Error.
func Is(e error, original error) bool {
	if e == original {
		return true
	}
	if e, ok := e.(*CommonError); ok {
		return Is(e.Err, original)
	}
	if original, ok := original.(*CommonError); ok {
		return Is(e, original.Err)
	}
	return false
}

// Errorf creates a new error with the given message. You can use it
// as a drop-in replacement for fmt.Errorf() to provide descriptive
// errors in return values.
func Errorf(format string, a ...interface{}) *CommonError {
	return Wrap(fmt.Errorf(format, a...), 1)
}

// Error returns the underlying error's message.
func (err *CommonError) Error() string {
	msg := err.Err.Error()
	if err.prefix != "" {
		msg = fmt.Sprintf("%s: %s", err.prefix, msg)
	}
	return msg
}

// Stack returns the callstack formatted the same way that go does
// in runtime/debug.Stack()
func (err *CommonError) Stack() []byte {
	var buf bytes.Buffer
	for _, frame := range err.StackFrames() {
		buf.WriteString(frame.String())
	}
	return buf.Bytes()
}

// ErrorStack returns a string that contains both the
// error message and the callstack.
func (err *CommonError) ErrorStack() string {
	return err.Error() + "\n" + string(err.Stack())
}

// TypeErrorStack returns a string that contains both the
// error message and the callstack.
func (err *CommonError) TypeErrorStack() string {
	return err.TypeName() + " " + err.Error() + "\n" + string(err.Stack())
}

// StackFrames returns an array of frames containing information about the
// stack.
func (err *CommonError) StackFrames() []StackFrame {
	if err.frames == nil {
		err.frames = make([]StackFrame, len(err.stack))
		for i, pc := range err.stack {
			err.frames[i] = NewStackFrame(pc)
		}
	}
	return err.frames
}

// TypeName returns the type this error. e.g. *errors.stringError.
func (err *CommonError) TypeName() string {
	if _, ok := err.Err.(uncaughtPanic); ok {
		return "panic"
	}
	return reflect.TypeOf(err.Err).String()
}
