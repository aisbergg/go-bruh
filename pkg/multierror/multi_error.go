// Package multierror provides an error type that can hold multiple errors.
package multierror

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"strconv"

	"github.com/aisbergg/go-bruh/pkg/bruh"
	"github.com/aisbergg/go-bruh/pkg/bruh/fmthelper"
	"github.com/aisbergg/go-bruh/pkg/ctxerror"
)

// UnwrapBehavior controls how multiple wrapped errors are unwrapped.
type UnwrapBehavior int

const (
	// UnwrapFirst returns only the first wrapped error when unwrapping.
	UnwrapFirst UnwrapBehavior = iota
	// UnwrapLast returns only the last wrapped error when unwrapping.
	UnwrapLast
	// UnwrapNone stops unwrapping at the multi error.
	UnwrapNone
)

// DefaultErrorsCapacity is the initial capacity for the errors slice.
const DefaultErrorsCapacity = 12

// FilterFunc is a predicate used to filter errors.
type FilterFunc func(error) bool

// MultiErrorer is an interface that is implemented by the [Err] type. It is
// used to distinguish between errors created by this package and other errors.
// It is not intended to be used by users of this package.
type MultiErrorer interface {
	error
	bruhMultiError()
	Message() string

	// Error collection access / mutation
	Errors() []error
	Grow(capacity int)
	Add(err ...error)
	Merge(err ...MultiErrorer)
	Unwrap() error

	// Nil/return helpers
	IsNil() bool
	ErrorOrNil() MultiErrorer
	SingleOrNil() error
}

// Options configures behavior of an [Err] created by [New] or [Errorf].
type Options struct {
	UnwrapBehavior UnwrapBehavior
	LimitPrint     int
	Filter         FilterFunc
}

// Err is an error that can hold multiple errors.
type Err struct {
	bruh.Err
	msg            string
	errors         []error
	unwrapBehavior UnwrapBehavior
	limitPrint     int
	filter         FilterFunc
}

// New creates a new [Err] with the given message and options. You can add more
// errors later by calling [Add].
func New(msg string, options ...Options) MultiErrorer {
	var opts Options
	if len(options) > 0 {
		opts = options[0]
	}
	merr := &Err{
		Err:            *bruh.NewSkip(1, ""),
		msg:            msg,
		unwrapBehavior: opts.UnwrapBehavior,
		limitPrint:     opts.LimitPrint,
		filter:         opts.Filter,
	}
	return merr
}

// Errorf creates a new [Err] with a formatted message.
func Errorf(options Options, format string, args ...any) MultiErrorer {
	return New(fmt.Sprintf(format, args...), options)
}

// -----------------------------------------------------------------------------

// bruhMultiError is a marker method to indicate that this error type is a multi
// error.
func (me *Err) bruhMultiError() {}

// Grow grows the internal slice of errors so it can accommodate at least the
// given number of errors without allocating.
func (me *Err) Grow(capacity int) {
	capacity = max(capacity, len(me.errors))
	if me.errors == nil {
		me.errors = make([]error, 0, capacity)
		return
	}
	if capacity <= cap(me.errors) {
		return
	}
	newErrors := make([]error, len(me.errors), capacity)
	copy(newErrors, me.errors)
	me.errors = newErrors
}

// Errors returns the errors that are stored in the [Err]. Might be nil if no
// errors have been added yet. Do not modify the returned slice.
func (me *Err) Errors() []error {
	return me.errors
}

