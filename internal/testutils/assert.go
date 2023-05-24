// Package testutils contains a minimal set of utils for testing. The package is
// inspired by testify.
package testutils

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

type tHelper interface {
	Helper()
}

// TestingT is an interface wrapper around *testing.T.
type TestingT interface {
	Name() string
	Errorf(format string, args ...any)
	Log(...any)
	FailNow()
}

// Assertions is a collection of assertion methods that can be used to test
// conditions in your tests.
type Assertions struct {
	t               TestingT
	failImmediately bool
}

// NewAssert returns a new Assertions object that will log assertion failures.
func NewAssert(t TestingT) Assertions {
	return Assertions{t, false}
}

// NewRequire returns a new Assertions object that will log assertion failures
// and stop test execution.
func NewRequire(t TestingT) Assertions {
	return Assertions{t, true}
}

// Equal asserts that two objects are equal.
func (a Assertions) Equal(exp, act any, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}

	if !equal(exp, act) {
		a.log(fmt.Sprintf("expected '%v', got: '%v'", exp, act), msgAndArgs...)
		return false
	}
	return true
}

// NotEqual asserts that two objects are not equal.
func (a Assertions) NotEqual(exp, act any, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}

	if !equal(exp, act) {
		a.log(fmt.Sprintf("expected '%v', got: '%v'", exp, act), msgAndArgs...)
		return false
	}
	return true
}

// Error asserts that a function returned an error (i.e. not `nil`).
func (a Assertions) Error(err error, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if err == nil {
		a.log("expected an error", msgAndArgs...)
		return false
	}
	return true
}

// NoError asserts that a function returned no error (i.e. `nil`).
func (a Assertions) NoError(err error, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if err != nil {
		a.log(fmt.Sprintf("expected no error, got: '%v'", err), msgAndArgs...)
		return false
	}
	return true
}

// EqualError asserts that a function returned an error (i.e. not `nil`) and
// that it is equal to the provided error.
func (a Assertions) EqualError(expErr, actErr error, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if expErr == nil {
		return a.NoError(actErr)
	} else if !errors.Is(actErr, expErr) {
		a.log(fmt.Sprintf("expected error '%v', got: '%v'", expErr, actErr), msgAndArgs...)
		return false
	}
	return true
}

// Panic asserts that the code inside the specified PanicTestFunc panics.
func (a Assertions) Panic(f func(), msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	defer func() {
		if r := recover(); r == nil {
			a.log("expected a panic", msgAndArgs...)
		}
	}()
	f()
	return true
}

// NotPanic asserts that the code inside the specified PanicTestFunc does not
// panic.
func (a Assertions) NotPanic(f func(), msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	defer func() {
		if r := recover(); r != nil {
			a.log(fmt.Sprintf("expected no panic, got: '%v'", r), msgAndArgs...)
		}
	}()
	f()
	return true
}

// False asserts that the specified value is false.
func (a Assertions) False(exp bool, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if exp {
		a.log("expected false, got true", msgAndArgs...)
		return false
	}
	return true
}

// True asserts that the specified value is true.
func (a Assertions) True(exp bool, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if !exp {
		a.log("expected true, got false", msgAndArgs...)
		return false
	}
	return true
}

// IsType asserts that the specified object is of the specified type.
func (a Assertions) IsType(expType, obj any, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if !equal(reflect.TypeOf(obj), reflect.TypeOf(expType)) {
		a.log(fmt.Sprintf("expected object to be of type %v, was %v", reflect.TypeOf(expType), reflect.TypeOf(obj)), msgAndArgs...)
		return false
	}
	return true
}

// Nil asserts that the specified object is nil.
func (a Assertions) Nil(obj any, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if obj != nil {
		a.log(fmt.Sprintf("expected object to be nil, was %v", obj), msgAndArgs...)
		return false
	}
	return true
}

// NotNil asserts that the specified object is not nil.
func (a Assertions) NotNil(obj any, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if obj == nil {
		a.log("expected object not to be nil", msgAndArgs...)
		return false
	}
	return true
}

// Len asserts that the specified object has specific length.
func (a Assertions) Len(obj any, length int, msgAndArgs ...any) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	rv := reflect.ValueOf(obj)
	switch rv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		if rv.Len() != length {
			a.log(fmt.Sprintf("expected object to have length %v, was %v", length, rv.Len()), msgAndArgs...)
			return false
		}
	default:
		a.log(fmt.Sprintf("expected object to be of type array, chan, map, slice or string, was %v", rv.Kind()), msgAndArgs...)
		return false
	}
	return true
}

// log is a helper function that formats the message and logs it.
func (a Assertions) log(defaultMsg string, msgAndArgs ...any) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	msg := defaultMsg
	if len(msgAndArgs) > 0 {
		msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	name := a.t.Name()
	if name != "" {
		msg = fmt.Sprintf("[%s] %s", name, msg)
	}
	if a.failImmediately {
		a.t.Log(msg)
		a.t.FailNow()
	} else {
		a.t.Errorf(msg)
	}
}

// equal is a helper function that compares two objects and returns true if they
// are equal.
func equal(expected, actual any) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}
