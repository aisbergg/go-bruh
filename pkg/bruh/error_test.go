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
	nilError := func(err *bruh.Err) error {
		if err == nil {
			return nil
		}
		return err
	}
	assertCreation := func(name string, inputErr error, message string, numStack int, shouldBeNil bool) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			if shouldBeNil {
				if inputErr != nil {
					t.Fatalf("expected nil value")
				}
				return
			}
			if inputErr == nil {
				t.Fatalf("expected non nil value")
			}

			err := inputErr.(*bruh.Err)
			if err.Message() != message {
				t.Errorf("expected message \"%v\", got \"%v\"", message, err.Message())
			}
			stack := err.Stack()
			if len(stack) != numStack {
				t.Errorf("expected stack of size %d, got %d", numStack, len(stack))
			}
		})
	}

	assertCreation("New", bruh.New("abc"), "abc", 2, false)
	assertCreation("NewSkip", bruh.NewSkip(1, "def"), "def", 1, false)
	assertCreation("Errorf", bruh.Errorf("%s %d", "ghi", 42), "ghi 42", 2, false)
	assertCreation("ErrorfSkip", bruh.ErrorfSkip(1, "%s %d", "jkl", 42), "jkl 42", 1, false)

	assertCreation("WrapNil", bruh.Wrap(nil, "abc"), "", 0, true)
	assertCreation("WrapfNil", bruh.Wrapf(nil, "abc"), "", 0, true)
	assertCreation("WrapSkipNil", nilError(bruh.WrapSkip(nil, 1, "abc")), "", 0, true)
	assertCreation("WrapfSkipNil", nilError(bruh.WrapfSkip(nil, 1, "abc")), "", 0, true)

	assertCreation("WrapBruhError", bruh.Wrap(bruh.New("abc"), "abc"), "abc", 3, false)
	assertCreation("WrapSkipBruhError", bruh.WrapSkip(bruh.New("abc"), 1, "abc"), "abc", 2, false)
	assertCreation("WrapfBruhError", bruh.Wrapf(bruh.New("abc"), "abc"), "abc", 3, false)
	assertCreation("WrapfSkipBruhError", bruh.WrapfSkip(bruh.New("abc"), 1, "abc"), "abc", 2, false)

	assertCreation("WrapGlobalErr", bruh.Wrap(globalErr, "mno"), "mno", 2, false)
	assertCreation("WrapSkipGlobalErr", bruh.WrapSkip(globalErr, 1, "pqr"), "pqr", 1, false)
	assertCreation("WrapfGlobalErr", bruh.Wrapf(globalErr, "%s %d", "stu", 42), "stu 42", 2, false)
	assertCreation("WrapfSkipGlobalErr", bruh.WrapfSkip(globalErr, 1, "%s %d", "vwx", 42), "vwx 42", 1, false)
}

func TestNewFromPanic(t *testing.T) {
	t.Parallel()
	assertNewFromPanic := func(name string, panicValue any, expectedMessage string, expectedCause error, checkCause, expectNil bool) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			err := bruh.NewFromPanic(panicValue)
			if expectNil {
				if err != nil {
					t.Fatalf("expected nil value")
				}
				return
			}
			if err == nil {
				t.Fatalf("expected non nil value")
			}
			if bruh.Message(err) != expectedMessage {
				t.Errorf("expected message { %v } got { %v }", expectedMessage, bruh.Message(err))
			}
			if checkCause && bruh.Cause(err) != expectedCause {
				t.Errorf("expected cause { %v } got { %v }", expectedCause, bruh.Cause(err))
			}
		})
	}

	assertNewFromPanic("NilPanicValue", nil, "", nil, false, true)

	panicErr := errors.New("panic error")
	assertNewFromPanic("ExternalError", panicErr, "panic: panic error", panicErr, true, false)

	t.Run("NonErrorValue", func(t *testing.T) {
		err := bruh.NewFromPanic(123)
		if bruh.Message(err) != "123" {
			t.Errorf("expected message { 123 } got { %v }", bruh.Message(err))
		}
		if bruh.Cause(err) != err {
			t.Errorf("expected cause to be the created error")
		}
	})

	t.Run("BruhErrorPassthrough", func(t *testing.T) {
		original := bruh.New("panic error")
		err := bruh.NewFromPanic(original)
		if err != original {
			t.Fatalf("expected the original bruh error to be returned")
		}
	})
}

