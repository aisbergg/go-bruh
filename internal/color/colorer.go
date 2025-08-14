// Package color provides a simple way to add ANSI color codes to strings using
// a [stringbuilder.StringBuilder].
package color

import (
	"github.com/aisbergg/go-bruh/internal/stringbuilder"
)

// ANSICode represents an ANSI escape code for text formatting.
type ANSICode string

// ANSI color codes for text formatting.
const (
	Reset     ANSICode = "\033[0m"
	Bold      ANSICode = "\033[1m"
	Faint     ANSICode = "\033[2m"
	Underline ANSICode = "\033[4m"
	Blink     ANSICode = "\033[5m"

	Black         ANSICode = "\033[30m"
	Red           ANSICode = "\033[31m"
	Green         ANSICode = "\033[32m"
	Yellow        ANSICode = "\033[33m"
	Blue          ANSICode = "\033[34m"
	Magenta       ANSICode = "\033[35m"
	Cyan          ANSICode = "\033[36m"
	White         ANSICode = "\033[37m"
	BrightBlack   ANSICode = "\033[90m"
	BrightRed     ANSICode = "\033[91m"
	BrightGreen   ANSICode = "\033[92m"
	BrightYellow  ANSICode = "\033[93m"
	BrightBlue    ANSICode = "\033[94m"
	BrightMagenta ANSICode = "\033[95m"
	BrightCyan    ANSICode = "\033[96m"
	BrightWhite   ANSICode = "\033[97m"

	BGBlack         ANSICode = "\033[40m"
	BGRed           ANSICode = "\033[41m"
	BGGreen         ANSICode = "\033[42m"
	BGYellow        ANSICode = "\033[43m"
	BGBlue          ANSICode = "\033[44m"
	BGMagenta       ANSICode = "\033[45m"
	BGCyan          ANSICode = "\033[46m"
	BGWhite         ANSICode = "\033[47m"
	BGBrightBlack   ANSICode = "\033[100m"
	BGBrightRed     ANSICode = "\033[101m"
	BGBrightGreen   ANSICode = "\033[102m"
	BGBrightYellow  ANSICode = "\033[103m"
	BGBrightBlue    ANSICode = "\033[104m"
	BGBrightMagenta ANSICode = "\033[105m"
	BGBrightCyan    ANSICode = "\033[106m"
	BGBrightWhite   ANSICode = "\033[107m"
)

// Colorer adds ANSI color codes to a string builder.
type Colorer struct {
	builder *stringbuilder.StringBuilder
	enabled bool
}

// NewColorer creates a new [Colorer].
func NewColorer(builder *stringbuilder.StringBuilder, enabled bool) Colorer {
	return Colorer{
		builder: builder,
		enabled: enabled,
	}
}

// Reset resets the text color to the default color.
func (c Colorer) Reset() {
	if c.enabled {
		c.builder.WriteString(string(Reset))
	}
}

// Color sets the text color to the specified color.
func (c Colorer) Color(color ...ANSICode) {
	if c.enabled {
		for _, code := range color {
			c.builder.WriteString(string(code))
		}
	}
}

// ColorRGB sets the text color to the specified RGB value.
func (c Colorer) ColorRGB(r, g, b int) {
	if c.enabled {
		r = max(min(r, 255), 0)
		g = max(min(g, 255), 0)
		b = max(min(b, 255), 0)
		c.builder.WriteString("\033[38;2;")
		c.builder.WriteInt(int64(r))
		c.builder.WriteByte(';')
		c.builder.WriteInt(int64(g))
		c.builder.WriteByte(';')
		c.builder.WriteInt(int64(b))
		c.builder.WriteByte('m')
	}
}

// BGColorRGB sets the background color to the specified RGB value.
func (c Colorer) BGColorRGB(r, g, b int) {
	if c.enabled {
		r = max(min(r, 255), 0)
		g = max(min(g, 255), 0)
		b = max(min(b, 255), 0)
		c.builder.WriteString("\033[48;2;")
		c.builder.WriteInt(int64(r))
		c.builder.WriteByte(';')
		c.builder.WriteInt(int64(g))
		c.builder.WriteByte(';')
		c.builder.WriteInt(int64(b))
		c.builder.WriteByte('m')
	}
}

// ColoredText writes the text in the specified color.
func (c Colorer) ColoredText(text string, color ...ANSICode) {
	if !c.enabled {
		c.builder.WriteString(text)
		return
	}
	c.Color(color...)
	c.builder.WriteString(text)
	c.Reset()
}

// ColoredInt writes the integer in the specified color.
func (c Colorer) ColoredInt(value int64, color ...ANSICode) {
	if !c.enabled {
		c.builder.WriteInt(value)
		return
	}
	c.Color(color...)
	c.builder.WriteInt(value)
	c.Reset()
}
