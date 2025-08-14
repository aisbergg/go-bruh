package bruh

import (
	"github.com/aisbergg/go-bruh/internal/stringbuilder"
)

// PythonTracebackFormatter is an error formatter that produces error traces
// similar to Python's tracebacks. Most recent calls are at the bottom.
//
// # Output Format
//
//	Traceback (most recent call last):
//	  File "<file3>", line <line3>, in <function3>
//	  File "<file2>", line <line2>, in <function2>
//	  File "<file1>", line <line1>, in <function1>
//	<typeName2>: <errorMsg2>
//
//	The above exception was the direct cause of the following exception:
//
//	Traceback (most recent call last):
//	  File "<file3>", line <line3>, in <function3>
//	  File "<file2>", line <line2>, in <function2>
//	  File "<file1>", line <line1>, in <function1>
//	<typeName1>: <errorMsg1>
func PythonTracebackFormatter(b []byte, unpacker *Unpacker) []byte {
	return formatPythonTraceback(b, unpacker, false)
}

// FormatPythonTracebackSourced is an error formatter that produces error
// traces similar to Python's tracebacks. It includes the source code of the
// lines in the trace. The source lines can only be included if the source code
// is available at runtime. Most recent calls are at the bottom.
//
// # Output Format
//
//	Traceback (most recent call last):
//	  File "<file3>", line <line3>, in <function3>
//	    <source line3>
//	  File "<file2>", line <line2>, in <function2>
//	    <source line2>
//	  File "<file1>", line <line1>, in <function1>
//	    <source line1>
//	<typeName2>: <errorMsg2>
//
//	The above exception was the direct cause of the following exception:
//
//	Traceback (most recent call last):
//	  File "<file3>", line <line3>, in <function3>
//	    <source line3>
//	  File "<file2>", line <line2>, in <function2>
//	    <source line2>
//	  File "<file1>", line <line1>, in <function1>
//	    <source line1>
//	<typeName1>: <errorMsg1>
func FormatPythonTracebackSourced(b []byte, unpacker *Unpacker) []byte {
	return formatPythonTraceback(b, unpacker, true)
}

func formatPythonTraceback(b []byte, unpacker *Unpacker, includeSource bool) []byte {
	if unpacker.Error() == nil {
		return b
	}
	upkErr := unpacker.Unpack()
	// allocate a large buffer to avoid later reallocations
	// fixed text: 110 per error
	// message: 80 per error
	// location: 160 per location
	// source line: 124 per location
	builder := stringbuilder.New(b)
	guessCap := len(upkErr) * (80 + 110)
	for _, upkElm := range upkErr {
		guessCap += len(upkElm.PartialStack) * 160
		if includeSource {
			guessCap += len(upkErr) * 124
		}
	}
	builder.Grow(guessCap)

	// get source code if available
	var sourceLines [][]SourceLines
	if includeSource {
		var err error
		sourceLines, err = unpacker.GetSourceLines(0, 120, true)
		if err != nil {
			includeSource = false
		}
	}

	for i := len(upkErr) - 1; i >= 0; i-- {
		upkElm := upkErr[i]
		if len(upkElm.PartialStack) > 0 {
			builder.WriteString("Traceback (most recent call last):")
			for j := len(upkElm.PartialStack) - 1; j >= 0; j-- {
				s := upkElm.PartialStack[j]
				builder.WriteString("\n  File \"")
				builder.WriteString(s.File)
				builder.WriteString("\", line ")
				builder.WriteInt(int64(s.Line))
				builder.WriteString(", in ")
				builder.WriteString(s.Name)
				if includeSource {
					builder.WriteString("\n    ")
					builder.WriteString(sourceLines[i][j][0].Source)
				}
			}
			builder.WriteByte('\n')
		}
		builder.WriteString(typeName(upkElm.Err))
		if upkElm.Msg != "" {
			builder.WriteString(": ")
			builder.WriteString(upkElm.Msg)
		}

		if i > 0 {
			builder.WriteString(
				"\n\nThe above exception was the direct cause of the following exception:\n\n",
			)
		}
	}
	return builder.Bytes()
}
