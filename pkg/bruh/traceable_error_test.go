package bruh

import (
	"errors"
	"fmt"
	"testing"
)

var (
	globalErr          = New("global error")
	formattedGlobalErr = Errorf("%v global error", "formatted")
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

// type customTraceableError struct {
// 	TraceableError
// 	code int
// }

// func newCustomTraceableError(msg string, code int) error {
// 	return &customTraceableError{
// 		TraceableError: *NewSkip(1, msg).(*TraceableError),
// 		code:           code,
// 	}
// }
// func (e *customTraceableError) Code() int { return e.code }

// type customContextableError struct {
// 	ContextableError
// 	code int
// }

// func newCustomContextableError(msg string, code int) error {
// 	err := &customContextableError{
// 		ContextableError: *CENewSkip(1, msg).(*ContextableError),
// 		code:             code,
// 	}
// 	err.context["code"] = code
// 	return err
// }
// func (e *customContextableError) Code() int { return e.code }

func setupTestCase(wrapf bool, cause error, input []string) error {
	err := cause
	for _, str := range input {
		if wrapf {
			err = Wrapf(err, "%v", str)
		} else {
			err = Wrap(err, str)
		}
	}
	return err
}

func TestErrorCreation(t *testing.T) {
	tests := []struct {
		err      error
		message  string
		numStack int
	}{
		{New("abc"), "abc", 2},
		{NewSkip(1, "def"), "def", 1},
		{Errorf("%s %d", "ghi", 42), "ghi 42", 2},
		{ErrorfSkip(1, "%s %d", "jkl", 42), "jkl 42", 1},

		{Wrap(nil, "abc"), "abc", 2},
		{WrapSkip(nil, 1, "abc"), "abc", 1},
		{Wrapf(nil, "abc"), "abc", 2},
		{WrapfSkip(nil, 1, "abc"), "abc", 1},
		{Wrap(globalErr, "mno"), "mno", 2},
		{WrapSkip(globalErr, 1, "pqr"), "pqr", 1},
		{Wrapf(globalErr, "%s %d", "stu", 42), "stu 42", 2},
		{WrapfSkip(globalErr, 1, "%s %d", "vwx", 42), "vwx 42", 1},
	}

	for i, test := range tests {
		if test.err == nil {
			t.Errorf("test %d: expected non nil value", i)
			continue
		}
		err := test.err.(*TraceableError)
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
			cause:  New("root error"),
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: root error",
		},
		"standard error wrapping with a local root cause (Errorf)": {
			cause:  Errorf("%v root error", "formatted"),
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: formatted root error",
		},
		"no error wrapping with a local root cause (Errorf)": {
			cause:  Errorf("%v root error", "formatted"),
			output: "formatted root error",
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)
			fmtdErr := ToString(err, false)
			if err != nil && tc.output != fmtdErr {
				t.Errorf("%v: expected { %v } got { %v }", desc, tc.output, fmtdErr)
			}
		})
	}
}

func TestExternalErrorWrapping(t *testing.T) {
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
			cause: fmt.Errorf("even more context: %w", fmt.Errorf("additional context: %w", errors.New("external error"))),
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
				inputErr = append(inputErr, ToCustomString(err, nil))
				err = Unwrap(err)
			}

			// compare each layer of the actual and expected output
			if len(inputErr) != len(tc.output) {
				t.Fatalf("%v: expected output to have '%v' layers but got '%v': { %#v } got { %#v }", desc, len(tc.output), len(inputErr), tc.output, inputErr)
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
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output []string // expected output
	}{
		"unwrapping error with internal root cause (New)": {
			cause: New("root error"),
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
				err = Unwrap(err)
			}
		})
	}
}

func TestErrorCause(t *testing.T) {
	globalErr := New("global error")
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
			cause := Cause(err)
			if tc.output != Cause(err) {
				t.Errorf("%v: expected { %v } got { %v }", desc, tc.output, cause)
			}
		})
	}
}

func TestErrorFormatting(t *testing.T) {
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output string   // expected output
	}{
		"standard error wrapping with internal root cause (New)": {
			cause:  New("root error"),
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: root error",
		},
		"standard error wrapping with external root cause (errors.New)": {
			cause:  errors.New("external error"),
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: external error",
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)
			if err != nil && tc.cause == nil {
				t.Errorf("%v: wrapping nil errors should return nil but got { %v }", desc, err)
			} else if err != nil && tc.output != err.Error() {
				t.Errorf("%v: expected { %v } got { %v }", desc, tc.output, err)
			}

			_ = fmt.Sprintf("error formatting results (%v):\n", desc)
			_ = fmt.Sprintf("%v\n", err)
			_ = fmt.Sprintf("%+v", err)
		})
	}
}

// func getFrames(pc []uintptr) []StackFrame {
// 	var stackFrames []StackFrame
// 	if len(pc) == 0 {
// 		return stackFrames
// 	}

// 	frames := runtime.CallersFrames(pc)
// 	for {
// 		frame, more := frames.Next()
// 		i := strings.LastIndex(frame.Function, "/")
// 		name := frame.Function[i+1:]
// 		stackFrames = append(stackFrames, StackFrame{
// 			Name: name,
// 			File: frame.File,
// 			Line: frame.Line,
// 		})
// 		if !more {
// 			break
// 		}
// 	}

// 	return stackFrames
// }

// func getFrameFromLink(link UnpackedElement) Stack {
// 	var stackFrames []StackFrame
// 	stackFrames = append(stackFrames, link.Frame)
// 	return Stack(stackFrames)
// }

// func TestStackFrames(t *testing.T) {
// 	tests := map[string]struct {
// 		cause     error    // root error
// 		input     []string // input for error wrapping
// 		isWrapErr bool     // flag for wrap error
// 	}{
// 		"root error": {
// 			cause:     New("root error"),
// 			isWrapErr: false,
// 		},
// 		"wrapped error": {
// 			cause:     New("root error"),
// 			input:     []string{"additional context", "even more context"},
// 			isWrapErr: true,
// 		},
// 		"external error": {
// 			cause:     errors.New("external error"),
// 			isWrapErr: false,
// 		},
// 		"wrapped external error": {
// 			cause:     errors.New("external error"),
// 			input:     []string{"additional context", "even more context"},
// 			isWrapErr: true,
// 		},
// 		"global root error": {
// 			cause:     globalErr,
// 			isWrapErr: false,
// 		},
// 		"wrapped error from global root error": {
// 			cause:     globalErr,
// 			input:     []string{"additional context", "even more context"},
// 			isWrapErr: true,
// 		},
// 		"nil error": {
// 			cause:     nil,
// 			isWrapErr: false,
// 		},
// 	}

// 	for desc, tc := range tests {
// 		t.Run(desc, func(t *testing.T) {
// 			err := setupTestCase(false, tc.cause, tc.input)
// 			uErr := Unpack(err)
// 			sFrames := Stack(getFrames(StackFrames(err)))
// 			if !tc.isWrapErr && !reflect.DeepEqual(uErr.ErrRoot.Stack, sFrames) {
// 				t.Errorf("%v: expected { %v } got { %v }", desc, uErr.ErrRoot.Stack, sFrames)
// 			}
// 			if tc.isWrapErr && !reflect.DeepEqual(getFrameFromLink(uErr.ErrChain[0]), sFrames) {
// 				t.Errorf("%v: expected { %v } got { %v }", desc, getFrameFromLink(uErr.ErrChain[0]), sFrames)
// 			}
// 		})
// 	}
// }
