// Package assert provides test assertions that report failures without
// stopping test execution. All functions take testing.TB as first argument.
//
// Each check takes an optional msg. If given, it is used verbatim as context
// and prepended to the failure description. Formatting can be done at the call
// site with fmt.Sprintf.
package assert

import (
	"testing"
)

// Equal asserts that got and want are deeply equal. It compares byte slices by
// content. Nil and empty slices are not equal. It compares maps without regard
// to key order. It compares slices with regard to element order.
func Equal(t testing.TB, got, want any, msg ...string) bool {
	t.Helper()
	return report(t, checkEqual(got, want), msg...)
}

// NotEqual asserts that got and want are not deeply equal.
func NotEqual(t testing.TB, got, want any, msg ...string) bool {
	t.Helper()
	return report(t, checkNotEqual(got, want), msg...)
}

// EqualUnordered asserts that got and want contain the same elements. It ignores
// element order. Duplicate elements must occur equally often in both slices. Nil
// and empty slices are equal.
func EqualUnordered[T any](t testing.TB, got, want []T, msg ...string) bool {
	t.Helper()
	return report(t, checkEqualUnordered(got, want), msg...)
}

// Error asserts that err is non-nil.
func Error(t testing.TB, err error, msg ...string) bool {
	t.Helper()
	return report(t, checkError(err), msg...)
}

// NoError asserts that err is nil.
func NoError(t testing.TB, err error, msg ...string) bool {
	t.Helper()
	return report(t, checkNoError(err), msg...)
}

// EqualError asserts that gotErr matches wantErr. Pass an error to use
// errors.Is comparison. Pass a string to compare error messages directly.
func EqualError(t testing.TB, gotErr error, wantErr any, msg ...string) bool {
	t.Helper()
	return report(t, checkEqualError(gotErr, wantErr), msg...)
}

// True asserts that value is true.
func True(t testing.TB, value bool, msg ...string) bool {
	t.Helper()
	return report(t, checkTrue(value), msg...)
}

// False asserts that value is false.
func False(t testing.TB, value bool, msg ...string) bool {
	t.Helper()
	return report(t, checkFalse(value), msg...)
}

// Nil asserts that obj is nil. It handles typed nil values like pointers, maps,
// slices, and interface-wrapped nils. Use this instead of a simple == nil check
// for complex types.
func Nil(t testing.TB, obj any, msg ...string) bool {
	t.Helper()
	return report(t, checkNil(obj), msg...)
}

// NotNil asserts that obj is not nil (see Nil for typed nil handling).
func NotNil(t testing.TB, obj any, msg ...string) bool {
	t.Helper()
	return report(t, checkNotNil(obj), msg...)
}

// Len asserts that obj has the given length. obj must be an array, chan,
// map, slice, or string.
func Len(t testing.TB, obj any, length int, msg ...string) bool {
	t.Helper()
	return report(t, checkLen(obj, length), msg...)
}

// Panic asserts that f panics.
func Panic(t testing.TB, f func(), msg ...string) bool {
	t.Helper()
	return report(t, checkPanic(f), msg...)
}

// NotPanic asserts that f does not panic.
func NotPanic(t testing.TB, f func(), msg ...string) bool {
	t.Helper()
	return report(t, checkNotPanic(f), msg...)
}

// Contains asserts that slice s contains element elem. Use this to check if a
// slice has a specific value. For string substring checks, use ContainsString.
func Contains[T comparable](t testing.TB, s []T, elem T, msg ...string) bool {
	t.Helper()
	return report(t, checkContainsSlice(s, elem), msg...)
}

// NotContains asserts that slice s does not contain element elem. Use this to
// check that a value is absent from a slice. For string substring checks, use
// NotContainsString.
func NotContains[T comparable](t testing.TB, s []T, elem T, msg ...string) bool {
	t.Helper()
	return report(t, checkNotContainsSlice(s, elem), msg...)
}

