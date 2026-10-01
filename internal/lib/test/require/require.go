// Package require provides test assertions that stop test execution on
// failure. All functions take testing.TB as first argument.
//
// Two message forms exist per check: the plain form accepts an optional
// msgAndArgs tail (a bare string is used verbatim, otherwise fmt.Sprintf
// style); the f-form takes a typed format string plus args and is checked by
// go vet's printf analyzer. The message is prepended to the failure
// description.
package require

import (
	"testing"

	"github.com/aisbergg/go-bruh/internal/lib/test/assert"
)

// Equal requires that got and want are deeply equal, failing the test
// immediately otherwise. See assert.Equal for the equality rules.
func Equal[T any](t testing.TB, got, want T, msgAndArgs ...any) {
	t.Helper()
	if !assert.Equal(t, got, want, msgAndArgs...) {
		t.FailNow()
	}
}

// Equalf requires that got and want are deeply equal, failing the test
// immediately otherwise, with a checked format message.
func Equalf[T any](t testing.TB, got, want T, format string, args ...any) {
	t.Helper()
	if !assert.Equalf(t, got, want, format, args...) {
		t.FailNow()
	}
}

// NotEqual requires that got and want are not deeply equal, failing the test
// immediately otherwise.
func NotEqual[T any](t testing.TB, got, want T, msgAndArgs ...any) {
	t.Helper()
	if !assert.NotEqual(t, got, want, msgAndArgs...) {
		t.FailNow()
	}
}

// NotEqualf requires that got and want are not deeply equal, failing the test
// immediately otherwise, with a checked format message.
func NotEqualf[T any](t testing.TB, got, want T, format string, args ...any) {
	t.Helper()
	if !assert.NotEqualf(t, got, want, format, args...) {
		t.FailNow()
	}
}

// EqualUnordered requires that got and want contain the same elements
// regardless of order (multiset semantics, nil equals empty), failing the test
// immediately otherwise.
func EqualUnordered[T any](t testing.TB, got, want []T, msgAndArgs ...any) {
	t.Helper()
	if !assert.EqualUnordered(t, got, want, msgAndArgs...) {
		t.FailNow()
	}
}

// EqualUnorderedf requires that got and want contain the same elements
// regardless of order, failing the test immediately otherwise, with a checked
// format message.
func EqualUnorderedf[T any](t testing.TB, got, want []T, format string, args ...any) {
	t.Helper()
	if !assert.EqualUnorderedf(t, got, want, format, args...) {
		t.FailNow()
	}
}

// Error requires that err is non-nil, failing the test immediately otherwise.
func Error(t testing.TB, err error, msgAndArgs ...any) {
	t.Helper()
	if !assert.Error(t, err, msgAndArgs...) {
		t.FailNow()
	}
}

// Errorf requires that err is non-nil, failing the test immediately
// otherwise, with a checked format message.
func Errorf(t testing.TB, err error, format string, args ...any) {
	t.Helper()
	if !assert.Errorf(t, err, format, args...) {
		t.FailNow()
	}
}

// NoError requires that err is nil, failing the test immediately otherwise.
func NoError(t testing.TB, err error, msgAndArgs ...any) {
	t.Helper()
	if !assert.NoError(t, err, msgAndArgs...) {
		t.FailNow()
	}
}

// NoErrorf requires that err is nil, failing the test immediately otherwise,
// with a checked format message.
func NoErrorf(t testing.TB, err error, format string, args ...any) {
	t.Helper()
	if !assert.NoErrorf(t, err, format, args...) {
		t.FailNow()
	}
}

// EqualError requires that gotErr matches wantErr via errors.Is, failing the
// test immediately otherwise. A nil wantErr requires gotErr to be nil.
func EqualError(t testing.TB, gotErr, wantErr error, msgAndArgs ...any) {
	t.Helper()
	if !assert.EqualError(t, gotErr, wantErr, msgAndArgs...) {
		t.FailNow()
	}
}

