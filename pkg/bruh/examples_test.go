package bruh_test

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

var (
	ErrUnexpectedEOF          = bruh.New("unexpected EOF")
	FormattedErrUnexpectedEOF = bruh.Errorf("unexpected %v", "EOF")
)

func ExampleMessage() {
	err := io.ErrUnexpectedEOF
	err = bruh.Wrapf(err, "reading file '%v'", "example.json")

	// generated via bruh.Message
	fmt.Println(bruh.Message(err))

	// or via fmt.Printf
	_ = fmt.Sprintf("%v\n", err) // %v: omit stack trace

	// Output:
	// reading file 'example.json': unexpected EOF
}

func ExampleString() {
	err := io.ErrUnexpectedEOF
	err = bruh.Wrapf(err, "reading file '%v'", "example.json")

	// error message with stack trace
	// ignore `clean`, it is just for consistent testing output
	fmt.Println(clean(bruh.String(err)))

	// same result as above when called on a *bruh.Error type
	// BEWARE: if it is not *bruh.Error type, the result will be something else
	_ = fmt.Sprintf("%+v\n", err) // %+v: includes the stack trace

	// Output:
	// reading file 'example.json': unexpected EOF
	//     at github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleString (.../pkg/bruh/examples_test.go:123)
	//     at testing.runExample (.../testing/run_example.go:123)
	//     at testing.runExamples (.../testing/example.go:123)
	//     at testing.(*M).Run (.../testing/testing.go:123)
	//     at main.main (_testmain.go:123)
}

func ExampleStringFormat() {
	err := io.ErrUnexpectedEOF
	err = bruh.Wrapf(err, "reading file '%v'", "example.json")

	// error message generated with custom format
	// ignore `clean`, it is just for consistent testing output
	fmt.Println(clean(bruh.StringFormat(err, bruh.BruhStackedFancyFormatter(false, false, true))))

	// Output:
	// *bruh.Err: reading file 'example.json'
	//     at github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleStringFormat (.../pkg/bruh/examples_test.go:123)
	//     at testing.runExample (.../testing/run_example.go:123)
	//     at testing.runExamples (.../testing/example.go:123)
	//     at testing.(*M).Run (.../testing/testing.go:123)
	//     at main.main (_testmain.go:123)
	// *errors.errorString: unexpected EOF
}

func ExampleWrapSkip() {
	type TimestampedError struct {
		bruh.Err
		timestamp time.Time
	}
	wrapTimestampedError := func(err error, msg string) *TimestampedError {
		// fixed timestamp for consistent testing output
		timestamp := time.Date(2025, 7, 30, 0, 0, 0, 0, time.FixedZone("UTC+2", 2*60*60))
		msg = fmt.Sprintf("[%s] %s", timestamp.Format(time.DateOnly), msg)
		return &TimestampedError{
			Err:       *bruh.WrapSkip(err, 1, msg),
			timestamp: timestamp,
		}
	}
	err := io.ErrUnexpectedEOF
	err = wrapTimestampedError(err, "reading file")

	// ignore `clean`, it is just for consistent testing output
	fmt.Println(clean(bruh.String(err)))
	// Output:
	// [2025-07-30] reading file: unexpected EOF
	//     at github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleWrapSkip (.../pkg/bruh/examples_test.go:123)
	//     at testing.runExample (.../testing/run_example.go:123)
	//     at testing.runExamples (.../testing/example.go:123)
	//     at testing.(*M).Run (.../testing/testing.go:123)
	//     at main.main (_testmain.go:123)
}

func ExampleWrapfSkip() {
	type TimestampedError struct {
		bruh.Err
		timestamp time.Time
	}
	wrapTimestampedError := func(err error, format string, args ...any) *TimestampedError {
		// fixed timestamp for consistent testing output
		timestamp := time.Date(2025, 7, 30, 0, 0, 0, 0, time.FixedZone("UTC+2", 2*60*60))
		format = "[%s] " + format
		return &TimestampedError{
			Err:       *bruh.WrapfSkip(err, 1, format, append([]any{timestamp.Format(time.DateOnly)}, args...)...),
			timestamp: timestamp,
		}
	}
	err := io.ErrUnexpectedEOF
	err = wrapTimestampedError(err, "reading file '%v'", "example.json")

	// ignore `clean`, it is just for consistent testing output
	fmt.Println(clean(bruh.String(err)))
	// Output:
	// [2025-07-30] reading file 'example.json': unexpected EOF
	//     at github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleWrapfSkip (.../pkg/bruh/examples_test.go:123)
	//     at testing.runExample (.../testing/run_example.go:123)
	//     at testing.runExamples (.../testing/example.go:123)
	//     at testing.(*M).Run (.../testing/testing.go:123)
	//     at main.main (_testmain.go:123)
}

// clean replaces the paths in the formatted error output to allow for consistent testing.
func clean(s string) string {
	s = strings.ReplaceAll(s, goTestingPath, "...")
	s = strings.ReplaceAll(s, repoDir, "...")
	s = regexp.MustCompile(`:\d+`).ReplaceAllLiteralString(s, ":123")
	return s
}
