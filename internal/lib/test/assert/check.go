// Unexported assertion checks. Each check returns a failure message, or the
// empty string on success. formatMsg combines a failure message with an
// optional custom message.

package assert

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

// checkEqual returns a failure message if got and want are not deeply equal.
// Byte slices are compared by content; nil and empty slices are not equal.
// Maps are compared without regard to key order; slices are compared with
// regard to element order.
func checkEqual[T any](got, want T) string {
	if equal(got, want) {
		return ""
	}
	return fmt.Sprintf("got:  %v\nwant: %v", got, want)
}

// checkNotEqual returns a failure message if got and want are deeply equal.
func checkNotEqual[T any](got, want T) string {
	if equal(got, want) {
		return fmt.Sprintf("got equal to want: %v", got)
	}
	return ""
}

// checkEqualUnordered returns a failure message if got and want do not contain
// the same elements regardless of order. Duplicate elements must occur equally
// often in both slices (multiset semantics). Elements are compared deeply. Nil
// and empty slices are equal.
func checkEqualUnordered[T any](got, want []T) string {
	matched := make([]bool, len(want))
	var extra []T
	for _, g := range got {
		found := false
		for i, w := range want {
			if !matched[i] && equal(g, w) {
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
		return "expected an error, got none"
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

// checkEqualError returns a failure message if gotErr does not match wantErr
// via errors.Is. A nil wantErr requires gotErr to be nil.
func checkEqualError(gotErr, wantErr error) string {
	if wantErr == nil {
		return checkNoError(gotErr)
	}
	if !errors.Is(gotErr, wantErr) {
		return fmt.Sprintf("got error:  %v\nwant error: %v", gotErr, wantErr)
	}
	return ""
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

// checkNil returns a failure message if obj is not nil, including typed nil
// values such as (*T)(nil), map(nil), slice(nil), and interfaces wrapping
// those nils.
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

// formatMsg builds the final failure message. defaultMsg describes the failed
// check and is always kept; a custom fmt.Sprintf-style message is prepended.
// With no args the message is used verbatim (no verb escaping needed). An
// empty defaultMsg yields the empty string (success).
func formatMsg(defaultMsg, msg string, args ...any) string {
	if defaultMsg == "" {
		return ""
	}
	if msg == "" {
		return defaultMsg
	}
	if len(args) == 0 {
		return msg + "\n" + defaultMsg
	}
	return fmt.Sprintf(msg, args...) + "\n" + defaultMsg
}

// formatMsgArgs builds the final failure message from a msgAndArgs tail:
// empty tail yields defaultMsg, a single string is used verbatim, otherwise
// the tail is rendered fmt.Sprintf-style if its first element is a string.
func formatMsgArgs(defaultMsg string, msgAndArgs ...any) string {
	switch len(msgAndArgs) {
	case 0:
		return defaultMsg
	case 1:
		if msg, ok := msgAndArgs[0].(string); ok {
			return formatMsg(defaultMsg, "%s", msg)
		}
	default:
		if format, ok := msgAndArgs[0].(string); ok {
			return formatMsg(defaultMsg, format, msgAndArgs[1:]...)
		}
	}
	return fmt.Sprint(msgAndArgs...) + "\n" + defaultMsg
}

// equal reports deep equality of two values, comparing byte slices by content.
func equal(expected, actual any) bool { //nolint: revive
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

// isNil reports whether obj is nil, including typed nil values such as
// unsafe.Pointer(nil), map(nil), slice(nil), and interface values wrapping
// those nils.
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
