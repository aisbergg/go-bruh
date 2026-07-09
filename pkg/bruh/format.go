package bruh

import (
	"reflect"
	"unsafe"
)

// Formatter turns an unpacked error into a formatted string.
type Formatter func(b []byte, unpacker *Unpacker) []byte

// Message returns the combined error message without a trace.
func Message(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// AppendMessage does the same as [Message] but appends the formatted message to
// the provided byte slice.
func AppendMessage(b []byte, err error) []byte {
	if err == nil {
		return b
	}
	b = append(b, err.Error()...)
	return b
}

// MessageLastN returns the combined error message of the last n errors in the
// chain. If n is greater than the number of errors in the chain, the message of
// all errors is returned.
func MessageLastN(err error, n int) string {
	if err == nil || n <= 0 {
		return ""
	}
	lastN := make([]error, n)
	lastNIdx := 0
	for uerr := err; uerr != nil; uerr = Unwrap(uerr) {
		lastN[lastNIdx] = uerr
		lastNIdx = (lastNIdx + 1) % n
	}
	for i := range n {
		if lastErr := lastN[(lastNIdx+i)%n]; lastErr != nil {
			return lastErr.Error()
		}
	}
	return ""
}

// String returns a formatted string representation of the provided error.
//
// The default error formatter [BruhFormatter] is used. If you want to use a
// different format, you can use [StringFormat] instead.:
//
//	StringFormat(err, bruh.BruhStackedFormatter)
//
// If err is nil, the returned string will be empty.
func String(err error) string {
	return StringFormat(err, BruhFormatter)
}

// AppendString does the same as [String] but appends the formatted string to
// the provided byte slice.
func AppendString(b []byte, err error) []byte {
	return AppendStringFormat(b, err, BruhFormatter)
}

// StringFormat returns a formatted string representation of the provided error
// and formatter. If the formatter is nil, the error will be formatted without a
// stack trace.
//
// In rare cases where you want errors without a stack trace (external errors)
// to be unpacked as well, set `unpackAll` to true. This will have the effect of
// non-trace errors to appear as discrete errors in the formatted output.
func StringFormat(err error, f Formatter, unpackAll ...bool) string {
	if err == nil {
		return ""
	}
	b := AppendStringFormat(nil, err, f, unpackAll...)
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// AppendStringFormat does the same as [StringFormat] but appends the formatted
// string to the provided byte slice.
func AppendStringFormat(b []byte, err error, f Formatter, unpackAll ...bool) []byte {
	if err == nil {
		return b
	}

	if f == nil {
		msg := err.Error()
		if msg == "" {
			return b
		}
		b = append(b, msg...)
		return b
	}

	// We use an unpacker struct that provides the formatter with detailed
	// information about the error. Because we have full control over the
	// unpacker here, we can allocate the unpacker here once and reuse it for
	// consecutive calls later.
	unpacker := newUnpacker(err, len(unpackAll) > 0 && unpackAll[0])
	b = f(b, unpacker)
	disposeUnpacker(unpacker)
	return b
}

// typeName returns the type of the error. e.g. `*bruh.Err`.
func typeName(err error) string {
	if err == nil {
		return "nil"
	}
	if _, ok := err.(*Err); ok {
		return "*bruh.Err"
	}
	return reflect.TypeOf(err).String()
}
