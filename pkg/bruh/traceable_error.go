package bruh

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// New creates a new TraceableError error with the given message.
func New(msg string) *TraceableError {
	return NewSkip(1, msg)
}

// NewSkip creates a new TraceableError error with the given message and skips
// the specified number of callers in the stack trace.
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

// Errorf creates a new TraceableError error with a formatted message.
func Errorf(format string, args ...interface{}) *TraceableError {
	return NewSkip(1, fmt.Sprintf(format, args...))
}

// ErrorfSkip creates a new TraceableError error with a formatted message and
// skips the specified number of callers in the stack trace.
func ErrorfSkip(skip uint, format string, args ...interface{}) *TraceableError {
	return NewSkip(skip+1, fmt.Sprintf(format, args...))
}

// Wrap wraps the given error by creating a new TraceableError error with the
// specified message.
func Wrap(err error, msg string) *TraceableError {
	return WrapSkip(err, 1, msg)
}

// WrapSkip wraps the given error by creating a new TraceableError error with
// the specified message and skips the specified number of callers in the stack
// trace.
func WrapSkip(err error, skip uint, msg string) *TraceableError {
	// skips this method, stack.callers, runtime.Callers and user defined number
	// of other callers
	stack := callers(3 + skip)

	// strip stack for globally defined errors
	if stack.isGlobal() {
		return &TraceableError{
			msg: msg,
			err: err,
		}
	}

	return &TraceableError{
		msg:   msg,
		err:   err,
		stack: stack,
	}
}

// Wrapf wraps the given error by creating a new TraceableError error with a
// formatted message.
func Wrapf(err error, format string, args ...interface{}) *TraceableError {
	return WrapSkip(err, 1, fmt.Sprintf(format, args...))
}

// WrapfSkip wraps the given error by creating a new TraceableError error with a
// formatted message and skips the specified number of callers in the stack
// trace.
func WrapfSkip(err error, skip uint, format string, args ...interface{}) *TraceableError {
	return WrapSkip(err, skip+1, fmt.Sprintf(format, args...))
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

// Cause returns the root cause of the error, which is defined as the first
// error in the chain.
func (e *TraceableError) Cause() error {
	return Cause(e)
}

// Stack returns the stack trace up to this error.
func (e *TraceableError) Stack() Stack {
	return e.stack.toStack()
}

// FullStack returns a combined stack trace of all errors in err's chain.
func (e *TraceableError) FullStack() Stack {
	var uerr error = e

	// unwrap error stack
	errs := make([]interface{ stackPC() stackPC }, 0, 30)
	errs = append(errs, e)
	for {
		uerr = Unwrap(uerr)
		if uerr == nil {
			break
		}
		terr, ok := uerr.(interface{ stackPC() stackPC })
		if !ok {
			break
		}
		errs = append(errs, terr)
	}

	cbdStk := errs[len(errs)-1].stackPC()
	for i := len(errs) - 2; i >= 0; i-- {
		curStk := errs[i].stackPC()
		relStk := cbdStk.relativeTo(curStk)
		cbdStk = append(relStk, curStk...)
	}

	return cbdStk.toStack()
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

// stackPC returns a copy of the program counters of function invocations.
func (e *TraceableError) stackPC() stackPC {
	stkCpy := make(stackPC, len(e.stack))
	copy(stkCpy, e.stack)
	return stkCpy
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
