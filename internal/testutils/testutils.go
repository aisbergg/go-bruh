// Package testutils contains a minimal set of utils for testing. The package is
// inspired by testify.
package testutils

//revive:disable:exported

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

type tHelper interface {
	Helper()
}

type TestingT interface {
	Name() string
	Errorf(format string, args ...interface{})
	Log(...any)
	FailNow()
}

type Assertions struct {
	t               TestingT
	failImmediately bool
}

func NewAssert(t TestingT) Assertions {
	return Assertions{t, false}
}

func NewRequire(t TestingT) Assertions {
	return Assertions{t, true}
}

func (a Assertions) Equal(exp interface{}, act interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}

	if !equal(exp, act) {
		a.log(fmt.Sprintf("expected '%v', got: '%v'", exp, act), msgAndArgs...)
		return false
	}
	return true
}

func (a Assertions) NotEqual(exp interface{}, act interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}

	if !equal(exp, act) {
		a.log(fmt.Sprintf("expected '%v', got: '%v'", exp, act), msgAndArgs...)
		return false
	}
	return true
}

func (a Assertions) Error(err error, msgAndArgs ...interface{}) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if err == nil {
		a.log("expected an error", msgAndArgs...)
		return false
	}
	return true
}

func (a Assertions) NoError(err error, msgAndArgs ...interface{}) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if err != nil {
		a.log(fmt.Sprintf("expected no error, got: '%v'", err), msgAndArgs...)
		return false
	}
	return true
}

func (a Assertions) EqualError(expErr, actErr error, msgAndArgs ...interface{}) bool {
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

func (a Assertions) False(exp bool, msgAndArgs ...interface{}) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if exp {
		a.log("expected false, got true", msgAndArgs...)
		return false
	}
	return true
}

func (a Assertions) True(exp bool, msgAndArgs ...interface{}) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if !exp {
		a.log("expected true, got false", msgAndArgs...)
		return false
	}
	return true
}

func (a Assertions) IsType(expType interface{}, obj interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if !equal(reflect.TypeOf(obj), reflect.TypeOf(expType)) {
		a.log(fmt.Sprintf("expected object to be of type %v, was %v", reflect.TypeOf(expType), reflect.TypeOf(obj)), msgAndArgs...)
		return false
	}
	return true
}

func (a Assertions) Nil(obj interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if obj != nil {
		a.log(fmt.Sprintf("expected object to be nil, was %v", obj), msgAndArgs...)
		return false
	}
	return true
}

func (a Assertions) NotNil(obj interface{}, msgAndArgs ...interface{}) bool {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	if obj == nil {
		a.log("expected object not to be nil", msgAndArgs...)
		return false
	}
	return true
}

func (a Assertions) Len(obj interface{}, length int, msgAndArgs ...interface{}) bool {
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

func (a Assertions) log(defaultMsg string, msgAndArgs ...interface{}) {
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

func equal(expected, actual interface{}) bool {
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
