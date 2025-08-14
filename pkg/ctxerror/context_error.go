// Package ctxerror provides an error type that allows you to attach additional
// context to it.
package ctxerror

import (
	"fmt"
	"log/slog"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

// Contexter is an error that includes additional context in form of a map.
type Contexter interface {
	error
	Context() map[string]any
}

// ModifiableContexter is an error that can modify its context.
type ModifiableContexter interface {
	Contexter
	Add(keyValuePair ...any) ModifiableContexter
	AddAll(context map[string]any) ModifiableContexter
	Remove(key ...string) ModifiableContexter
}

var (
	_ Contexter           = (*Err)(nil)
	_ ModifiableContexter = (*Err)(nil)
)

// Err is an error that let's you attach additional context to it.
// E.g. you can attach a request or user ID that you can later retrieve and
// write to your logs. The context is not part of the error message. You can
// retrieve it by calling the [*Err.Context] or
// [*Err.FullContext] method.
type Err struct {
	bruh.Err
	// using a slice is more efficient than a map for small sizes
	context map[string]any
}

// New creates a new [Err] with the given message and the given
// key-value pairs.
func New(msg string, keyValuePair ...any) ModifiableContexter {
	cerr := &Err{
		Err:     *bruh.NewSkip(1, msg),
		context: make(map[string]any, 12),
	}
	cerr.Add(keyValuePair...) //nolint:errcheck
	return cerr
}

// NewSkip behaves like [New] but skips the given number of callers when
// creating a stack trace. You should only use this if you are implementing a
// new error type on top of [Err].
func NewSkip(skip int, msg string, keyValuePair ...any) *Err {
	cerr := &Err{
		Err:     *bruh.NewSkip(skip+1, msg),
		context: make(map[string]any, 12),
	}
	cerr.Add(keyValuePair...) //nolint:errcheck
	return cerr
}

// Wrap wraps the given error by creating a new [Err] with the specified message
// and the given key-value pairs. If the given error is nil, nil is returned.
func Wrap(err error, msg string, keyValuePair ...any) ModifiableContexter {
	if err == nil {
		return nil
	}
	cerr := &Err{
		Err:     *bruh.WrapSkip(err, 1, msg),
		context: make(map[string]any, 12),
	}
	cerr.Add(keyValuePair...) //nolint:errcheck
	return cerr
}

// WrapSkip behaves like [Wrap] but skips the given number of callers when
// creating a stack trace. You should only use this if you are implementing a
// new error type on top of [Err].
func WrapSkip(err error, skip int, msg string, keyValuePair ...any) *Err {
	if err == nil {
		return nil
	}
	cerr := &Err{
		Err:     *bruh.WrapSkip(err, skip+1, msg),
		context: make(map[string]any, 12),
	}
	cerr.Add(keyValuePair...) //nolint:errcheck
	return cerr
}

// Add adds the given key-value pairs to the error context. Any key that
// already exists, will be overwritten.
func (e *Err) Add(keyValuePair ...any) ModifiableContexter {
	l := len(keyValuePair) - len(keyValuePair)%2 // silently drop a key without a value
	for i := 0; i < l; i += 2 {
		if key, ok := keyValuePair[i].(string); ok {
			e.context[key] = keyValuePair[i+1]
		} else {
			e.context[fmt.Sprint(keyValuePair[i])] = keyValuePair[i+1]
		}
	}
	return e
}

// AddAll adds all key-value pairs to the error context.
func (e *Err) AddAll(context map[string]any) ModifiableContexter {
	for key, value := range context {
		e.context[key] = value
	}
	return e
}

// Remove removes the given keys from the error context.
func (e *Err) Remove(key ...string) ModifiableContexter {
	for _, k := range key {
		delete(e.context, k)
	}
	return e
}

// Context returns the context for the error. It does not include context from
// other errors in the chain. If you want to get the full context, use
// [GetContext] instead.
func (e *Err) Context() map[string]any {
	return e.context
}

// // Implements the redacter interface of github.com/aisbergg/go-redact and allows
// // sensitive information to be redacted from the error chain.
// func (e *Err) Redact(redacter interface{ Redact(value any) }) {
// 	redacter.Redact(e.context)
// 	// redact other errors down the chain
// 	e.Err.Redact(redacter)
// }

// -----------------------------------------------------------------------------
//
// Convenience Functions
//
// -----------------------------------------------------------------------------

// GetContext returns the full (combined) context of the given error chain.
func GetContext(err error) map[string]any {
	if err == nil {
		return make(map[string]any)
	}
	ctx := make(map[string]any, 32)
	// fill the context map
	for ; err != nil; err = bruh.Unwrap(err) {
		if e, ok := err.(Contexter); ok {
			for k, v := range e.Context() {
				if _, ok := ctx[k]; !ok {
					ctx[k] = v
				}
			}
		}
	}
	return ctx
}

// RangeContext iterates over the full context of the given error chain and
// calls the given function for each key-value pair.
func RangeContext(err error, fn func(key string, value any) bool) {
	IterContext(err)(fn)
}

// IterContext iterates over the full context of the given error chain. It is
// much like [RangeContext], but adheres to the function definition of Gos new
// range over function iterators.
//
// Example:
//
//	err := ctxerror.New("test error", "key1", "value1", "key2", "value2")
//	for k, v := range IterContext(err) {
//	  fmt.Printf("%s: %v\n", k, v)
//	}
func IterContext(err error) func(yield func(k string, value any) bool) {
	return func(yield func(k string, value any) bool) {
		yielded := make(map[string]struct{}, 32)
		for ; err != nil; err = bruh.Unwrap(err) {
			if e, ok := err.(Contexter); ok {
				for k, v := range e.Context() {
					if _, ok := yielded[k]; ok {
						continue
					}
					if !yield(k, v) {
						return
					}
					yielded[k] = struct{}{}
				}
			}
		}
	}
}

// AsSLogValue returns turns the given error into a slog.Value that can be used
// for structured logging.
func AsSLogValue(err error) slog.Value {
	if err == nil {
		return slog.Value{}
	}
	ctx := make(map[string]any, 32)
	ctx["message"] = err.Error()
	// fill the context map
	for ; err != nil; err = bruh.Unwrap(err) {
		switch e := err.(type) {
		case slog.LogValuer:
			value := e.LogValue()
			kind := value.Kind()
			if kind == slog.KindGroup {
				for _, attr := range value.Group() {
					if _, ok := ctx[attr.Key]; !ok {
						ctx[attr.Key] = attr.Value
					}
				}
			}
		case Contexter:
			for k, v := range e.Context() {
				if _, ok := ctx[k]; !ok {
					ctx[k] = v
				}
			}
		}
	}

	// convert the map to a slog.Value
	attrs := make([]slog.Attr, 0, len(ctx))
	for k, v := range ctx {
		attrs = append(attrs, slog.Any(k, v))
	}
	return slog.GroupValue(attrs...)
}
