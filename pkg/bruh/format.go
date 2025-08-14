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
	// If the error implements the messager interface, we have to build the full
	// error message ourselves.
	if _, ok := err.(messager); ok {
		b := formatMessageOnly(nil, err)
		return unsafe.String(unsafe.SliceData(b), len(b)) //nolint:gosec
	}
	// The full concatenated error message is likely already available in the
	// error itself, so we can just return it.
	return err.Error()
}

// AppendMessage does the same as [Message] but appends the formatted message to
// the provided byte slice.
func AppendMessage(b []byte, err error) []byte {
	return formatMessageOnly(b, err)
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
	return unsafe.String(unsafe.SliceData(b), len(b)) //nolint:gosec
}

// AppendStringFormat does the same as [StringFormat] but appends the formatted
// string to the provided byte slice.
func AppendStringFormat(b []byte, err error, f Formatter, unpackAll ...bool) []byte {
	if err == nil {
		return b
	}

	// Format error without a stack trace by simply combining all error
	// messages. (fast path)
	if f == nil {
		return formatMessageOnly(b, err)
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

// formatMessageOnly formats an error by concatenating all error messages.
//
// # Output Format
//
//	errorMsg1: errorMsg2: errorMsgN
func formatMessageOnly(b []byte, err error) []byte {
	// allocate 80 bytes per message to avoid reallocations later on
	numBytesToAlloc := 0
	for uerr := err; uerr != nil; uerr = Unwrap(uerr) {
		numBytesToAlloc += 80
	}
	if cap(b) == 0 {
		b = make([]byte, 0, numBytesToAlloc)
	} else if cap(b)-len(b) < numBytesToAlloc {
		newBuf := make([]byte, len(b), cap(b)+numBytesToAlloc)
		copy(newBuf, b)
		b = newBuf
	}
	return appendMessageToBuffer(b, err)
}

func appendMessageToBuffer(buf []byte, err error) []byte {
	var messageWritten bool
	for ; err != nil; err = Unwrap(err) {
		var msg string
		var cont bool
		if e, ok := err.(messager); ok {
			msg = e.Message()
			cont = true
		} else {
			msg = err.Error()
		}

		if msg == "" {
			continue
		}

		if messageWritten {
			buf = append(buf, ": "...)
		}
		buf = append(buf, msg...)
		messageWritten = true

		if !cont {
			break
		}
	}
	return buf
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
