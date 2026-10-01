// Package assert provides test assertions that report failures without
// stopping test execution. All functions take testing.TB as first argument.
//
// Two message forms exist per check: the plain form accepts an optional
// msgAndArgs tail (a bare string is used verbatim, otherwise fmt.Sprintf
// style); the f-form takes a typed format string plus args and is checked by
// go vet's printf analyzer. The message is prepended to the failure
// description.
package assert

import (
	"testing"
)

// Equal asserts that got and want are deeply equal. Byte slices are compared
// by content; nil and empty slices are not equal. Maps are compared without
// regard to key order; slices are compared with regard to element order.
func Equal[T any](t testing.TB, got, want T, msgAndArgs ...any) bool { //nolint:revive
	t.Helper()
	return report(t, checkEqual(got, want), msgAndArgs...)
}

// Equalf asserts that got and want are deeply equal, with a checked format
// message. See Equal for the equality rules.
func Equalf[T any](t testing.TB, got, want T, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkEqual(got, want), format, args...)
}

// NotEqual asserts that got and want are not deeply equal.
func NotEqual[T any](t testing.TB, got, want T, msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkNotEqual(got, want), msgAndArgs...)
}

// NotEqualf asserts that got and want are not deeply equal, with a checked
// format message.
func NotEqualf[T any](t testing.TB, got, want T, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkNotEqual(got, want), format, args...)
}

// EqualUnordered asserts that got and want contain the same elements
// regardless of order (multiset semantics, nil equals empty).
func EqualUnordered[T any](t testing.TB, got, want []T, msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkEqualUnordered(got, want), msgAndArgs...)
}

// EqualUnorderedf asserts that got and want contain the same elements
// regardless of order, with a checked format message.
func EqualUnorderedf[T any](t testing.TB, got, want []T, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkEqualUnordered(got, want), format, args...)
}

// Error asserts that err is non-nil.
func Error(t testing.TB, err error, msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkError(err), msgAndArgs...)
}

// Errorf asserts that err is non-nil, with a checked format message.
func Errorf(t testing.TB, err error, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkError(err), format, args...)
}

// NoError asserts that err is nil.
func NoError(t testing.TB, err error, msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkNoError(err), msgAndArgs...)
}

// NoErrorf asserts that err is nil, with a checked format message.
func NoErrorf(t testing.TB, err error, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkNoError(err), format, args...)
}

// EqualError asserts that gotErr matches wantErr via errors.Is. A nil
// wantErr requires gotErr to be nil.
func EqualError(t testing.TB, gotErr, wantErr error, msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkEqualError(gotErr, wantErr), msgAndArgs...)
}

// EqualErrorf asserts that gotErr matches wantErr via errors.Is, with a
// checked format message.
func EqualErrorf(t testing.TB, gotErr, wantErr error, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkEqualError(gotErr, wantErr), format, args...)
}

// True asserts that value is true.
func True(t testing.TB, value bool, msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkTrue(value), msgAndArgs...)
}

// Truef asserts that value is true, with a checked format message.
func Truef(t testing.TB, value bool, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkTrue(value), format, args...)
}

// False asserts that value is false.
func False(t testing.TB, value bool, msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkFalse(value), msgAndArgs...)
}

// Falsef asserts that value is false, with a checked format message.
func Falsef(t testing.TB, value bool, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkFalse(value), format, args...)
}

// Nil asserts that obj is nil, including typed nil values such as (*T)(nil),
// map(nil), slice(nil), and interfaces wrapping those nils.
func Nil(t testing.TB, obj any, msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkNil(obj), msgAndArgs...)
}

// Nilf asserts that obj is nil, with a checked format message.
func Nilf(t testing.TB, obj any, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkNil(obj), format, args...)
}

// NotNil asserts that obj is not nil (see Nil for typed nil handling).
func NotNil(t testing.TB, obj any, msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkNotNil(obj), msgAndArgs...)
}

// NotNilf asserts that obj is not nil, with a checked format message.
func NotNilf(t testing.TB, obj any, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkNotNil(obj), format, args...)
}

// Len asserts that obj has the given length. obj must be an array, chan,
// map, slice, or string.
func Len(t testing.TB, obj any, length int, msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkLen(obj, length), msgAndArgs...)
}

// Lenf asserts that obj has the given length, with a checked format message.
func Lenf(t testing.TB, obj any, length int, format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkLen(obj, length), format, args...)
}

// Panic asserts that f panics.
func Panic(t testing.TB, f func(), msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkPanic(f), msgAndArgs...)
}

// Panicf asserts that f panics, with a checked format message.
func Panicf(t testing.TB, f func(), format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkPanic(f), format, args...)
}

// NotPanic asserts that f does not panic.
func NotPanic(t testing.TB, f func(), msgAndArgs ...any) bool {
	t.Helper()
	return report(t, checkNotPanic(f), msgAndArgs...)
}

// NotPanicf asserts that f does not panic, with a checked format message.
func NotPanicf(t testing.TB, f func(), format string, args ...any) bool {
	t.Helper()
	return reportf(t, checkNotPanic(f), format, args...)
}

// report logs the failure message via t.Errorf. An empty defaultMsg means the
// check passed; report returns whether it did.
func report(t testing.TB, defaultMsg string, msgAndArgs ...any) bool {
	t.Helper()
	if defaultMsg == "" {
		return true
	}
	t.Errorf("%s", formatMsgArgs(defaultMsg, msgAndArgs...))
	return false
}

// reportf is report with a typed format message, enabling go vet to check
// format verbs against args at call sites.
func reportf(t testing.TB, defaultMsg, format string, args ...any) bool {
	t.Helper()
	if defaultMsg == "" {
		return true
	}
	t.Errorf("%s", formatMsg(defaultMsg, format, args...))
	return false
}
