package fmthelper

import (
	"testing"

	"github.com/aisbergg/go-bruh/internal/testutils"
)

func TestDigitsInNumber(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	assertDigits := func(n, exp int) {
		t.Helper()
		assert.Equal(exp, DigitsInNumber(n))
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
		assert := testutils.NewAssert(t)
		builder := New([]byte("ab"))

		n, err := builder.Write([]byte("cd"))
		assert.NoError(err)
		assert.Equal(2, n)

		err = builder.WriteByte('e')
		assert.NoError(err)

		n, err = builder.WriteString("fg")
		assert.NoError(err)
		assert.Equal(2, n)

		assert.Equal(7, builder.Len())
		assert.Equal("abcdefg", builder.String())
		assert.Equal([]byte("abcdefg"), builder.Bytes())
	})

	t.Run("GrowPreservesContent", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		builder := New(make([]byte, 0, 1))

		_, err := builder.WriteString("x")
		assert.NoError(err)

		builder.Grow(8)

		assert.Equal("x", builder.String())
		assert.True(cap(builder.Bytes()) >= builder.Len()+8)
	})

	t.Run("WriteStringIndent", func(t *testing.T) {
		assert := testutils.NewAssert(t)

		singleLine := New(nil)
		n, err := singleLine.WriteStringIndent("abc", "  ")
		assert.NoError(err)
		assert.Equal(3, n)
		assert.Equal("abc", singleLine.String())

		multiLine := New(nil)
		n, err = multiLine.WriteStringIndent("a\nb\nc", "  ")
		assert.NoError(err)
		assert.Equal(9, n)
		assert.Equal("a\n  b\n  c", multiLine.String())
	})

	t.Run("IntegerWriters", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		builder := New(nil)

		builder.WriteInt(-42)
		builder.WriteByte('|')
		builder.WriteIntAsHex(255)
		builder.WriteByte('|')
		builder.WriteUint(42)
		builder.WriteByte('|')
		builder.WriteUintAsHex(255)

		assert.Equal("-42|ff|42|ff", builder.String())
	})
}

func TestColorer(t *testing.T) {
	t.Parallel()

	t.Run("DisabledSkipsAnsiCodes", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		builder := New(nil)
		colorer := NewColorer(builder, false)

		colorer.Color(Red, Bold)
		colorer.ColorRGB(1, 2, 3)
		colorer.BGColorRGB(4, 5, 6)
		colorer.Reset()
		colorer.ColoredText("hello", Green)
		colorer.ColoredInt(7, Blue)

		assert.Equal("hello7", builder.String())
	})

	t.Run("EnabledWritesAnsiCodes", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		builder := New(nil)
		colorer := NewColorer(builder, true)

		colorer.Color(Red, Bold)
		colorer.ColoredText("hello", Green)
		colorer.ColoredInt(7, Blue)
		colorer.ColorRGB(-1, 42, 999)
		colorer.BGColorRGB(1, 2, 3)
		colorer.Reset()

		assert.Equal(
			string(Red)+string(Bold)+
				string(Green)+"hello"+string(Reset)+
				string(Blue)+"7"+string(Reset)+
				"\033[38;2;0;42;255m"+
				"\033[48;2;1;2;3m"+
				string(Reset),
			builder.String(),
		)
	})
}
