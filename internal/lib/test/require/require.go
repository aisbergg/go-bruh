// Package require provides test assertions that stop test execution on
// failure. All functions take testing.TB as first argument.
//
// Each check takes an optional msg tail (variadic). If given, it is used
// verbatim as context and prepended to the failure description. Formatting
// can be done at the call site with fmt.Sprintf.
package require

import (
	"testing"

	"github.com/aisbergg/go-bruh/internal/lib/test/assert"
)

// Equal asserts that got and want are deeply equal. It stops test execution on
// failure. Use this when subsequent code depends on the values being equal. See
// assert.Equal for the equality rules.
func Equal(t testing.TB, got, want any, msg ...string) {
	t.Helper()
	if !assert.Equal(t, got, want, msg...) {
		t.FailNow()
	}
}

// NotEqual asserts that got and want are not deeply equal. It stops test
// execution on failure. Use this when values must differ for the test to
// continue meaningfully.
func NotEqual(t testing.TB, got, want any, msg ...string) {
	t.Helper()
	if !assert.NotEqual(t, got, want, msg...) {
		t.FailNow()
	}
}

// EqualUnordered asserts that got and want contain the same elements. It ignores
// element order. Duplicate elements must occur equally often in both slices. Nil
// and empty slices are equal. It stops test execution on failure.
func EqualUnordered[T any](t testing.TB, got, want []T, msg ...string) {
	t.Helper()
	if !assert.EqualUnordered(t, got, want, msg...) {
		t.FailNow()
	}
}

// Error asserts that err is not nil. It stops test execution on failure. Use
// this when an error is expected and subsequent code cannot proceed without it.
func Error(t testing.TB, err error, msg ...string) {
	t.Helper()
	if !assert.Error(t, err, msg...) {
		t.FailNow()
	}
}

// NoError asserts that err is nil. It stops test execution on failure. Use this
// when an operation must succeed for the test to continue meaningfully.
func NoError(t testing.TB, err error, msg ...string) {
	t.Helper()
	if !assert.NoError(t, err, msg...) {
		t.FailNow()
	}
}

// EqualError asserts that gotErr matches wantErr. Pass an error to use errors.Is
// comparison. Pass a string to compare error messages directly. It stops test
// execution on failure.
func EqualError(t testing.TB, gotErr error, wantErr any, msg ...string) {
	t.Helper()
	if !assert.EqualError(t, gotErr, wantErr, msg...) {
		t.FailNow()
	}
}

// True asserts that value is true. It stops test execution on failure. Use this
// when a condition must hold for the test to continue meaningfully.
func True(t testing.TB, value bool, msg ...string) {
	t.Helper()
	if !assert.True(t, value, msg...) {
		t.FailNow()
	}
}

// False asserts that value is false. It stops test execution on failure. Use
// this when a condition must not hold for the test to continue meaningfully.
func False(t testing.TB, value bool, msg ...string) {
	t.Helper()
	if !assert.False(t, value, msg...) {
		t.FailNow()
	}
}

// Nil asserts that obj is nil. It handles typed nil values like pointers, maps,
// slices, and interface-wrapped nils. It stops test execution on failure. See
// assert.Nil for details on typed nil handling.
func Nil(t testing.TB, obj any, msg ...string) {
	t.Helper()
	if !assert.Nil(t, obj, msg...) {
		t.FailNow()
	}
}

// NotNil asserts that obj is not nil. It handles typed nil values like pointers,
// maps, slices, and interface-wrapped nils. It stops test execution on failure.
// See assert.Nil for details on typed nil handling.
func NotNil(t testing.TB, obj any, msg ...string) {
	t.Helper()
	if !assert.NotNil(t, obj, msg...) {
		t.FailNow()
	}
}

// Len asserts that obj has the given length. It stops test execution on failure.
// Pass an array, channel, map, slice, or string as obj. Use this when size is
// critical for subsequent operations.
func Len(t testing.TB, obj any, length int, msg ...string) {
	t.Helper()
	if !assert.Len(t, obj, length, msg...) {
		t.FailNow()
	}
}

// Panic asserts that function f panics when called. It stops test execution on
// failure. Use this to verify expected panic behavior in your code.
func Panic(t testing.TB, f func(), msg ...string) {
	t.Helper()
	if !assert.Panic(t, f, msg...) {
		t.FailNow()
	}
}

// NotPanic asserts that function f does not panic when called. It stops test
// execution on failure. Use this to verify normal operation without unexpected
// panics.
func NotPanic(t testing.TB, f func(), msg ...string) {
	t.Helper()
	if !assert.NotPanic(t, f, msg...) {
		t.FailNow()
	}
}

// Contains asserts that slice s contains element elem. It stops test execution
// on failure. Use this when a specific value must be present in a collection.
func Contains[T comparable](t testing.TB, s []T, elem T, msg ...string) {
	t.Helper()
	if !assert.Contains(t, s, elem, msg...) {
		t.FailNow()
	}
}

// NotContains asserts that slice s does not contain element elem. It stops test
// execution on failure. Use this when a value must be absent from a collection.
func NotContains[T comparable](t testing.TB, s []T, elem T, msg ...string) {
	t.Helper()
	if !assert.NotContains(t, s, elem, msg...) {
		t.FailNow()
	}
}

