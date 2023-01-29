package bruh

// Contexter is an error that includes additional context in form of a map.
type Contexter interface {
	error
	Context() map[string]interface{}
}

// ContextAdder is an error that can add context to itself.
type ContextAdder interface {
	error
	Add(key string, value interface{}) ContextAdder
	AddAll(context map[string]interface{}) ContextAdder
}

var _ Contexter = (*ContextableError)(nil)

// ContextableError is an error that is temporary and can be retried.
type ContextableError struct {
	TraceableError
	context map[string]interface{}
}

// CENew creates a new ContextableError error with the given message.
func CENew(msg string) *ContextableError {
	return &ContextableError{
		TraceableError: *NewSkip(1, msg),
		context:        make(map[string]interface{}),
	}
}

// CENewSkip creates a new ContextableError error with the given message and
// skips the specified number of callers in the stack trace.
func CENewSkip(skip uint, msg string) *ContextableError {
	return &ContextableError{
		TraceableError: *NewSkip(skip+1, msg),
		context:        make(map[string]interface{}),
	}
}

// CEErrorf creates a new ContextableError error with a formatted message.
func CEErrorf(format string, args ...interface{}) *ContextableError {
	return &ContextableError{
		TraceableError: *ErrorfSkip(1, format, args...),
		context:        make(map[string]interface{}),
	}
}

// CEErrorfSkip creates a new ContextableError error with a formatted message
// and skips the specified number of callers in the stack trace.
func CEErrorfSkip(skip uint, format string, args ...interface{}) *ContextableError {
	return &ContextableError{
		TraceableError: *ErrorfSkip(skip+1, format, args...),
		context:        make(map[string]interface{}),
	}
}

// CEWrap wraps the given error by creating a new ContextableError error with
// the specified message.
func CEWrap(err error, msg string) *ContextableError {
	return &ContextableError{
		TraceableError: *WrapSkip(err, 1, msg),
		context:        make(map[string]interface{}),
	}
}

// CEWrapSkip wraps the given error by creating a new ContextableError error
// with the specified message and skips the specified number of callers in the
// stack trace.
func CEWrapSkip(err error, skip uint, msg string) *ContextableError {
	return &ContextableError{
		TraceableError: *WrapSkip(err, skip+1, msg),
		context:        make(map[string]interface{}),
	}
}

// CEWrapf wraps the given error by creating a new ContextableError error with a
// formatted message.
func CEWrapf(err error, format string, args ...interface{}) *ContextableError {
	return &ContextableError{
		TraceableError: *WrapfSkip(err, 1, format, args...),
		context:        make(map[string]interface{}),
	}
}

// CEWrapfSkip wraps the given error by creating a new ContextableError error
// with a formatted message and skips the specified number of callers in the
// stack trace.
func CEWrapfSkip(err error, skip uint, format string, args ...interface{}) *ContextableError {
	return &ContextableError{
		TraceableError: *WrapfSkip(err, skip+1, format, args...),
		context:        make(map[string]interface{}),
	}
}

// Add adds a key-value pair to the error context. If the key already exists, it will be overwritten. If the value is nil, the key will be removed.
func (e *ContextableError) Add(key string, value interface{}) *ContextableError {
	if value == nil {
		delete(e.context, key)
		return e
	}
	e.context[key] = value
	return e
}

// AddAll adds all key-value pairs to the error context.
func (e *ContextableError) AddAll(context map[string]interface{}) *ContextableError {
	for key, value := range context {
		e.Add(key, value)
	}
	return e
}

// Context returns the context for the error.
func (e *ContextableError) Context() map[string]interface{} {
	return e.context
}

// FullContext returns the context of the whole error chain.
func (e *ContextableError) FullContext() map[string]interface{} {
	return GetFullContext(e)
}

// -----------------------------------------------------------------------------
//
// Convenience Functions
//
// -----------------------------------------------------------------------------

// GetContext returns the attached context of the given error. If the error does
// not have any context, nil is returned.
func GetContext(err error) map[string]interface{} {
	if e, ok := err.(Contexter); ok {
		return e.Context()
	}
	return nil
}

// GetFullContext returns the attached context of the whole error chain. If the
// error does not have any context, nil is returned.
func GetFullContext(err error) map[string]interface{} {
	if err == nil {
		return make(map[string]interface{})
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
