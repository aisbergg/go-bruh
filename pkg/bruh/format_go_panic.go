package bruh

import (
	"github.com/aisbergg/go-bruh/internal/stringbuilder"
)

// GoPanicFormatter is an error formatter that produces error traces similar to
// Go's panics. Most recent calls are at the top.
//
// # Output Format
//
//	errorMsg1: errorMsg2: errorMsgN
//
//	function1
//		file1:line1 +0x123456
//	function2
//		file2:line2 +0x123456
//	functionN
//		fileN:lineN +0x123456
func GoPanicFormatter(b []byte, unpacker *Unpacker) []byte {
	if unpacker.Error() == nil {
		return b
	}
	stack := unpacker.CombinedStack()
	// allocate a large buffer to avoid later reallocations
	// message: 80 per error
	// location: 160 per location
	builder := stringbuilder.New(b)
	guessCap := unpacker.ChainLen() * 80
	if len(stack) != 0 {
		guessCap += len(stack) * 160
	}
	builder.Grow(guessCap)
	unpacker.AppendMessageBuilder(builder)

	if len(stack) != 0 {
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		for i, s := range stack {
			builder.WriteString(s.Name)
			builder.WriteString("()\n\t")
			builder.WriteString(s.File)
			builder.WriteByte(':')
			builder.WriteInt(int64(s.Line))
			builder.WriteString(" +0x")
			builder.WriteIntAsHex(int64(s.ProgramCounter2))
			if i < len(stack)-1 {
				builder.WriteByte('\n')
			}
		}
	}
	return builder.Bytes()
}
