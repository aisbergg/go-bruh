// Package ctxerror provides an error type that allows you to attach additional
// context to it.
package ctxerror

import "github.com/aisbergg/go-bruh/pkg/bruh"

// Contexter is an error that includes additional context in form of a map.
type Contexter interface {
	error
	Context() map[string]any
}

// ContextAdder is an error that can add context to itself.
type ContextAdder interface {
	error
	Add(key string, value any) ContextAdder
	AddAll(context map[string]any) ContextAdder
}

var _ Contexter = (*ContextableError)(nil)

// ContextableError is an error that let's you attach additional context to it.
// E.g. you can attach a request or user ID that you can later retrieve and
// write to your logs. The context is not part of the error message. You can
// retrieve it by calling the [*ContextableError.Context] or
// [*ContextableError.FullContext]` method.
type ContextableError struct {
	bruh.TraceableError
	context map[string]any
}

// New creates a new [ContextableError] with the given message.
func New(msg string) error {
	return &ContextableError{
		TraceableError: *bruh.NewSkip(1, msg).(*bruh.TraceableError),
		context:        make(map[string]any),
	}
}

// NewSkip creates a new [ContextableError] with the given message and
// skips the specified number of callers in the stack trace.
func NewSkip(skip uint, msg string) error {
	return &ContextableError{
		TraceableError: *bruh.NewSkip(skip+1, msg).(*bruh.TraceableError),
		context:        make(map[string]any),
	}
}

// Errorf creates a new [ContextableError] with a formatted message.
func Errorf(format string, args ...any) error {
	return &ContextableError{
		TraceableError: *bruh.ErrorfSkip(1, format, args...).(*bruh.TraceableError),
		context:        make(map[string]any),
	}
}

// ErrorfSkip creates a new [ContextableError] with a formatted message and
// skips the specified number of callers in the stack trace.
func ErrorfSkip(skip uint, format string, args ...any) error {
	return &ContextableError{
		TraceableError: *bruh.ErrorfSkip(skip+1, format, args...).(*bruh.TraceableError),
		context:        make(map[string]any),
	}
}

// Wrap wraps the given error by creating a new [ContextableError] with the
// specified message.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &ContextableError{
		TraceableError: *bruh.WrapSkip(err, 1, msg).(*bruh.TraceableError),
		context:        make(map[string]any),
	}
}

// WrapSkip wraps the given error by creating a new [ContextableError] with the
// specified message and skips the specified number of callers in the stack
// trace.
func WrapSkip(err error, skip uint, msg string) error {
	if err == nil {
		return nil
	}
	return &ContextableError{
		TraceableError: *bruh.WrapSkip(err, skip+1, msg).(*bruh.TraceableError),
		context:        make(map[string]any),
	}
}

// Wrapf wraps the given error by creating a new [ContextableError] with a
// formatted message.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &ContextableError{
		TraceableError: *bruh.WrapfSkip(err, 1, format, args...).(*bruh.TraceableError),
		context:        make(map[string]any),
	}
}

// WrapfSkip wraps the given error by creating a new [ContextableError] with a
// formatted message and skips the specified number of callers in the stack
// trace.
func WrapfSkip(err error, skip uint, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &ContextableError{
		TraceableError: *bruh.WrapfSkip(err, skip+1, format, args...).(*bruh.TraceableError),
		context:        make(map[string]any),
	}
}

// Add adds a key-value pair to the error context. If the key already exists, it
// will be overwritten. If the value is nil, the key will be removed.
func (e *ContextableError) Add(key string, value any) error {
	if value == nil {
		delete(e.context, key)
		return e
	}
	e.context[key] = value
	return e
}

// AddAll adds all key-value pairs to the error context.
func (e *ContextableError) AddAll(context map[string]any) error {
	for key, value := range context {
		_ = e.Add(key, value)
	}
	return e
}

// Context returns the context for the error. It does not include context from
// other errors in the chain. If you want to get the full context, use
// [*ContextableError.FullContext] instead.
func (e *ContextableError) Context() map[string]any {
	return e.context
}

// FullContext returns the combined context of the whole error chain.
func (e *ContextableError) FullContext() map[string]any {
	return GetFullContext(e)
}

// -----------------------------------------------------------------------------
//
// Convenience Functions
//
// -----------------------------------------------------------------------------

// GetContext returns the attached context of the given error. If the error does
// not have any context, nil is returned.
func GetContext(err error) map[string]any {
	if e, ok := err.(Contexter); ok {
		return e.Context()
	}
	return nil
}

// GetFullContext returns the attached context of the whole error chain. If the
// error does not have any context, nil is returned.
func GetFullContext(err error) map[string]any {
	if err == nil {
		return make(map[string]any)
	}
	ctx := GetFullContext(bruh.Unwrap(err))
	if e, ok := err.(Contexter); ok {
		for k, v := range e.Context() {
			ctx[k] = v
		}
	}
	if len(ctx) == 0 {
		return nil
	}
	return ctx
}