// Empty asserts that obj is empty. It stops test execution on failure. Use this
// when a collection must be empty for subsequent operations to make sense.
func Empty(t testing.TB, obj any, msg ...string) {
	t.Helper()
	if !assert.Empty(t, obj, msg...) {
		t.FailNow()
	}
}

// NotEmpty asserts that obj is not empty. It stops test execution on failure.
// Use this when a collection must have content for subsequent operations to make
// sense.
func NotEmpty(t testing.TB, obj any, msg ...string) {
	t.Helper()
	if !assert.NotEmpty(t, obj, msg...) {
		t.FailNow()
	}
}

// ErrorContains asserts that err is not nil and its message contains substr. It
// stops test execution on failure. Use this to verify specific error details in
// the message text.
func ErrorContains(t testing.TB, err error, substr string, msg ...string) {
	t.Helper()
	if !assert.ErrorContains(t, err, substr, msg...) {
		t.FailNow()
	}
}

// GreaterOrEqual asserts that got is greater than or equal to want. It supports
// all numeric types and strings. It stops test execution on failure. Use this
// when a value must meet a minimum threshold.
func GreaterOrEqual(t testing.TB, got, want any, msg ...string) {
	t.Helper()
	if !assert.GreaterOrEqual(t, got, want, msg...) {
		t.FailNow()
	}
}

// LessOrEqual asserts that got is less than or equal to want. It supports all
// numeric types and strings. It stops test execution on failure. Use this when a
// value must not exceed a maximum threshold.
func LessOrEqual(t testing.TB, got, want any, msg ...string) {
	t.Helper()
	if !assert.LessOrEqual(t, got, want, msg...) {
		t.FailNow()
	}
}

// Zero asserts that obj equals the zero value for its type. It stops test
// execution on failure. Use this to verify default or uninitialized state.
func Zero[T any](t testing.TB, obj T, msg ...string) {
	t.Helper()
	if !assert.Zero(t, obj, msg...) {
		t.FailNow()
	}
}

// NotZero asserts that obj does not equal the zero value for its type. It stops
// test execution on failure. Use this to verify a value has been set or changed
// from default state.
func NotZero[T any](t testing.TB, obj T, msg ...string) {
	t.Helper()
	if !assert.NotZero(t, obj, msg...) {
		t.FailNow()
	}
}

// ErrorIs asserts that err matches target using errors.Is. It stops test
// execution on failure. Use this to verify specific sentinel error values in
// error chains.
func ErrorIs(t testing.TB, err, target error, msg ...string) {
	t.Helper()
	if !assert.ErrorIs(t, err, target, msg...) {
		t.FailNow()
	}
}

// ContainsString asserts that string s contains substring substr. It stops test
// execution on failure. Use this to verify specific text appears in output or
// error messages.
func ContainsString(t testing.TB, s, substr string, msg ...string) {
	t.Helper()
	if !assert.ContainsString(t, s, substr, msg...) {
		t.FailNow()
	}
}

// NotContainsString asserts that string s does not contain substring substr. It
// stops test execution on failure. Use this to verify specific text is absent
// from output or error messages.
func NotContainsString(t testing.TB, s, substr string, msg ...string) {
	t.Helper()
	if !assert.NotContainsString(t, s, substr, msg...) {
		t.FailNow()
	}
}

// NotPanics asserts that function f does not panic when called. It stops test
// execution on failure. Use this to verify normal operation without unexpected
// panics. This is an alias for NotPanic.
func NotPanics(t testing.TB, f func(), msg ...string) {
	t.Helper()
	if !assert.NotPanics(t, f, msg...) {
		t.FailNow()
	}
}

// ErrorAs asserts that err can be unwrapped to match the target type. Pass a
// pointer to the expected error type as target. It stops test execution on
// failure. Use this to verify specific error types in complex error chains.
func ErrorAs(t testing.TB, err error, target any, msg ...string) {
	t.Helper()
	if !assert.ErrorAs(t, err, target, msg...) {
		t.FailNow()
	}
}

// JSONEq asserts that two JSON strings represent equal values. It parses both
// strings before comparing. It stops test execution on failure. Use this to
// verify JSON output matches expected structure and content.
func JSONEq(t testing.TB, expected, actual string, msg ...string) {
	t.Helper()
	if !assert.JSONEq(t, expected, actual, msg...) {
		t.FailNow()
	}
}

// Implements asserts that object satisfies the interface iface. Pass a pointer
// to the interface type as iface. It stops test execution on failure. Use this
// to verify interface compliance at runtime.
func Implements(t testing.TB, iface, object any, msg ...string) {
	t.Helper()
	if !assert.Implements(t, iface, object, msg...) {
		t.FailNow()
	}
}

// IsType asserts that object has the same type as expectedType. It stops test
// execution on failure. Use this to verify exact type matches rather than
// interface satisfaction.
func IsType(t testing.TB, expectedType, object any, msg ...string) {
	t.Helper()
	if !assert.IsType(t, expectedType, object, msg...) {
		t.FailNow()
	}
}
