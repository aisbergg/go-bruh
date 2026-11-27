// Unexported assertion checks. Each check returns a failure message, or the
// empty string on success.

package assert

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// checkEqual returns a failure message if got and want are not deeply equal. It
// compares byte slices by content. Nil and empty slices are not equal. It
// compares maps without regard to key order. It compares slices with regard to
// element order.
func checkEqual(got, want any) string {
	if deepEqual(got, want) {
		return ""
	}
	return fmt.Sprintf("got:  %v\nwant: %v", got, want)
}

// checkNotEqual returns a failure message if got and want are deeply equal.
func checkNotEqual(got, want any) string {
	if deepEqual(got, want) {
		return fmt.Sprintf("got equal to want: %v", got)
	}
	return ""
}

// checkEqualUnordered returns a failure message if got and want do not contain
// the same elements. It ignores element order. Duplicate elements must occur
// equally often in both slices. It compares elements deeply. Nil and empty
// slices are equal.
func checkEqualUnordered[T any](got, want []T) string {
	matched := make([]bool, len(want))
	var extra []T
	for _, g := range got {
		found := false
		for i, w := range want {
			if !matched[i] && deepEqual(g, w) {
				matched[i] = true
				found = true
				break
			}
		}
		if !found {
			extra = append(extra, g)
		}
	}
	var missing []T
	for i, w := range want {
		if !matched[i] {
			missing = append(missing, w)
		}
	}
	if len(extra) == 0 && len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf("slices differ (ignoring order):\nextra:   %v\nmissing: %v", extra, missing)
}

// checkError returns a failure message if err is nil.
func checkError(err error) string {
	if err == nil {
		return "expected an error, got none" //nolint:goconst
	}
	return ""
}

// checkNoError returns a failure message if err is non-nil.
func checkNoError(err error) string {
	if err != nil {
		return fmt.Sprintf("expected no error, got: %v", err)
	}
	return ""
}

// checkEqualError returns a failure message if gotErr does not match wantErr. It
// uses errors.Is when wantErr is an error. It compares error messages directly
// when wantErr is a string. This matches testify behavior.
func checkEqualError(gotErr error, wantErr any) string {
	switch w := wantErr.(type) {
	case nil:
		return checkNoError(gotErr)
	case error:
		if !errors.Is(gotErr, w) {
			return fmt.Sprintf("got error:  %v\nwant error: %v", gotErr, w)
		}
		return ""
	case string:
		if gotErr == nil {
			return "expected an error, got none"
		}
		if gotErr.Error() != w {
			return fmt.Sprintf("got error:  %q\nwant error: %q", gotErr.Error(), w)
		}
		return ""
	default:
		return fmt.Sprintf("unsupported wantErr type: %T", wantErr)
	}
}

// checkTrue returns a failure message if value is false.
func checkTrue(value bool) string {
	if !value {
		return "expected true, got false"
	}
	return ""
}

// checkFalse returns a failure message if value is true.
func checkFalse(value bool) string {
	if value {
		return "expected false, got true"
	}
	return ""
}

// checkNil returns a failure message if obj is not nil. It handles typed nil
// values like pointers, maps, slices, and interface-wrapped nils. Use this for
// accurate nil checks on complex types.
func checkNil(obj any) string {
	if !isNil(obj) {
		return fmt.Sprintf("expected nil, got: %v", obj)
	}
	return ""
}

// checkNotNil returns a failure message if obj is nil (see isNil for typed nil
// handling).
func checkNotNil(obj any) string {
	if isNil(obj) {
		return "expected not nil, got nil"
	}
	return ""
}

// checkLen returns a failure message if obj does not have the given length.
// obj must be an array, chan, map, slice, or string.
func checkLen(obj any, length int) string {
	rv := reflect.ValueOf(obj)
	switch rv.Kind() { //nolint:exhaustive
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		if rv.Len() != length {
			return fmt.Sprintf("got length: %d, want: %d", rv.Len(), length)
		}
		return ""
	default:
		return fmt.Sprintf("expected length of array, chan, map, slice or string, got: %v", rv.Kind())
	}
}

