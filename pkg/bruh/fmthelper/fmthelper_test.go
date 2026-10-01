package fmthelper

import (
	"testing"

	"github.com/aisbergg/go-bruh/internal/lib/test/assert"
)

func TestDigitsInNumber(t *testing.T) {
	t.Parallel()

	assertDigits := func(n, exp int) {
		t.Helper()
		assert.Equal(t, DigitsInNumber(n), exp)
	}

	assertDigits(0, 1)
	assertDigits(7, 1)
	assertDigits(10, 2)
	assertDigits(999, 3)
	assertDigits(9999, 4)
	assertDigits(-12345, 5)
	assertDigits(999999, 6)
	assertDigits(1000000, 7)
	assertDigits(99999999, 8)
	assertDigits(100000000, 9)
	assertDigits(1000000000, 10)
}

func TestStringBuilder(t *testing.T) {
	t.Parallel()

	t.Run("WriteMethods", func(t *testing.T) {
		builder := New([]byte("ab"))

		builder.Write([]byte("cd"))

		builder.WriteByte('e')

		builder.WriteString("fg")

		assert.Equal(t, builder.Len(), 7)
		assert.Equal(t, builder.String(), "abcdefg")
		assert.Equal(t, builder.Bytes(), []byte("abcdefg"))
	})

	t.Run("GrowPreservesContent", func(t *testing.T) {
		builder := New(make([]byte, 0, 1))

		builder.WriteString("x")

		builder.Grow(8)

		assert.Equal(t, builder.String(), "x")
		assert.True(t, cap(builder.Bytes()) >= builder.Len()+8)
	})

	t.Run("WriteStringIndent", func(t *testing.T) {
		singleLine := New(nil)
		singleLine.WriteStringIndent("abc", "  ")
		assert.Equal(t, singleLine.String(), "abc")
		assert.Equal(t, singleLine.Len(), 3)

		multiLine := New(nil)
		multiLine.WriteStringIndent("a\nb\nc", "  ")
		assert.Equal(t, multiLine.String(), "a\n  b\n  c")
		assert.Equal(t, multiLine.Len(), 9)
	})

	t.Run("IntegerWriters", func(t *testing.T) {
		builder := New(nil)

		builder.WriteInt(-42)
		builder.WriteByte('|')
		builder.WriteIntAsHex(255)
		builder.WriteByte('|')
		builder.WriteUint(42)
		builder.WriteByte('|')
		builder.WriteUintAsHex(255)

		assert.Equal(t, builder.String(), "-42|ff|42|ff")
	})
}

func TestColorer(t *testing.T) {
	t.Parallel()

	t.Run("DisabledSkipsAnsiCodes", func(t *testing.T) {
		builder := New(nil)
		colorer := NewColorer(builder, false)

		colorer.Color(Red, Bold)
		colorer.ColorRGB(1, 2, 3)
		colorer.BGColorRGB(4, 5, 6)
		colorer.Reset()
		colorer.ColoredText("hello", Green)
		colorer.ColoredInt(7, Blue)

		assert.Equal(t, builder.String(), "hello7")
	})

	t.Run("EnabledWritesAnsiCodes", func(t *testing.T) {
		builder := New(nil)
		colorer := NewColorer(builder, true)

		colorer.Color(Red, Bold)
		colorer.ColoredText("hello", Green)
		colorer.ColoredInt(7, Blue)
		colorer.ColorRGB(-1, 42, 999)
		colorer.BGColorRGB(1, 2, 3)
		colorer.Reset()

		assert.Equal(t, builder.String(), string(Red)+string(Bold)+
			string(Green)+"hello"+string(Reset)+
			string(Blue)+"7"+string(Reset)+
			"\033[38;2;0;42;255m"+
			"\033[48;2;1;2;3m"+
			string(Reset))
	})
}