func (me *Err) Error() string {
	if me.IsNil() {
		return ""
	}
	limit := me.limitPrint
	if limit <= 0 || limit > len(me.errors) {
		limit = len(me.errors)
	}
	message := me.msg
	builder := fmthelper.New([]byte{})
	guessCap := len(message) + limit*180
	builder.Grow(guessCap)
	if message != "" {
		builder.WriteString(message)
		builder.WriteByte('\n')
	}
	digits := fmthelper.DigitsInNumber(len(me.errors))
	for i, err := range me.errors {
		if limit <= 0 {
			remaining := len(me.errors) - i
			builder.WriteString("  ... and ")
			builder.WriteString(strconv.Itoa(remaining))
			if remaining == 1 {
				builder.WriteString(" more error")
			} else {
				builder.WriteString(" more errors")
			}
			break
		}
		limit--

		builder.WriteString("  #")
		for range digits - fmthelper.DigitsInNumber(i) {
			builder.WriteByte('0')
		}
		builder.WriteInt(int64(i))
		builder.WriteString(": ")
		builder.WriteStringIndent(err.Error(), "  ")

		if i < len(me.errors)-1 {
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}

// Message returns the single, unformatted message of this error.
func (me *Err) Message() string {
	return me.Error()
}

// Format implements the fmt.Formatter interface. Use fmt.Sprintf("%v", err) to
// get a string representation of the error without an stack trace and
// fmt.Sprintf("%+v", err) with a stack trace included.
func (me *Err) Format(s fmt.State, verb rune) {
	var str string
	if verb == 'v' && s.Flag('+') {
		str = bruh.String(me)
	} else {
		str = bruh.Message(me)
	}
	_, _ = io.WriteString(s, str)
}

// ErrorOrNil returns nil if the [Err] is nil or if it contains no errors.
// Otherwise it returns the [Err] itself.
func (me *Err) ErrorOrNil() MultiErrorer {
	if me.IsNil() {
		return nil
	}
	return me
}

// SingleOrNil is similar to [ErrorOrNil] but returns the single error if the
// [Err] contains exactly one error. If the [Err] is nil or contains no errors,
// nil is returned. If it contains more than one error, the [Err] itself is
// returned.
func (me *Err) SingleOrNil() error {
	if me.IsNil() {
		return nil
	}
	if len(me.errors) == 1 {
		return me.errors[0]
	}
	return me
}

// Unwrap returns nil if the [Err] is nil or if it contains no errors. If it
// contains exactly one error, that error is returned. If it contains more than
// one error, a chain of the errors is returned.
func (me *Err) Unwrap() error {
	if me.IsNil() {
		return nil
	}
	switch me.unwrapBehavior {
	case UnwrapFirst:
		return me.errors[0]
	case UnwrapLast:
		return me.errors[len(me.errors)-1]
	case UnwrapNone:
		fallthrough
	default:
		return nil
	}
}

// IsNil returns true if the [Err] is nil or if it contains no errors.
func (me *Err) IsNil() bool {
	return me == nil || len(me.errors) == 0
}

// Add adds the given errors to the [Err].
func (me *Err) Add(err ...error) {
	count := 0
	for _, e := range err {
		if e != nil {
			count++
		}
	}
	if count == 0 {
		return
	}
	me.Grow(len(me.errors) + count)
	for _, e := range err {
		if e == nil || me.filter != nil && !me.filter(e) {
			continue
		}
		me.errors = append(me.errors, e)
	}
}

// Merge merges the given multi errors into the current error.
func (me *Err) Merge(errs ...MultiErrorer) {
	if errs == nil {
		return
	}
	growGuess := len(me.errors)
	for _, e := range errs {
		if e == nil {
			continue
		}
		growGuess += len(e.Errors())
	}
	if growGuess == 0 {
		return
	}
	me.Grow(growGuess)
	for _, err := range errs {
		if err == nil {
			continue
		}
		me.Add(err.Errors()...)
	}
}

// As implements [errors.As] and attempts to map the target to any of the errors
// contained in the [Err]. If any of the errors contained in the [Err] can be
// mapped to the target, true is returned and the target is set to the mapped
// error. Otherwise, false is returned and the target is not modified.
//
// It respects the UnwrapBehavior option. If UnwrapFirst is set, the errors are
// checked in the order they were added. If UnwrapLast is set, the errors are
// checked in reverse order. If UnwrapNone is set, the target is not mapped to
// any of the contained errors and false is returned.
func (me *Err) As(target any) bool {
	switch me.unwrapBehavior {
	case UnwrapFirst:
		for _, err := range me.errors {
			if errors.As(err, target) {
				return true
			}
		}
	case UnwrapLast:
		for i := len(me.errors) - 1; i >= 0; i-- {
			if errors.As(me.errors[i], target) {
				return true
			}
		}
	case UnwrapNone:
		fallthrough
	default:
	}
	return false
}

// Is implements [errors.Is] and compares the target to any of the errors
// contained in the [Err]. If any of the errors contained in the [Err] matches
// the target, true is returned. Otherwise, false is returned.
//
// It respects the UnwrapBehavior option. If UnwrapFirst is set, the errors are
// checked in the order they were added. If UnwrapLast is set, the errors are
// checked in reverse order. If UnwrapNone is set, the target is not compared to
// any of the contained errors and false is returned.
func (me *Err) Is(target error) bool {
	switch me.unwrapBehavior {
	case UnwrapFirst:
		for _, err := range me.errors {
			if errors.Is(err, target) {
				return true
			}
		}
	case UnwrapLast:
		for i := len(me.errors) - 1; i >= 0; i-- {
			if errors.Is(me.errors[i], target) {
				return true
			}
		}
	case UnwrapNone:
		fallthrough
	default:
	}
	return false
}

// Context implements [ctxerror.contexter] by combining the contexts of all
// contained errors. If multiple errors have the same context key, the value of
// the last error with that key is used.
func (me *Err) Context() map[string]map[string]any {
	context := make(map[string]map[string]any, len(me.errors))
	for _, err := range me.errors {
		subCtx := ctxerror.GetContext(err)
		maps.Copy(context, subCtx)
	}
	return context
}

// AppendContext implements [ctxerror.contextAppender] by combining the contexts of all
// contained errors. If multiple errors have the same context key, the value of
// the last error with that key is used.
func (me *Err) AppendContext(context ctxerror.Context) {
	for _, err := range me.errors {
		subCtx := ctxerror.GetContext(err)
		for k, v := range subCtx {
			if _, ok := context[k]; !ok {
				context[k] = v
				continue
			}
			maps.Copy(context[k], v)
		}
	}
}

// Tags implements [ctxerror.contexter] by combining the tags of all contained
// errors. If multiple errors have the same tag key, the value of the last error
// with that key is used.
func (me *Err) Tags() map[string]string {
	tags := make(map[string]string, len(me.errors))
	for _, err := range me.errors {
		subTags := ctxerror.GetTags(err)
		maps.Copy(tags, subTags)
	}
	return tags
}

// AppendTags implements [ctxerror.tagsAppender] by combining the tags of all contained
// errors. If multiple errors have the same tag key, the value of the last error
// with that key is used.
func (me *Err) AppendTags(tags ctxerror.Tags) {
	for _, err := range me.errors {
		subTags := ctxerror.GetTags(err)
		maps.Copy(tags, subTags)
	}
}