// checkPanic returns a failure message if f does not panic.
func checkPanic(f func()) string {
	if panicked, _ := recoverPanic(f); !panicked {
		return "expected a panic, got none"
	}
	return ""
}

// checkNotPanic returns a failure message if f panics.
func checkNotPanic(f func()) string {
	if panicked, value := recoverPanic(f); panicked {
		return fmt.Sprintf("expected no panic, got: %v", value)
	}
	return ""
}

// recoverPanic runs f and reports whether it panicked, with the recovered value.
func recoverPanic(f func()) (panicked bool, value any) {
	defer func() {
		if r := recover(); r != nil {
			panicked, value = true, r
		}
	}()
	f()
	return false, nil
}

// deepEqual reports deep equality of two values, comparing byte slices by content.
func deepEqual(expected, actual any) bool {
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

// isNil reports whether obj is nil. It handles typed nil values like pointers,
// maps, slices, channels, functions, and interface-wrapped nils. Use this for
// accurate nil detection on all reference types.
func isNil(obj any) bool {
	if obj == nil {
		return true
	}
	rv := reflect.ValueOf(obj)
	switch rv.Kind() { //nolint:exhaustive
	case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice,
		reflect.UnsafePointer:
		return rv.IsNil()
	default:
		return false
	}
}

// checkContainsSlice returns a failure message if s does not contain elem.
func checkContainsSlice[T comparable](s []T, elem T) string {
	if slices.Contains(s, elem) {
		return ""
	}
	return fmt.Sprintf("expected slice to contain %v, got: %v", elem, s)
}

// checkNotContainsSlice returns a failure message if s contains elem.
func checkNotContainsSlice[T comparable](s []T, elem T) string {
	if slices.Contains(s, elem) {
		return fmt.Sprintf("expected slice not to contain %v", elem)
	}
	return ""
}

// checkEmpty returns a failure message if obj is not empty.
func checkEmpty(obj any) string {
	rv := reflect.ValueOf(obj)
	switch rv.Kind() { //nolint:exhaustive
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		if rv.Len() == 0 {
			return ""
		}
		return fmt.Sprintf("expected empty, got length %d", rv.Len())
	case reflect.Chan:
		if rv.IsNil() || rv.Cap() == 0 {
			return ""
		}
		return "expected empty channel"
	default:
		if isNil(obj) {
			return ""
		}
		return fmt.Sprintf("expected empty, got %v", obj)
	}
}

// checkNotEmpty returns a failure message if obj is empty.
func checkNotEmpty(obj any) string {
	rv := reflect.ValueOf(obj)
	switch rv.Kind() { //nolint:exhaustive
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		if rv.Len() > 0 {
			return ""
		}
		return "expected non-empty, got empty"
	case reflect.Chan:
		if !rv.IsNil() && rv.Cap() > 0 {
			return ""
		}
		return "expected non-empty channel"
	default:
		if !isNil(obj) {
			return ""
		}
		return "expected non-nil, got nil"
	}
}

// checkErrorContains returns a failure message if err is nil or does not contain substr.
func checkErrorContains(err error, substr string) string {
	if err == nil {
		return "expected an error, got none"
	}
	if !strings.Contains(err.Error(), substr) {
		return fmt.Sprintf("error %q does not contain %q", err.Error(), substr)
	}
	return ""
}

// checkGreaterOrEqual returns a failure message if got is less than want. It
// supports all numeric types and strings. Use this for value comparison checks.
func checkGreaterOrEqual(got, want any) string {
	if reflect.ValueOf(got).Interface() == reflect.ValueOf(want).Interface() {
		return ""
	}
	gv := reflect.ValueOf(got)
	wv := reflect.ValueOf(want)
	switch gv.Kind() { //nolint:exhaustive
	case reflect.String:
		if gv.String() >= wv.String() {
			return ""
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if gv.Int() >= wv.Int() {
			return ""
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if gv.Uint() >= wv.Uint() {
			return ""
		}
	default:
		return fmt.Sprintf("unsupported type for GreaterOrEqual: %T", got)
	}
	return fmt.Sprintf("expected %v >= %v", got, want)
}

// checkLessOrEqual returns a failure message if got is greater than want. It
// supports all numeric types and strings. Use this for value comparison checks.
func checkLessOrEqual(got, want any) string {
	if reflect.ValueOf(got).Interface() == reflect.ValueOf(want).Interface() {
		return ""
	}
	gv := reflect.ValueOf(got)
	wv := reflect.ValueOf(want)
	switch gv.Kind() { //nolint:exhaustive
	case reflect.String:
		if gv.String() <= wv.String() {
			return ""
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if gv.Int() <= wv.Int() {
			return ""
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if gv.Uint() <= wv.Uint() {
			return ""
		}
	default:
		return fmt.Sprintf("unsupported type for LessOrEqual: %T", got)
	}
	return fmt.Sprintf("expected %v <= %v", got, want)
}

// checkErrorIs returns a failure message if errors.Is(err, target) is false.
func checkErrorIs(err, target error) string {
	if err == nil && target == nil {
		return ""
	}
	if !errors.Is(err, target) {
		return fmt.Sprintf("error %v does not match %v", err, target)
	}
	return ""
}

// checkContainsString returns a failure message if s does not contain substr.
func checkContainsString(s, substr string) string {
	if strings.Contains(s, substr) {
		return ""
	}
	return fmt.Sprintf("string %q does not contain %q", s, substr)
}

// checkNotContainsString returns a failure message if s contains substr.
func checkNotContainsString(s, substr string) string {
	if !strings.Contains(s, substr) {
		return ""
	}
	return fmt.Sprintf("string %q should not contain %q", s, substr)
}

// checkErrorAs returns a failure message if errors.As(err, target) is false.
func checkErrorAs(err error, target any) string {
	if err == nil && target == nil {
		return ""
	}
	if !errors.As(err, target) {
		return fmt.Sprintf("error %v does not match type %T", err, target)
	}
	return ""
}

// checkJSONEq returns a failure message if the two JSON strings are not equal.
func checkJSONEq(expected, actual string) string {
	var expObj, actObj any
	if err := json.Unmarshal([]byte(expected), &expObj); err != nil {
		return fmt.Sprintf("failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(actual), &actObj); err != nil {
		return fmt.Sprintf("failed to unmarshal actual JSON: %v", err)
	}
	if !deepEqual(expObj, actObj) {
		return fmt.Sprintf("got:  %s\nwant: %s", actual, expected)
	}
	return ""
}

// checkImplements returns a failure message if object does not implement iface.
// iface must be a pointer to an interface type (e.g., (*MyInterface)(nil)).
func checkImplements(object, iface any) string {
	ifaceType := reflect.TypeOf(iface).Elem()
	objType := reflect.TypeOf(object)
	if !objType.Implements(ifaceType) {
		return fmt.Sprintf("type %s does not implement interface %s", objType.Name(), ifaceType.Name())
	}
	return ""
}

// checkIsType returns a failure message if object is not the same type as expectedType.
func checkIsType(expectedType, object any) string {
	if reflect.TypeOf(expectedType) == reflect.TypeOf(object) {
		return ""
	}
	return fmt.Sprintf("expected type %T, got type %T", expectedType, object)
}

// checkNotPanics returns a failure message if f panics, or the empty string on success.
func checkNotPanics(f func()) string {
	ch := make(chan string)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- fmt.Sprintf("unexpected panic: %v", r)
			}
		}()
		f()
		ch <- ""
	}()
	return <-ch
}