// EqualErrorf requires that gotErr matches wantErr via errors.Is, failing the
// test immediately otherwise, with a checked format message.
func EqualErrorf(t testing.TB, gotErr, wantErr error, format string, args ...any) {
	t.Helper()
	if !assert.EqualErrorf(t, gotErr, wantErr, format, args...) {
		t.FailNow()
	}
}

// True requires that value is true, failing the test immediately otherwise.
func True(t testing.TB, value bool, msgAndArgs ...any) {
	t.Helper()
	if !assert.True(t, value, msgAndArgs...) {
		t.FailNow()
	}
}

// Truef requires that value is true, failing the test immediately otherwise,
// with a checked format message.
func Truef(t testing.TB, value bool, format string, args ...any) {
	t.Helper()
	if !assert.Truef(t, value, format, args...) {
		t.FailNow()
	}
}

// False requires that value is false, failing the test immediately otherwise.
func False(t testing.TB, value bool, msgAndArgs ...any) {
	t.Helper()
	if !assert.False(t, value, msgAndArgs...) {
		t.FailNow()
	}
}

// Falsef requires that value is false, failing the test immediately otherwise,
// with a checked format message.
func Falsef(t testing.TB, value bool, format string, args ...any) {
	t.Helper()
	if !assert.Falsef(t, value, format, args...) {
		t.FailNow()
	}
}

// Nil requires that obj is nil (see assert.Nil for typed nil handling),
// failing the test immediately otherwise.
func Nil(t testing.TB, obj any, msgAndArgs ...any) {
	t.Helper()
	if !assert.Nil(t, obj, msgAndArgs...) {
		t.FailNow()
	}
}

// Nilf requires that obj is nil, failing the test immediately otherwise, with
// a checked format message.
func Nilf(t testing.TB, obj any, format string, args ...any) {
	t.Helper()
	if !assert.Nilf(t, obj, format, args...) {
		t.FailNow()
	}
}

// NotNil requires that obj is not nil (see assert.Nil for typed nil
// handling), failing the test immediately otherwise.
func NotNil(t testing.TB, obj any, msgAndArgs ...any) {
	t.Helper()
	if !assert.NotNil(t, obj, msgAndArgs...) {
		t.FailNow()
	}
}

// NotNilf requires that obj is not nil, failing the test immediately
// otherwise, with a checked format message.
func NotNilf(t testing.TB, obj any, format string, args ...any) {
	t.Helper()
	if !assert.NotNilf(t, obj, format, args...) {
		t.FailNow()
	}
}

// Len requires that obj has the given length, failing the test immediately
// otherwise. obj must be an array, chan, map, slice, or string.
func Len(t testing.TB, obj any, length int, msgAndArgs ...any) {
	t.Helper()
	if !assert.Len(t, obj, length, msgAndArgs...) {
		t.FailNow()
	}
}

// Lenf requires that obj has the given length, failing the test immediately
// otherwise, with a checked format message.
func Lenf(t testing.TB, obj any, length int, format string, args ...any) {
	t.Helper()
	if !assert.Lenf(t, obj, length, format, args...) {
		t.FailNow()
	}
}

// Panic requires that f panics, failing the test immediately otherwise.
func Panic(t testing.TB, f func(), msgAndArgs ...any) {
	t.Helper()
	if !assert.Panic(t, f, msgAndArgs...) {
		t.FailNow()
	}
}

// Panicf requires that f panics, failing the test immediately otherwise, with
// a checked format message.
func Panicf(t testing.TB, f func(), format string, args ...any) {
	t.Helper()
	if !assert.Panicf(t, f, format, args...) {
		t.FailNow()
	}
}

// NotPanic requires that f does not panic, failing the test immediately
// otherwise.
func NotPanic(t testing.TB, f func(), msgAndArgs ...any) {
	t.Helper()
	if !assert.NotPanic(t, f, msgAndArgs...) {
		t.FailNow()
	}
}

// NotPanicf requires that f does not panic, failing the test immediately
// otherwise, with a checked format message.
func NotPanicf(t testing.TB, f func(), format string, args ...any) {
	t.Helper()
	if !assert.NotPanicf(t, f, format, args...) {
		t.FailNow()
	}
}
