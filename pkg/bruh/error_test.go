package bruh_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

var (
	globalErr          = bruh.New("global error")
	formattedGlobalErr = bruh.Errorf("%v global error", "formatted")
)

type withMessage struct {
	msg string
}

func (e withMessage) Error() string { return e.msg }
func (e withMessage) Is(target error) bool {
	if err, ok := target.(withMessage); ok {
		return e.msg == err.msg
	}
	return e.msg == target.Error()
}

type withLayer struct {
	err error
	msg string
}

func (e withLayer) Error() string { return e.msg + ": " + e.err.Error() }
func (e withLayer) Unwrap() error { return e.err }
func (e withLayer) Is(target error) bool {
	if err, ok := target.(withLayer); ok {
		return e.msg == err.msg
	}
	return e.msg == target.Error()
}

type withEmptyLayer struct {
	err error
}

func (e withEmptyLayer) Error() string { return e.err.Error() }
func (e withEmptyLayer) Unwrap() error { return e.err }

func setupTestCase(wrapf bool, cause error, input []string) error {
	err := cause
	for _, str := range input {
		if wrapf {
			err = bruh.Wrapf(err, "%v", str)
		} else {
			err = bruh.Wrap(err, str)
		}
	}
	return err
}

func TestErrorCreation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err         error
		message     string
		numStack    int
		shouldBeNil bool
	}{
		{bruh.New("abc"), "abc", 2, false},
		{bruh.NewSkip(1, "def"), "def", 1, false},
		{bruh.Errorf("%s %d", "ghi", 42), "ghi 42", 2, false},
		{bruh.ErrorfSkip(1, "%s %d", "jkl", 42), "jkl 42", 1, false},

		{bruh.Wrap(nil, "abc"), "abc", 0, true},
		{bruh.Wrapf(nil, "abc"), "abc", 0, true},
		{bruh.Wrap(bruh.New("abc"), "abc"), "abc", 3, false},
		{bruh.WrapSkip(bruh.New("abc"), 1, "abc"), "abc", 2, false},
		{bruh.Wrapf(bruh.New("abc"), "abc"), "abc", 3, false},
		{bruh.WrapfSkip(bruh.New("abc"), 1, "abc"), "abc", 2, false},
		{bruh.Wrap(globalErr, "mno"), "mno", 2, false},
		{bruh.WrapSkip(globalErr, 1, "pqr"), "pqr", 1, false},
		{bruh.Wrapf(globalErr, "%s %d", "stu", 42), "stu 42", 2, false},
		{bruh.WrapfSkip(globalErr, 1, "%s %d", "vwx", 42), "vwx 42", 1, false},
	}

	for i, test := range tests {
		if test.shouldBeNil {
			if test.err != nil {
				t.Errorf("test %d: expected nil value", i)
				continue
			}
			continue
		} else if test.err == nil {
			t.Errorf("test %d: expected non nil value", i)
			continue
		}
		err := test.err.(*bruh.Err)
		if err.Message() != test.message {
			t.Errorf("test %d: expected message \"%v\", got \"%v\"", i, test.message, err.Message())
		}
		stack := err.Stack()
		if len(stack) != test.numStack {
			t.Errorf("test %d: expected stack of size %d, got %d", i, test.numStack, len(stack))
		}
	}
}

func TestErrorWrapping(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output string   // expected output
	}{
		"nil root error": {
			cause:  nil,
			input:  []string{"additional context"},
			output: "additional context",
		},
		"standard error wrapping with a global root cause": {
			cause:  globalErr,
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: global error",
		},
		"formatted error wrapping with a global root cause": {
			cause:  formattedGlobalErr,
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: formatted global error",
		},
		"standard error wrapping with a local root cause": {
			cause:  bruh.New("root error"),
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: root error",
		},
		"standard error wrapping with a local root cause (bruh.Errorf)": {
			cause:  bruh.Errorf("%v root error", "formatted"),
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: formatted root error",
		},
		"no error wrapping with a local root cause (bruh.Errorf)": {
			cause:  bruh.Errorf("%v root error", "formatted"),
			output: "formatted root error",
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)
			fmtdErr := bruh.Message(err)
			if err != nil && tc.output != fmtdErr {
				t.Errorf("%v: expected { %v } got { %v }", desc, tc.output, fmtdErr)
			}
		})
	}
}

