// Package multierror provides an error type that can hold multiple errors.
package multierror

import (
	"fmt"

	"github.com/aisbergg/go-bruh/internal/stringbuilder"
	"github.com/aisbergg/go-bruh/internal/util"
	"github.com/aisbergg/go-bruh/pkg/bruh"
)

// Err is an error that can hold multiple errors.
type Err struct {
	bruh.Err
	errors []error
}

// New creates a new [Err] with the given message and the given errors. You can
// add more errors later by calling [Add].
func New(msg string, errs ...error) *Err {
	return &Err{
		Err:    *bruh.NewSkip(1, msg),
		errors: errs,
	}
}

// Errorf creates a new [Err] with a formatted message.
func Errorf(format string, args ...any) *Err {
	return New(fmt.Sprintf(format, args...))
}

// Wrap creates a new [Err] and adds the wrapped error to the list of errors.
func Wrap(err error, msg string) *Err {
	return New(msg, err)
}

// Wrapf creates a new [Err] with a formatted message and adds the wrapped error
// to the list of errors.
func Wrapf(err error, format string, args ...any) *Err {
	return New(fmt.Sprintf(format, args...), err)
}

// Errors returns the errors that are stored in the [Err].
func (me *Err) Errors() []error {
	return me.errors
}

func (me *Err) Error() string {
	if me.IsNil() {
		return ""
	}
	message := me.Message()
	builder := stringbuilder.New([]byte{})
	guessCap := len(message) + len(me.errors)*180
	builder.Grow(guessCap)
	if message != "" {
		builder.WriteString(message)
		builder.WriteByte('\n')
	}
	digits := util.DigitsInNumber(len(me.errors))
	for i, err := range me.errors {
		builder.WriteString("Error ")
		for range digits - util.DigitsInNumber(i) {
			builder.WriteByte('0')
		}
		builder.WriteInt(int64(i))
		builder.WriteString(": ")
		builder.WriteString(err.Error())

		if i < len(me.errors)-1 {
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}

// ErrorOrNil returns nil if the [Err] is nil or if it contains no errors.
// Otherwise it returns the [Err] itself.
func (me *Err) ErrorOrNil() error {
	if me.IsNil() {
		return nil
	}
	return me
}

// Unwrap for [Err] always returns nil. You have to check the wrapped
// errors yourself.
func (me *Err) Unwrap() error {
	return nil
}

// IsNil returns true if the [Err] is nil or if it contains no errors.
func (me *Err) IsNil() bool {
	return me == nil || len(me.errors) == 0
}

// Add adds the given errors to the [Err].
func (me *Err) Add(err ...error) {
	me.errors = append(me.errors, err...)
}
