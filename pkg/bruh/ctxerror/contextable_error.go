// Package ctxerror provides an error type that allows you to attach additional
// context to it.
package ctxerror

import (
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

// Contexter is an error that includes additional context in form of a map.
type Contexter interface {
	error
	Context() map[string]any
}

// ContextableErrorer is an error that can modify its context.
type ContextableErrorer interface {
	error
	Context() map[string]any
	FullContext() map[string]any
	RangeContext(func(key string, value any) bool)
	Add(keyValuePair ...any) ContextableErrorer
	AddAll(context map[string]any) ContextableErrorer
	Remove(key ...string) ContextableErrorer
}

type keyValue struct {
	key   string
	value any
}

var (
	_ Contexter          = (*ContextableError)(nil)
	_ ContextableErrorer = (*ContextableError)(nil)
)

// ContextableError is an error that let's you attach additional context to it.
// E.g. you can attach a request or user ID that you can later retrieve and
// write to your logs. The context is not part of the error message. You can
// retrieve it by calling the [*ContextableError.Context] or
// [*ContextableError.FullContext] method.
type ContextableError struct {
	bruh.TraceableError
	// using a slice is more efficient than a map for small sizes
	context []keyValue
}

// New creates a new [ContextableError] with the given message and the given
// key-value pairs.
func New(msg string, keyValuePair ...any) ContextableErrorer {
	cerr := &ContextableError{
		TraceableError: *bruh.NewSkip(1, msg),
		context:        make([]keyValue, 0, 8),
	}
	cerr.Add(keyValuePair...)
	return cerr
}

// NewSkip behaves like [New] but skips the given number of callers when
// creating a stack trace. You should only use this if you are implementing a
// new error type on top of [ContextableError].
func NewSkip(skip uint, msg string, keyValuePair ...any) *ContextableError {
	cerr := &ContextableError{
		TraceableError: *bruh.NewSkip(skip+1, msg),
		context:        make([]keyValue, 0, 8),
	}
	cerr.Add(keyValuePair...)
	return cerr
}

// Wrap wraps the given error by creating a new [ContextableError] with the
// specified message and the given key-value pairs.
func Wrap(err error, msg string, keyValuePair ...any) ContextableErrorer {
	cerr := &ContextableError{
		TraceableError: *bruh.WrapSkip(err, 1, msg),
		context:        make([]keyValue, 0, 8),
	}
	cerr.Add(keyValuePair...)
	return cerr
}

// WrapSkip behaves like [Wrap] but skips the given number of callers when
// creating a stack trace. You should only use this if you are implementing a
// new error type on top of [ContextableError].
func WrapSkip(err error, skip uint, msg string, keyValuePair ...any) *ContextableError {
	cerr := &ContextableError{
		TraceableError: *bruh.WrapSkip(err, skip+1, msg),
		context:        make([]keyValue, 0, 8),
	}
	cerr.Add(keyValuePair...)
	return cerr
}

// Add adds the given key-value pairs to the error context. Any key that
// already exists, will be overwritten.
func (e *ContextableError) Add(keyValuePair ...any) ContextableErrorer {
	l := len(keyValuePair) - len(keyValuePair)%2 // silently drop a key without a value
	for i := 0; i < l; i += 2 {
		if key, ok := keyValuePair[i].(string); ok {
			e.addPair(key, keyValuePair[i+1])
		} else {
			e.addPair(fmt.Sprint(keyValuePair[i]), keyValuePair[i+1])
		}
	}
	return e
}

// addPair adds the given single key-value pair to the error context. Any key
// that already exists, will be overwritten.
func (e *ContextableError) addPair(key string, value any) {
	// check if the key already exists and update the value if it does
	for i, kv := range e.context {
		if kv.key == key {
			e.context[i].value = value
			return
		}
	}
	e.context = append(e.context, keyValue{key, value})
}

// AddAll adds all key-value pairs to the error context.
func (e *ContextableError) AddAll(context map[string]any) ContextableErrorer {
	for key, value := range context {
		e.addPair(key, value)
	}
	return e
}

// Remove removes the given keys from the error context.
func (e *ContextableError) Remove(key ...string) ContextableErrorer {
	for i := 0; i < len(e.context); i++ {
		for _, k := range key {
			if e.context[i].key == k {
				e.context = append(e.context[:i], e.context[i+1:]...)
				i--
				break
			}
		}
	}
	return e
}

// Context returns the context for the error. It does not include context from
// other errors in the chain. If you want to get the full context, use
// [*ContextableError.FullContext] instead.
func (e *ContextableError) Context() map[string]any {
	ctx := make(map[string]any, len(e.context))
	for _, kv := range e.context {
		ctx[kv.key] = kv.value
	}
	return ctx
}

// FullContext returns the combined context of the whole error chain.
func (e *ContextableError) FullContext() map[string]any {
	return GetFullContext(e)
}

// RangeContext iterates over the context of the error and calls the given
// function for each key-value pair.
func (e *ContextableError) RangeContext(fn func(key string, value any) bool) {
	RangeContext(e, fn)
}

// -----------------------------------------------------------------------------
//
// Convenience Functions
//
// -----------------------------------------------------------------------------

// GetContext returns the attached context of the given error. An empty map is
// returned if the error does not have any context.
func GetContext(err error) map[string]any {
	if err == nil {
		return make(map[string]any)
	}
	if e, ok := err.(Contexter); ok {
		return e.Context()
	}
	return make(map[string]any)
}

// GetFullContext returns the attached context of the whole error chain. An
// empty map is returned if the error does not have any context.
func GetFullContext(err error) map[string]any {
	if err == nil {
		return make(map[string]any)
	}
	curErr := err
	// calculate size of the context map
	size := 0
	for {
		if curErr == nil {
			break
		}
		if e, ok := curErr.(*ContextableError); ok {
			size += len(e.context)
		} else if e2, ok := curErr.(Contexter); ok {
			size += len(e2.Context())
		}
		curErr = bruh.Unwrap(curErr)
	}
	ctx := make(map[string]any, size)
	// fill the context map
	for {
		if err == nil {
			break
		}
		if e, ok := err.(*ContextableError); ok {
			for _, kv := range e.context {
				ctx[kv.key] = kv.value
			}
		} else if e2, ok := err.(Contexter); ok {
			for k, v := range e2.Context() {
				ctx[k] = v
			}
		}
		err = bruh.Unwrap(err)
	}
	return ctx
}

// RangeContext iterates over the context of the given error and calls the given
// function for each key-value pair.
func RangeContext(err error, fn func(key string, value any) bool) {
	if err == nil {
		return
	}
	for {
		if err == nil {
			break
		}
		if e, ok := err.(*ContextableError); ok {
			for _, kv := range e.context {
				if !fn(kv.key, kv.value) {
					return
				}
			}
		} else if e2, ok := err.(Contexter); ok {
			for k, v := range e2.Context() {
				if !fn(k, v) {
					return
				}
			}
		}
		err = bruh.Unwrap(err)
	}
}
