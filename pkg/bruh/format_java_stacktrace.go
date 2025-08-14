package bruh

import (
	"github.com/aisbergg/go-bruh/internal/stringbuilder"
)

// JavaStackTraceFormatter is an error formatter that produces error traces similar
// to Java's stack traces. Most recent calls are at the top.
//
// # Output Format
//
//	<typeNameN>: <errorMsgN>
//	    at <function1> (<file1>:<line1>)
//	    at <function2> (<file2>:<line2>)
//	    at <functionN> (<fileN>:<lineN>)
//	Caused by: <typeName2>: <errorMsg2>
//	    at <function1> (<file1>:<line1>)
//	    at <function2> (<file2>:<line2>)
//	    at <functionN> (<fileN>:<lineN>)
//	Caused by: <typeName1>: <errorMsg1>
//	    at <function1> (<file1>:<line1>)
//	    at <function2> (<file2>:<line2>)
//	    at <functionN> (<fileN>:<lineN>)
func JavaStackTraceFormatter(b []byte, unpacker *Unpacker) []byte {
	if unpacker.Error() == nil {
		return b
	}
	upkErr := unpacker.Unpack()
	// allocate a large buffer to avoid later reallocations
	// message: 80 per error
	// location: 160 per location
	builder := stringbuilder.New(b)
	guessCap := len(upkErr) * 80
	for _, upkElm := range upkErr {
		guessCap += len(upkElm.PartialStack) * 160
	}
	builder.Grow(guessCap)

	for i, upkElm := range upkErr {
		if i > 0 {
			builder.WriteString("Caused by: ")
		}
		builder.WriteString(typeName(upkElm.Err))
		builder.WriteString(": ")
		if upkElm.Msg != "" {
			builder.WriteString(upkElm.Msg)
		} else {
			builder.WriteString("_")
		}
		for _, s := range upkElm.PartialStack {
			builder.WriteByte('\n')
			builder.WriteString("    at ")
			builder.WriteString(s.Name)
			builder.WriteString(" (")
			builder.WriteString(s.File)
			builder.WriteByte(':')
			builder.WriteInt(int64(s.Line))
			builder.WriteByte(')')
		}
		if i < len(upkErr)-1 {
			builder.WriteByte('\n')
		}
	}
	return builder.Bytes()
}