// Empty asserts that obj is empty (nil, zero-length slice/map/string).
func Empty(t testing.TB, obj any, msg ...string) bool {
	t.Helper()
	return report(t, checkEmpty(obj), msg...)
}

// NotEmpty asserts that obj is not empty.
func NotEmpty(t testing.TB, obj any, msg ...string) bool {
	t.Helper()
	return report(t, checkNotEmpty(obj), msg...)
}

// ErrorContains asserts that err is non-nil and its message contains substr.
func ErrorContains(t testing.TB, err error, substr string, msg ...string) bool {
	t.Helper()
	return report(t, checkErrorContains(err, substr), msg...)
}

// GreaterOrEqual asserts that got is greater than or equal to want. It supports
// all numeric types and strings. Use this for value comparison in tests.
func GreaterOrEqual(t testing.TB, got, want any, msg ...string) bool {
	t.Helper()
	return report(t, checkGreaterOrEqual(got, want), msg...)
}

// LessOrEqual asserts that got is less than or equal to want. It supports all
// numeric types and strings. Use this for value comparison in tests.
func LessOrEqual(t testing.TB, got, want any, msg ...string) bool {
	t.Helper()
	return report(t, checkLessOrEqual(got, want), msg...)
}

// Zero asserts that obj is the zero value for its type.
func Zero[T any](t testing.TB, obj T, msg ...string) bool {
	t.Helper()
	var zero T
	return report(t, checkEqual(obj, zero), msg...)
}

// NotZero asserts that obj is not the zero value for its type.
func NotZero[T any](t testing.TB, obj T, msg ...string) bool {
	t.Helper()
	var zero T
	return report(t, checkNotEqual(obj, zero), msg...)
}

// ErrorIs asserts that err matches target via errors.Is.
func ErrorIs(t testing.TB, err, target error, msg ...string) bool {
	t.Helper()
	return report(t, checkErrorIs(err, target), msg...)
}

// ContainsString asserts that s contains substr.
func ContainsString(t testing.TB, s, substr string, msg ...string) bool {
	t.Helper()
	return report(t, checkContainsString(s, substr), msg...)
}

// NotContainsString asserts that s does not contain substr.
func NotContainsString(t testing.TB, s, substr string, msg ...string) bool {
	t.Helper()
	return report(t, checkNotContainsString(s, substr), msg...)
}

// NotPanics asserts that f does not panic.
func NotPanics(t testing.TB, f func(), msg ...string) bool {
	t.Helper()
	return report(t, checkNotPanics(f), msg...)
}

// ErrorAs asserts that err can be unwrapped to match the target type. Pass a
// pointer to the expected error type as target. It uses errors.As internally.
func ErrorAs(t testing.TB, err error, target any, msg ...string) bool {
	t.Helper()
	return report(t, checkErrorAs(err, target), msg...)
}

// JSONEq asserts that two JSON strings are equal (after parsing).
func JSONEq(t testing.TB, expected, actual string, msg ...string) bool {
	t.Helper()
	return report(t, checkJSONEq(expected, actual), msg...)
}

// Implements asserts that object satisfies the interface iface. Pass a pointer
// to the interface type as iface, for example (*MyInterface)(nil). Use this to
// verify interface compliance in tests.
func Implements(t testing.TB, iface, object any, msg ...string) bool {
	t.Helper()
	return report(t, checkImplements(object, iface), msg...)
}

// IsType asserts that object is the same type as expectedType.
func IsType(t testing.TB, expectedType, object any, msg ...string) bool {
	t.Helper()
	return report(t, checkIsType(expectedType, object), msg...)
}

// report logs the failure message via t.Errorf. An empty defaultMsg means the
// check passed. report returns whether it did.
func report(t testing.TB, defaultMsg string, msg ...string) bool {
	t.Helper()
	if defaultMsg == "" {
		return true
	}
	if len(msg) > 0 && msg[0] != "" {
		t.Errorf("%s\n%s", msg[0], defaultMsg)
	} else {
		t.Errorf("%s", defaultMsg)
	}
	return false
}
