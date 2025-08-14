// Package stringbuilder provides a simple string builder implementation
// that allows for efficient string concatenation and manipulation.
package stringbuilder

import (
	"strconv"
	"unsafe"
)

// StringBuilder is a simple string builder that uses a byte slice as its
// internal buffer. It is mostly the same as [strings.Builder] but it provides
// more specialized methods for writing integers, hexadecimal values and such.
type StringBuilder struct {
	buf []byte
}

// New creates and returns a new StringBuilder initialized with the provided buffer.
func New(b []byte) *StringBuilder {
	return &StringBuilder{b}
}

// Grow ensures that the internal buffer has enough capacity to accommodate n more bytes.
// If there is not enough space, it allocates a new buffer with additional capacity.
func (b *StringBuilder) Grow(n int) {
	if cap(b.buf)-len(b.buf) < n {
		newBuf := make([]byte, len(b.buf), cap(b.buf)+n)
		copy(newBuf, b.buf)
		b.buf = newBuf
	}
}

// Len returns the length of the buffer.
func (b *StringBuilder) Len() int {
	return len(b.buf)
}

// String returns the accumulated string.
func (b *StringBuilder) String() string {
	return unsafe.String(unsafe.SliceData(b.buf), len(b.buf)) //nolint:gosec
}

// Bytes returns the internally used bytes buffer.
func (b *StringBuilder) Bytes() []byte {
	return b.buf
}

// Write appends the contents of p to b's buffer.
// Write always returns len(p), nil.
func (b *StringBuilder) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// WriteByte appends the byte c to b's buffer.
// The returned error is always nil.
func (b *StringBuilder) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

// WriteString appends the contents of s to b's buffer.
// It returns the length of s and a nil error.
func (b *StringBuilder) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

// WriteInt appends the given integer to b's buffer.
func (b *StringBuilder) WriteInt(value int64) {
	b.buf = strconv.AppendInt(b.buf, value, 10)
}

// WriteIntAsHex formats the given integer value as hexadecimal string and
// appends it to b's buffer.
func (b *StringBuilder) WriteIntAsHex(value int64) {
	b.buf = strconv.AppendInt(b.buf, value, 16)
}
