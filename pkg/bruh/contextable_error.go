package bruh

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
// retrieve it by calling the [*ContextableError.Context] or [*ContextableError.FullContext]` method.
type ContextableError struct {
	TraceableError
	context map[string]any
}

// CENew creates a new [ContextableError] with the given message.
func CENew(msg string) error {
	return &ContextableError{
		TraceableError: *NewSkip(1, msg).(*TraceableError),
		context:        make(map[string]any),
	}
}

// CENewSkip creates a new [ContextableError] with the given message and
// skips the specified number of callers in the stack trace.
func CENewSkip(skip uint, msg string) error {
	return &ContextableError{
		TraceableError: *NewSkip(skip+1, msg).(*TraceableError),
		context:        make(map[string]any),
	}
}

// CEErrorf creates a new [ContextableError] with a formatted message.
func CEErrorf(format string, args ...any) error {
	return &ContextableError{
		TraceableError: *ErrorfSkip(1, format, args...).(*TraceableError),
		context:        make(map[string]any),
	}
}

// CEErrorfSkip creates a new [ContextableError] with a formatted message
// and skips the specified number of callers in the stack trace.
func CEErrorfSkip(skip uint, format string, args ...any) error {
	return &ContextableError{
		TraceableError: *ErrorfSkip(skip+1, format, args...).(*TraceableError),
		context:        make(map[string]any),
	}
}

// CEWrap wraps the given error by creating a new [ContextableError] with
// the specified message.
func CEWrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &ContextableError{
		TraceableError: *WrapSkip(err, 1, msg).(*TraceableError),
		context:        make(map[string]any),
	}
}

// CEWrapSkip wraps the given error by creating a new [ContextableError]
// with the specified message and skips the specified number of callers in the
// stack trace.
func CEWrapSkip(err error, skip uint, msg string) error {
	if err == nil {
		return nil
	}
	return &ContextableError{
		TraceableError: *WrapSkip(err, skip+1, msg).(*TraceableError),
		context:        make(map[string]any),
	}
}

// CEWrapf wraps the given error by creating a new [ContextableError] with a
// formatted message.
func CEWrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &ContextableError{
		TraceableError: *WrapfSkip(err, 1, format, args...).(*TraceableError),
		context:        make(map[string]any),
	}
}

// CEWrapfSkip wraps the given error by creating a new [ContextableError]
// with a formatted message and skips the specified number of callers in the
// stack trace.
func CEWrapfSkip(err error, skip uint, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &ContextableError{
		TraceableError: *WrapfSkip(err, skip+1, format, args...).(*TraceableError),
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
	ctx := GetFullContext(Unwrap(err))
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
