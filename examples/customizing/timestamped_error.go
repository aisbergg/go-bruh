package main

import (
	"fmt"
	"time"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	err := foo()
	switch t := err.(type) {
	case *TimestampedError:
		fmt.Printf("timestamp error:\n  time: %s\n  msg: %s\n", t.Timestamp().Format(time.RFC3339), t)
	case *bruh.TraceableError:
		fmt.Println("bruh error:", t)
	default:
		fmt.Println("other error:", t)
	}
}

// -----------------------------------------------------------------------------
// The custom error
// -----------------------------------------------------------------------------

// TimestampedError is a custom error that contains a timestamp.
type TimestampedError struct {
	bruh.TraceableError
	timestamp time.Time
}

// TErrorf creates a new TimestampedError with a formatted message.
func TErrorf(format string, args ...any) error {
	return &TimestampedError{
		// skip is required to skip the current function and thus exclude this
		// function from the stack trace
		TraceableError: *bruh.ErrorfSkip(1, format, args...).(*bruh.TraceableError),
		timestamp:      time.Now(),
	}
}

// TEWrapf wraps an error with a formatted message.
func TEWrapf(err error, format string, args ...any) error {
	// Easiest way to change the final error message is to change it early on.
	// Here we include the timestamp in the message.
	ts := time.Now()
	msg := fmt.Sprintf(format, args...)
	msg = fmt.Sprintf("[%s] %s", ts.Format(time.RFC3339), msg)

	return &TimestampedError{
		TraceableError: *bruh.WrapSkip(err, 1, msg).(*bruh.TraceableError),
		timestamp:      ts,
	}
}

// Timestamp returns the timestamp of the error.
func (te *TimestampedError) Timestamp() time.Time {
	return te.timestamp
}

// -----------------------------------------------------------------------------
// Just some functions that return errors
// -----------------------------------------------------------------------------

func foo() error {
	err := bar()
	if err != nil {
		return TEWrapf(err, "foo: failed to read config file")
	}
	return nil
}

func bar() error {
	err := baz()
	if err != nil {
		return bruh.Wrapf(err, "bar: failed to parse")
	}
	return nil
}

func baz() error {
	// external error
	return fmt.Errorf("oh no")
}