func TestErrorWrapping(t *testing.T) {
	t.Parallel()
	assertWrappedMessage := func(name string, cause error, input []string, output string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			err := setupTestCase(false, cause, input)
			fmtdErr := bruh.Message(err)
			if err != nil && output != fmtdErr {
				t.Errorf("expected { %v } got { %v }", output, fmtdErr)
			}
		})
	}

	assertWrappedMessage("NilRootError", nil, []string{"additional context"}, "additional context")
	assertWrappedMessage(
		"StandardGlobalRootCause",
		globalErr,
		[]string{"additional context", "even more context"},
		"even more context: additional context: global error",
	)
	assertWrappedMessage(
		"FormattedGlobalRootCause",
		formattedGlobalErr,
		[]string{"additional context", "even more context"},
		"even more context: additional context: formatted global error",
	)
	assertWrappedMessage(
		"StandardLocalRootCause",
		bruh.New("root error"),
		[]string{"additional context", "even more context"},
		"even more context: additional context: root error",
	)
	assertWrappedMessage(
		"FormattedLocalRootCause",
		bruh.Errorf("%v root error", "formatted"),
		[]string{"additional context", "even more context"},
		"even more context: additional context: formatted root error",
	)
	assertWrappedMessage(
		"NoAdditionalWrapping",
		bruh.Errorf("%v root error", "formatted"),
		nil,
		"formatted root error",
	)
}

func TestExternalErrorWrapping(t *testing.T) {
	t.Parallel()
	assertExternalWrapping := func(name string, cause error, input, output []string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			err := setupTestCase(false, cause, input)

			var inputErr []string
			for err != nil {
				inputErr = append(inputErr, bruh.StringFormat(err, nil))
				err = bruh.Unwrap(err)
			}

			if len(inputErr) != len(output) {
				t.Fatalf(
					"expected output to have '%v' layers but got '%v': { %#v } got { %#v }",
					len(output),
					len(inputErr),
					output,
					inputErr,
				)
			}
			for i := 0; i < len(inputErr); i++ {
				if inputErr[i] != output[i] {
					t.Errorf("expected { %#v } got { %#v }", output[i], inputErr[i])
				}
			}
		})
	}

	assertExternalWrapping("NoWrappingErrorsNew", errors.New("external error"), nil, []string{"external error"})
	assertExternalWrapping(
		"StandardWrappingErrorsNew",
		errors.New("external error"),
		[]string{"additional context", "even more context"},
		[]string{
			"even more context: additional context: external error",
			"additional context: external error",
			"external error",
		},
	)
	assertExternalWrapping(
		"WrappedThirdPartyRootCause",
		fmt.Errorf("additional context: %w", errors.New("external error")),
		[]string{"even more context"},
		[]string{
			"even more context: additional context: external error",
			"additional context: external error",
			"external error",
		},
	)
	assertExternalWrapping(
		"WrappedThirdPartyMultipleLayers",
		fmt.Errorf(
			"even more context: %w",
			fmt.Errorf("additional context: %w", errors.New("external error")),
		),
		[]string{"way too much context"},
		[]string{
			"way too much context: even more context: additional context: external error",
			"even more context: additional context: external error",
			"additional context: external error",
			"external error",
		},
	)
	assertExternalWrapping(
		"WrappedThirdPartyWithEmptyLayer",
		fmt.Errorf(": %w", errors.New("external error")),
		[]string{"even more context"},
		[]string{
			"even more context: : external error",
			": external error",
			"external error",
		},
	)
	assertExternalWrapping(
		"WrappedThirdPartyWithoutDelimiter",
		fmt.Errorf("%w", errors.New("external error")),
		[]string{"even more context"},
		[]string{
			"even more context: external error",
			"external error",
			"external error",
		},
	)
	assertExternalWrapping(
		"PkgErrorsStyleWithoutMessages",
		&withLayer{
			msg: "additional context",
			err: &withEmptyLayer{
				err: &withMessage{
					msg: "external error",
				},
			},
		},
		[]string{"even more context"},
		[]string{
			"even more context: additional context: external error",
			"additional context: external error",
			"external error",
			"external error",
		},
	)
}

