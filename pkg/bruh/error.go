package bruh

import (
	"errors"
	"fmt"
	"io"
	"runtime"
)

// New creates a new [Err] with the given message.
func New(msg string) error {
	return NewSkip(1, msg)
}

// NewSkip behaves like [New] but skips the given number of callers when
// creating a stack trace. You should only use this if you are implementing a
// new error type on top of [Err].
func NewSkip(skip int, msg string) *Err {
	// skips this method, runtime.Callers and user defined number of other
	// callers
	berr := &Err{msg: msg}
	berr.stackSize = runtime.Callers(2+max(skip, 0), berr.stackStore[:])
	return berr
}

// Errorf creates a new [Err] with a formatted message.
func Errorf(format string, args ...any) error {
	return NewSkip(1, fmt.Sprintf(format, args...))
}

// ErrorfSkip behaves like [Errorf] but skips the given number of callers when
// creating a stack trace. It is intended for implementing custom error types on
// top of [Err].
func ErrorfSkip(skip int, format string, args ...any) *Err {
	return NewSkip(skip+1, fmt.Sprintf(format, args...))
}

// Wrap wraps the given error by creating a new [Err] with the
// specified message. If the given error is nil, nil is returned.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return WrapSkip(err, 1, msg)
}

// WrapSkip behaves like [Wrap] but skips the given number of callers when
// creating a stack trace. You should only use this if you are implementing a
// custom error type on top of [Err].
func WrapSkip(err error, skip int, msg string) *Err {
	if err == nil {
		return nil
	}
	// skips this method, runtime.Callers and user defined number of other
	// callers
	berr := &Err{msg: msg, err: err}
	berr.stackSize = runtime.Callers(2+max(skip, 0), berr.stackStore[:])
	return berr
}

// Wrapf wraps the given error by creating a new [Err] with a
// formatted message. If the given error is nil, nil is returned.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return WrapSkip(err, 1, fmt.Sprintf(format, args...))
}

// WrapfSkip behaves like [Wrapf] but skips the given number of callers when
// creating a stack trace. You should only use this if you are implementing a
// custom error type on top of [Err].
func WrapfSkip(err error, skip int, format string, args ...any) *Err {
	return WrapSkip(err, skip+1, fmt.Sprintf(format, args...))
}

// Err is an easily wrappable error with a stack trace.
type Err struct {
	msg string
	err error
	// we store the stack pointers inline with the error
	stackStore [MAX_STACK_DEPTH]uintptr
	stackSize  int
}

// Message returns the single, unformatted message of this error.
func (e *Err) Message() string {
	return e.msg
}

// Error returns the formatted error message including the messages of wrapped
// errors.
func (e *Err) Error() string {
	return Message(e)
}

// Format implements the fmt.Formatter interface. Use fmt.Sprintf("%v", err) to
// get a string representation of the error without an stack trace and
// fmt.Sprintf("%+v", err) with a stack trace included.
func (e *Err) Format(s fmt.State, verb rune) {
	var str string
	if verb == 'v' && s.Flag('+') {
		str = String(e)
	} else {
		str = Message(e)
	}
	_, _ = io.WriteString(s, str)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's type
// contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
func (e *Err) Unwrap() error {
	return e.err
}

// Cause returns the root cause of the error, which is defined as the first
// error in the chain.
func (e *Err) Cause() error {
	return Cause(e)
}

// Stack returns a combined stack trace of all errors in the chain.
func (e *Err) Stack() Stack {
	stack := *new4xStack()
	stack = stack[:combinedStack(e, stack)]
	return stack
}

// StackFrames is an alias for [*Err.Stack]. [Sentry] looks for
// this particularly named method.
//
// [Sentry]: https://github.com/getsentry/sentry-go
func (e *Err) StackFrames() Stack {
	return e.Stack()
}

// Callers returns the recorded caller stack. It implements Bugsnag's
// [ErrorWithCallers] interface.
//
// [ErrorWithCallers]: https://github.com/bugsnag/bugsnag-go/blob/46ba8d9aa46bb1d208bfcf408d0b5cff1fd371ab/v2/errors/error.go#L27-L30
func (e *Err) Callers() []uintptr {
	return e.stackStore[:e.stackSize]
}

// // Implements the redacter interface of github.com/aisbergg/go-redact and allows
// // sensitive information to be redacted from the error chain.
// func (e *Err) Redact(redacter interface{ Redact(value any) }) {
// 	redacter.Redact(e.err)
// }

// -----------------------------------------------------------------------------
//
// Convenience Functions
//
// -----------------------------------------------------------------------------

// Unwrap returns the result of calling the Unwrap method on err, if err's type
// contains an [errors.Unwrap] method returning error. Otherwise, Unwrap returns
// nil.
//
// See Go's [errors.Unwrap] for more information.
func Unwrap(err error) error {
	if u, ok := err.(unwraper); ok {
		return u.Unwrap()
	}
	return nil
}

// Cause returns the root cause of the error, which is defined as the first
// error in the chain. The original error is returned if it does not implement
// [errors.Unwrap] and nil is returned if the error is nil.
func Cause(err error) error {
	for {
		unwrappedErr := Unwrap(err)
		if unwrappedErr == nil {
			return err
		}
		err = unwrappedErr
	}
}

// As finds the first error in err's tree that matches target, and if one is found, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// More information can be found in Go's [errors.As] documentation.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is reports whether any error in err's tree matches target.
//
// More information can be found in Go's [errors.Is] documentation.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// -----------------------------------------------------------------------------
//
// Interfaces
//
// -----------------------------------------------------------------------------

type callerser interface {
	Callers() []uintptr
}

type messager interface {
	Message() string
}

type unwraper interface {
	Unwrap() error
}
