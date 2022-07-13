package errors

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// New creates a new error with the given message.
func New(msg string) *TraceableError {
	return NewSkip(1, msg)
}

// NewSkip creates a new error with the given message and skips the
// specified number of callers in the stack trace.
func NewSkip(skip uint, msg string) *TraceableError {
	// skips this method, stack.callers, runtime.Callers and user defined number
	// of other callers
	stack := callers(3 + skip)

	// strip stack for globally defined errors
	if stack.isGlobal() {
		return &TraceableError{
			msg: msg,
		}
	}

	return &TraceableError{
		msg:   msg,
		stack: stack,
	}
}

// Errorf creates a new error with a formatted message.
func Errorf(format string, args ...interface{}) *TraceableError {
	return NewSkip(1, fmt.Sprintf(format, args...))
}

// ErrorfSkip creates a new error with a formatted message and skips the
// specified number of callers in the stack trace.
func ErrorfSkip(skip uint, format string, args ...interface{}) *TraceableError {
	return NewSkip(skip+1, fmt.Sprintf(format, args...))
}

// Wrap wraps the given error by creating a new error with the specified
// message.
func Wrap(err error, msg string) *TraceableError {
	return wrap(err, 0, fmt.Sprint(msg))
}

// WrapSkip wraps the given error by creating a new error with the specified
// message and skips the specified number of callers in the stack trace.
func WrapSkip(err error, skip uint, msg string) *TraceableError {
	return wrap(err, skip, fmt.Sprint(msg))
}

// Wrapf wraps the given error by creating a new error with a formatted message.
func Wrapf(err error, format string, args ...interface{}) *TraceableError {
	return wrap(err, 0, fmt.Sprintf(format, args...))
}

// WrapfSkip wraps the given error by creating a new error with a formatted
// message and skips the specified number of callers in the stack trace.
func WrapfSkip(err error, skip uint, format string, args ...interface{}) *TraceableError {
	return wrap(err, skip, fmt.Sprintf(format, args...))
}

func wrap(err error, skip uint, msg string) *TraceableError {
	if err == nil {
		return nil
	}

	return &TraceableError{
		msg:   msg,
		err:   err,
		stack: callers(4 + skip),
	}
}

// TraceableError is an easily wrappable error with stack trace.
type TraceableError struct {
	msg   string
	err   error
	stack stackPC
}

// Error returns the error message.
func (e *TraceableError) Error() string {
	return e.msg
}

// Format implements the fmt.Formatter interface. Use fmt.Sprintf("%v", err) to
// get a string representation of the error without an stack trace and
// fmt.Sprintf("%+v", err) with a stack trace included.
func (e *TraceableError) Format(s fmt.State, verb rune) {
	var withTrace bool
	switch verb {
	case 'v':
		if s.Flag('+') {
			withTrace = true
		}
	}
	str := ToString(e, withTrace)
	io.WriteString(s, str)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's type
// contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
func (e *TraceableError) Unwrap() error {
	return e.err
}

// Is reports whether any error in err's chain matches target.
func (e *TraceableError) Is(target error) bool {
	return Is(e, target)
}

// As finds the first error in err's chain that matches target, and if one is
// found, sets target to that error value and returns true. Otherwise, it
// returns false.
func (e *TraceableError) As(target interface{}) bool {
	return As(e, target)
}

// Cause returns the root cause of the error, which is defined as the first
// error in the chain.
func (e *TraceableError) Cause() error {
	return Cause(e)
}

// Stack returns the stack trace for thins error instance in form of a list of
// StackFrames.
func (e *TraceableError) Stack() Stack {
	return e.stack.toStack()
}

// FullStack returns a combined stack trace of all errors in err's chain.
func (e *TraceableError) FullStack() Stack {
	cbdStack := e.stack
	var uerr error = e
	for {
		uerr = Unwrap(uerr)
		if uerr == nil {
			break
		}

		terr, ok := uerr.(interface{ stackPC() stackPC })
		if !ok {
			break
		}

		nxtStack := terr.stackPC()
		nxtStack = nxtStack.RelativeTo(cbdStack)
		cbdStack = append(nxtStack, cbdStack...)
	}

	return cbdStack.toStack()
}

// StackFrames is an alias for FullStack. getsentry/sentry-go looks for this
// particularly named method.
func (e *TraceableError) StackFrames() Stack {
	return e.FullStack()
}

// TypeName returns the type of this error. e.g. errors.TraceableError.
func (e *TraceableError) TypeName() string {
	return TypeName(e)
}

// Stack returns the stack trace of the error in the form of a program counter
// slice.
func (e *TraceableError) stackPC() stackPC {
	return e.stack
}

// -----------------------------------------------------------------------------
//
// Convenience Functions
//
// -----------------------------------------------------------------------------

// Unwrap returns the result of calling the Unwrap method on err, if err's type
// contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
//
// See Go's errors.Unwrap for more information.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Is reports whether any error in err's chain matches target.
//
// See Go's errors.Is for more information.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if one is
// found, sets target to that error value and returns true. Otherwise, it
// returns false.
//
// See Go's errors.As for more information.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Cause returns the root cause of the error, which is defined as the first
// error in the chain. The original error is returned if it does not implement
// `Unwrap() error` and nil is returned if the error is nil.
func Cause(err error) error {
	for {
		uerr := Unwrap(err)
		if uerr == nil {
			return err
		}
		err = uerr
	}
}

// TypeName returns the type of the error. e.g. errors.TraceableError.
func TypeName(err error) string {
	typeName := reflect.TypeOf(err).String()
	typeName = strings.TrimPrefix(typeName, "*")
	return typeName
}