func TestErrorUnwrap(t *testing.T) {
	t.Parallel()
	assertUnwrap := func(name string, cause error, input, output []string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			err := setupTestCase(true, cause, input)
			for _, out := range output {
				if err == nil {
					t.Errorf("unwrapping error returned nil but expected { %v }", out)
				} else if out != err.Error() {
					t.Errorf("expected { %v } got { %v }", out, err)
				}
				err = bruh.Unwrap(err)
			}
		})
	}

	assertUnwrap(
		"InternalRootCause",
		bruh.New("root error"),
		[]string{"additional context", "even more context"},
		[]string{
			"even more context: additional context: root error",
			"additional context: root error",
			"root error",
		},
	)
	assertUnwrap(
		"ExternalRootCauseErrorsNew",
		errors.New("external error"),
		[]string{"additional context", "even more context"},
		[]string{
			"even more context: additional context: external error",
			"additional context: external error",
			"external error",
		},
	)
	assertUnwrap(
		"ExternalRootCauseCustomType",
		&withMessage{msg: "external error"},
		[]string{"additional context", "even more context"},
		[]string{
			"even more context: additional context: external error",
			"additional context: external error",
			"external error",
		},
	)
}

func TestErrorCause(t *testing.T) {
	t.Parallel()
	globalErr := bruh.New("global error")
	extErr := errors.New("external error")
	customErr := withMessage{
		msg: "external error",
	}
	assertCause := func(name string, cause error, input []string, output error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			err := setupTestCase(false, cause, input)
			actual := bruh.Cause(err)
			if output != actual {
				t.Errorf("expected { %v } got { %v }", output, actual)
			}
		})
	}

	assertCause("InternalRootError", globalErr, []string{"additional context", "even more context"}, globalErr)
	assertCause("ExternalError", extErr, []string{"additional context", "even more context"}, extErr)
	assertCause("ExternalErrorCustomType", customErr, []string{"additional context", "even more context"}, customErr)
	assertCause("NilError", nil, nil, nil)
}

func TestErrMethodAliases(t *testing.T) {
	t.Parallel()
	t.Run("Cause", func(t *testing.T) {
		root := errors.New("root error")
		err := bruh.Wrap(root, "additional context").(*bruh.Err)
		if err.Cause() != root {
			t.Errorf("expected cause { %v } got { %v }", root, err.Cause())
		}
	})

	t.Run("StackFrames", func(t *testing.T) {
		err := bruh.Wrap(bruh.New("root error"), "additional context").(*bruh.Err)
		stack := err.Stack()
		stackFrames := err.StackFrames()
		if len(stackFrames) != len(stack) {
			t.Fatalf("expected stack frame size %d, got %d", len(stack), len(stackFrames))
		}
		if stackFrames.String() != stack.String() {
			t.Errorf("expected StackFrames to match Stack")
		}
	})

	t.Run("Callers", func(t *testing.T) {
		err := bruh.New("root error").(*bruh.Err)
		callers := err.Callers()
		if len(callers) == 0 {
			t.Fatalf("expected callers to be recorded")
		}
		if callers[0] == 0 {
			t.Errorf("expected caller program counter to be set")
		}
		for i, caller := range callers {
			if caller == 0 {
				t.Errorf("expected non-zero caller program counter at index %d", i)
			}
		}
		if len(callers) < len(err.Stack()) {
			t.Errorf(
				"expected callers size to be at least stack size, got callers=%d stack=%d",
				len(callers),
				len(err.Stack()),
			)
		}
	})
}

func TestAs(t *testing.T) {
	t.Parallel()
	t.Run("FindsWrappedTarget", func(t *testing.T) {
		err := bruh.Wrap(&withMessage{msg: "external error"}, "additional context")
		var target *withMessage
		if !bruh.As(err, &target) {
			t.Fatalf("expected As to find the wrapped target")
		}
		if target == nil || target.msg != "external error" {
			t.Errorf("expected target message { external error } got { %v }", target)
		}
	})
}

func TestIs(t *testing.T) {
	t.Parallel()
	t.Run("MatchesWrappedTarget", func(t *testing.T) {
		err := bruh.Wrap(&withMessage{msg: "external error"}, "additional context")
		if !bruh.Is(err, withMessage{msg: "external error"}) {
			t.Fatalf("expected Is to match the wrapped target")
		}
	})
}