func TestExternalErrorWrapping(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output []string // expected output
	}{
		"no error wrapping with a third-party root cause (errors.New)": {
			cause: errors.New("external error"),
			output: []string{
				"external error",
			},
		},
		"standard error wrapping with a third-party root cause (errors.New)": {
			cause: errors.New("external error"),
			input: []string{"additional context", "even more context"},
			output: []string{
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
			},
		},
		"wrapping a wrapped third-party root cause (errors.New and fmt.Errorf)": {
			// rather not wrap errors using %w
			cause: fmt.Errorf("additional context: %w", errors.New("external error")),
			input: []string{"even more context"},
			output: []string{
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
			},
		},
		"wrapping a wrapped third-party root cause (multiple layers)": {
			cause: fmt.Errorf(
				"even more context: %w",
				fmt.Errorf("additional context: %w", errors.New("external error")),
			),
			input: []string{"way too much context"},
			output: []string{
				"way too much context: even more context: additional context: external error",
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
			},
		},
		"wrapping a wrapped third-party root cause that contains an empty layer": {
			cause: fmt.Errorf(": %w", errors.New("external error")),
			input: []string{"even more context"},
			output: []string{
				"even more context: : external error",
				": external error",
				"external error",
			},
		},
		"wrapping a wrapped third-party root cause that contains an empty layer without a delimiter": {
			cause: fmt.Errorf("%w", errors.New("external error")),
			input: []string{"even more context"},
			output: []string{
				"even more context: external error",
				"external error",
				"external error",
			},
		},
		"wrapping a pkg/errors style error (contains layers without messages)": {
			cause: &withLayer{ // var to mimic wrapping a pkg/errors style error
				msg: "additional context",
				err: &withEmptyLayer{
					err: &withMessage{
						msg: "external error",
					},
				},
			},
			input: []string{"even more context"},
			output: []string{
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
				"external error",
			},
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)

			// unwrap to make sure external errors are actually wrapped properly
			var inputErr []string
			for err != nil {
				inputErr = append(inputErr, bruh.StringFormat(err, nil))
				err = bruh.Unwrap(err)
			}

			// compare each layer of the actual and expected output
			if len(inputErr) != len(tc.output) {
				t.Fatalf(
					"%v: expected output to have '%v' layers but got '%v': { %#v } got { %#v }",
					desc,
					len(tc.output),
					len(inputErr),
					tc.output,
					inputErr,
				)
			}
			for i := 0; i < len(inputErr); i++ {
				if inputErr[i] != tc.output[i] {
					t.Errorf("%v: expected { %#v } got { %#v }", desc, inputErr[i], tc.output[i])
				}
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output []string // expected output
	}{
		"unwrapping error with internal root cause (bruh.New)": {
			cause: bruh.New("root error"),
			input: []string{"additional context", "even more context"},
			output: []string{
				"even more context: additional context: root error",
				"additional context: root error",
				"root error",
			},
		},
		"unwrapping error with external root cause (errors.New)": {
			cause: errors.New("external error"),
			input: []string{"additional context", "even more context"},
			output: []string{
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
			},
		},
		"unwrapping error with external root cause (custom type)": {
			cause: &withMessage{
				msg: "external error",
			},
			input: []string{"additional context", "even more context"},
			output: []string{
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
			},
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(true, tc.cause, tc.input)
			for _, out := range tc.output {
				if err == nil {
					t.Errorf("%v: unwrapping error returned nil but expected { %v }", desc, out)
				} else if out != err.Error() {
					t.Errorf("%v: expected { %v } got { %v }", desc, out, err)
				}
				err = bruh.Unwrap(err)
			}
		})
	}
}

func TestErrorCause(t *testing.T) {
	t.Parallel()
	globalErr := bruh.New("global error")
	extErr := errors.New("external error")
	customErr := withMessage{
		msg: "external error",
	}

	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output error    // expected output
	}{
		"internal root error": {
			cause:  globalErr,
			input:  []string{"additional context", "even more context"},
			output: globalErr,
		},
		"external error": {
			cause:  extErr,
			input:  []string{"additional context", "even more context"},
			output: extErr,
		},
		"external error (custom type)": {
			cause:  customErr,
			input:  []string{"additional context", "even more context"},
			output: customErr,
		},
		"nil error": {
			cause:  nil,
			output: nil,
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)
			cause := bruh.Cause(err)
			if tc.output != bruh.Cause(err) {
				t.Errorf("%v: expected { %v } got { %v }", desc, tc.output, cause)
			}
		})
	}
}
