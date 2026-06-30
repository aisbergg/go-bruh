package bruh_test

import (
	"errors"
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

func ExampleNew() {
	err := bruh.New("opening file")
	fmt.Println(err)

	// Output:
	// opening file
}

func ExampleNewFromPanic() {
	var err error
	func() {
		defer func() {
			err = bruh.NewFromPanic(recover())
		}()
		panic("something went wrong")
	}()

	fmt.Println(bruh.Message(err))

	// Output:
	// something went wrong
}

func ExampleNewSkip() {
	type CustomError struct {
		bruh.Err
	}
	err := &CustomError{
		Err: *bruh.NewSkip(0, "opening file"),
	}
	fmt.Println(err)

	// Output:
	// opening file
}

func ExampleErrorf() {
	err := bruh.Errorf("opening file %q", "example.json")
	fmt.Println(err)

	// Output:
	// opening file "example.json"
}

func ExampleErrorfSkip() {
	type CustomError struct {
		bruh.Err
	}
	err := &CustomError{
		Err: *bruh.ErrorfSkip(0, "opening file %q", "example.json"),
	}
	fmt.Println(err)

	// Output:
	// opening file "example.json"
}

func ExampleWrap() {
	err := bruh.Wrap(io.ErrUnexpectedEOF, "reading file")
	fmt.Println(err)

	// Output:
	// reading file: unexpected EOF
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

func ExampleWrapf() {
	err := bruh.Wrapf(io.ErrUnexpectedEOF, "reading file %q", "example.json")
	fmt.Println(err)

	// Output:
	// reading file "example.json": unexpected EOF
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

// -----------------------------------------------------------------------------
// Methods of *bruh.Err
// -----------------------------------------------------------------------------

func ExampleErr_Message() {
	err := bruh.Wrap(io.ErrUnexpectedEOF, "reading file").(*bruh.Err)
	fmt.Println(err.Message())

	// Output:
	// reading file
}

func ExampleErr_Error() {
	err := bruh.Wrap(io.ErrUnexpectedEOF, "reading file").(*bruh.Err)
	fmt.Println(err.Error())

	// Output:
	// reading file: unexpected EOF
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

func ExampleErr_Format() {
	err := bruh.Wrap(io.ErrUnexpectedEOF, "reading file").(*bruh.Err)
	fmt.Printf("%v\n", err)

	// Output:
	// reading file: unexpected EOF
}

func ExampleErr_Unwrap() {
	err := bruh.Wrap(io.ErrUnexpectedEOF, "reading file").(*bruh.Err)
	fmt.Println(err.Unwrap() == io.ErrUnexpectedEOF)

	// Output:
	// true
}

func ExampleErr_Cause() {
	err := bruh.Wrap(bruh.Wrap(io.ErrUnexpectedEOF, "inner"), "outer").(*bruh.Err)
	fmt.Println(err.Cause() == io.ErrUnexpectedEOF)

	// Output:
	// true
}

func ExampleErr_Stack() {
	err := bruh.New("opening file").(*bruh.Err)
	// ignore `clean`, it is just for consistent testing output
	fmt.Println(clean(err.Stack().String()))

	// Output:
	// github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleErr_Stack
	// 	.../pkg/bruh/examples_test.go:123 pc=0x12345
	// testing.runExample
	// 	.../testing/run_example.go:123 pc=0x12345
	// testing.runExamples
	// 	.../testing/example.go:123 pc=0x12345
	// testing.(*M).Run
	// 	.../testing/testing.go:123 pc=0x12345
	// main.main
	// 	_testmain.go:123 pc=0x12345
}

func ExampleErr_StackFrames() {
	err := bruh.New("opening file").(*bruh.Err)
	// ignore `clean`, it is just for consistent testing output
	fmt.Println(clean(err.StackFrames().String()))

	// Output:
	// github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleErr_StackFrames
	// 	.../pkg/bruh/examples_test.go:123 pc=0x12345
	// testing.runExample
	// 	.../testing/run_example.go:123 pc=0x12345
	// testing.runExamples
	// 	.../testing/example.go:123 pc=0x12345
	// testing.(*M).Run
	// 	.../testing/testing.go:123 pc=0x12345
	// main.main
	// 	_testmain.go:123 pc=0x12345
}

func ExampleUnwrap() {
	err := bruh.Wrap(io.ErrUnexpectedEOF, "reading file")
	fmt.Println(bruh.Unwrap(err) == io.ErrUnexpectedEOF)

	// Output:
	// true
}

func ExampleCause() {
	err := bruh.Wrap(bruh.Wrap(io.ErrUnexpectedEOF, "inner"), "outer")
	fmt.Println(bruh.Cause(err) == io.ErrUnexpectedEOF)

	// Output:
	// true
}

func ExampleAs() {
	err := bruh.Wrap(io.ErrUnexpectedEOF, "reading file")

	var target *bruh.Err
	if errors.As(err, &target) {
		fmt.Printf("error is of type *bruh.Err with message: %s\n", target.Message())
	}

	// Output:
	// error is of type *bruh.Err with message: reading file
}

func ExampleIs() {
	sentinel := errors.New("sentinel")
	err := bruh.Wrap(sentinel, "reading file")
	fmt.Println(bruh.Is(err, sentinel))

	// Output:
	// true
}

// clean replaces the paths in the formatted error output to allow for consistent testing.
func clean(s string) string {
	s = strings.ReplaceAll(s, goTestingPath, "...")
	s = strings.ReplaceAll(s, repoDir, "...")
	s = regexp.MustCompile(`:\d+`).ReplaceAllLiteralString(s, ":123")
	// pc=0x12345 is a memory address that can change between runs, so we replace it with a fixed value
	s = regexp.MustCompile(`0x[0-9a-fA-F]+`).ReplaceAllLiteralString(s, "0x12345")
	return s
}
