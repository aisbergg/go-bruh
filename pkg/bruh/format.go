package bruh

import (
	"reflect"
	"strings"
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

// MessageSummary returns the messages of the first head and last tail errors in
// the chain, joined by ": ". If errors in between are skipped, ellipsis is
// inserted in their place. An empty ellipsis inserts nothing. Errors with an
// empty message are not counted.
//
//	err := bruh.Wrap(bruh.Wrap(io.ErrUnexpectedEOF, "decoding data"), "reading config")
//	bruh.MessageSummary(err, 1, 1, "...") // "reading config: ...: unexpected EOF"
func MessageSummary(err error, head, tail int, ellipsis string) string {
	chainDepth := 1
	for uerr := err; uerr != nil; uerr = Unwrap(uerr) {
		chainDepth++
	}
	msgs := make([]string, 0, chainDepth)
	for uerr := err; uerr != nil; uerr = Unwrap(uerr) {
		if msg := ownMessage(uerr); msg != "" {
			msgs = append(msgs, msg)
		}
	}
	head = max(head, 0)
	tail = max(tail, 0)
	if head+tail >= len(msgs) {
		return Message(err)
	}
	parts := append([]string{}, msgs[:head]...)
	if ellipsis != "" {
		parts = append(parts, ellipsis)
	}
	parts = append(parts, msgs[len(msgs)-tail:]...)
	return strings.Join(parts, ": ")
}

// ownMessage returns the message of err without the text of its wrapped error.
func ownMessage(err error) string {
	if m, ok := err.(messager); ok {
		return m.Message()
	}
	msg := err.Error()
	if next := Unwrap(err); next != nil {
		if nextMsg := next.Error(); strings.HasSuffix(msg, nextMsg) {
			msg = strings.TrimSuffix(msg[:len(msg)-len(nextMsg)], ": ")
		}
	}
	return msg
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
