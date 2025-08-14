package bruh

import (
	"strings"

	"github.com/aisbergg/go-bruh/internal/color"
	"github.com/aisbergg/go-bruh/internal/stringbuilder"
	"github.com/aisbergg/go-bruh/internal/util"
)

// BruhFormatter is the default error formatter that produces a single error
// message with a combined verbose stack trace that includes the function and
// the location of each stack frame. Most recent calls are at the top.
//
// # Output Format
//
//	errorMsg1: errorMsg2: externalErrorMsg
//	    at function1 (file1:line1)
//	    at function2 (file2:line2)
//	    at functionN (fileN:lineN)
func BruhFormatter(b []byte, unpacker *Unpacker) []byte { //nolint:revive
	return formatBruhSourced(b, unpacker, false, false)
}

// BruhFancyFormatter returns a [Formatter] that produces a single error
// message with a combined verbose stack trace, including the function name and
// location of each stack frame. Most recent calls are at the top. Optional
// coloring and source code snippets can be enabled.
//
// Parameters:
//   - colored: enables ANSI-colored output (default: false)
//   - sourced: enables inclusion of source code snippets from the current working directory, if available (default: false)
//
// # Output Format
//
// Standard Format:
//
//	errorMsg1: errorMsg2: externalErrorMsg
//	    at function1 (file1:line1)
//	    at function2 (file2:line2)
//	    at functionN (fileN:lineN)
//
// With "sourced" enabled and source code available:
//
//	errorMsg1: errorMsg2: externalErrorMsg
//
//	at function1 (file1:line1)
//	    162│ // foo returns an error
//	    163│ func foo() error {
//	  → 164│    return bruh.Wrap(ErrExternal, "errorMsg1")
//	    165│ }
//	    166│
//	at functionN (fileN:lineN)
//	    162│ // bar returns an error
//	    163│ func bar() error {
//	  → 164│    return bruh.Wrap(foo(), "errorMsg2")
//	    165│ }
//	    166│
func BruhFancyFormatter(colored, sourced bool) Formatter { //nolint:revive
	return func(b []byte, unpacker *Unpacker) []byte {
		return formatBruhSourced(b, unpacker, colored, sourced)
	}
}

// BruhStackedFormatter is an error formatter that produces verbose error traces
// including the function and the location of each stack frame. The format is
// pretty close to the format produced by Java or JavaScript. Most recent calls
// are at the top.
//
// # Output Format
//
//	errorMsg1
//	    at function1 (file1:line1)
//	    at function2 (file2:line2)
//	    at functionN (fileN:lineN)
//	errorMsg2
//	    at function1 (file1:line1)
//	    at function2 (file2:line2)
//	    at functionN (fileN:lineN)
//	externalErrorMsg
func BruhStackedFormatter(b []byte, unpacker *Unpacker) []byte { //nolint:revive
	return formatBruhStacked(b, unpacker, false, false, false)
}

// BruhStackedFancyFormatter returns a [Formatter] that produces verbose error
// traces including the function and location of each stack frame, with optional
// coloring, source code snippets, and type annotations.
//
// Parameters:
//   - colored: enables ANSI-colored output (default: false)
//   - sourced: enables inclusion of source code snippets from the current working directory, if available (default: false)
//   - typed: enables inclusion of error type annotations (default: false)
//
// # Output Format
//
// Standard Format:
//
//	errorMsg1
//	    at function1 (file1:line1)
//	    at function2 (file2:line2)
//	    at functionN (fileN:lineN)
//	errorMsg2
//	    at function1 (file1:line1)
//	    at function2 (file2:line2)
//	    at functionN (fileN:lineN)
//	externalErrorMsg
//
// With "typed" enabled:
//
//	typeName1: errorMsg1
//	    at function1 (file1:line1)
//	    at function2 (file2:line2)
//	    at functionN (fileN:lineN)
//	typeName2: errorMsg2
//	    at function1 (file1:line1)
//	    at function2 (file2:line2)
//	    at functionN (fileN:lineN)
//	typeNameN: externalErrorMsg
//
// With "sourced" enabled and source code available:
//
//	errorMsg1
//
//	at function1 (file1:line1)
//	    162│ // foo returns an error
//	    163│ func foo() error {
//	  → 164│ 	return bruh.New("foo")
//	    165│ }
//	    166│
//	at functionN (fileN:lineN)
//	    162│ // bar returns an error
//	    163│ func bar() error {
//	  → 164│ 	return bruh.Wrap(foo(), "bar")
//	    165│ }
//	    166│
//
//	errorMsg2
//
//	at function1 (file1:line1)
//	    162│ // foo returns an error
//	    163│ func foo() error {
//	  → 164│ 	return bruh.New("foo")
//	    165│ }
//	    166│
//	at functionN (fileN:lineN)
//	    162│ // bar returns an error
//	    163│ func bar() error {
//	  → 164│ 	return bruh.Wrap(foo(), "bar")
//	    165│ }
//	    166│
//
//	externalErrorMsg
func BruhStackedFancyFormatter(colored, sourced, typed bool) Formatter { //nolint:revive
	return func(b []byte, unpacker *Unpacker) []byte {
		return formatBruhStacked(b, unpacker, colored, sourced, typed)
	}
}

// -----------------------------------------------------------------------------

func formatBruhStacked(b []byte, unpacker *Unpacker, colored, sourced, typed bool) []byte {
	if unpacker.Error() == nil {
		return b
	}
	upkErr := unpacker.Unpack()

	var sourceLines [][]SourceLines
	var errSourceLines error
	if sourced {
		sourceLines, errSourceLines = unpacker.GetSourceLines(2, 120, true)
	}

	// allocate a large buffer to avoid later reallocations.
	// type and message: 120 per error
	// location: 160 per location
	// source line: 250 per location
	builder := stringbuilder.New(b)
	guessCap := len(upkErr) * 120
	for _, upkElm := range upkErr {
		guessCap += len(upkElm.PartialStack) * 160
		if sourced && errSourceLines == nil {
			guessCap += len(upkElm.PartialStack) * 250
		}
	}
	builder.Grow(guessCap)
	colorer := color.NewColorer(builder, colored)

	lastIndex := len(upkErr) - 1
	for i, upkElm := range upkErr {
		msg := upkElm.Msg
		if msg == "" {
			msg = "<no message>"
		}
		if typed {
			colorer.ColoredText(typeName(upkElm.Err), color.Bold, color.BrightRed)
			colorer.Color(color.Bold)
			builder.WriteString(": ")
			builder.WriteString(msg)
			colorer.Reset()
		} else {
			colorer.ColoredText(msg, color.Bold, color.BrightRed)
		}
		if sourced && errSourceLines == nil {
			if len(upkElm.PartialStack) > 0 {
				builder.WriteString("\n")
			}
			for j, s := range upkElm.PartialStack {
				formatSingleStackWithSourceCode(s, sourceLines[i][j], builder, colorer)
			}
		} else {
			for _, s := range upkElm.PartialStack {
				builder.WriteString("\n    at ")
				colorer.ColoredText(s.Name, color.BrightCyan)
				builder.WriteString(" (")
				colorer.ColoredText(s.File, color.BrightGreen)
				builder.WriteByte(':')
				builder.WriteInt(int64(s.Line))
				builder.WriteByte(')')
			}
		}
		if i < lastIndex {
			builder.WriteString("\n")
		}
	}
	return builder.Bytes()
}

func formatBruhSourced(b []byte, unpacker *Unpacker, colored, sourced bool) []byte {
	if unpacker.Error() == nil {
		return b
	}
	stack := unpacker.CombinedStack()

	var sourceLines []SourceLines
	var errSourceLines error
	if sourced {
		sourceLines, errSourceLines = stack.GetSourceLines(2, 120, true)
	}

	// allocate a large buffer to avoid later reallocations
	// message: 80 per error
	// location: 160 per location
	// source line: 250 per location
	builder := stringbuilder.New(b)
	guessCap := unpacker.ChainLen() * 80
	if len(stack) != 0 {
		guessCap += len(stack) * 160
		if sourced && errSourceLines == nil {
			guessCap += len(stack) * 250
		}
	}
	builder.Grow(guessCap)
	colorer := color.NewColorer(builder, colored)
	colorer.Color(color.Bold, color.BrightRed)
	emptyLen := builder.Len()
	unpacker.AppendMessageBuilder(builder)
	if builder.Len() == emptyLen {
		// If we got no message we want to state that. I think this is better
		// than the alternatives of writing nothing, presenting the user a
		// generic 'error' message or breaking the layout of the formatted
		// error.
		builder.WriteString("<no message>")
	}
	colorer.Reset()

	if len(stack) != 0 {
		if sourced && errSourceLines == nil {
			builder.WriteByte('\n')
			for i, s := range stack {
				formatSingleStackWithSourceCode(s, sourceLines[i], builder, colorer)
			}
		} else {
			for _, s := range stack {
				builder.WriteString("\n    at ")
				colorer.ColoredText(s.Name, color.BrightCyan)
				builder.WriteString(" (")
				colorer.ColoredText(s.File, color.BrightGreen)
				builder.WriteByte(':')
				builder.WriteInt(int64(s.Line))
				builder.WriteByte(')')
			}
		}
	}

	return builder.Bytes()
}

func formatSingleStackWithSourceCode(
	s StackFrame,
	sourceLines SourceLines,
	builder *stringbuilder.StringBuilder,
	colorer color.Colorer,
) {
	builder.WriteString("\nat ")
	colorer.ColoredText(s.Name, color.BrightCyan)
	builder.WriteString(" (")
	colorer.ColoredText(s.File, color.BrightGreen)
	builder.WriteByte(':')
	builder.WriteInt(int64(s.Line))
	builder.WriteByte(')')
	// add source code
	numDigits := util.DigitsInNumber(max(
		sourceLines[0].LineNum,
		sourceLines[1].LineNum,
		sourceLines[2].LineNum,
		sourceLines[3].LineNum,
		sourceLines[4].LineNum,
	))
	for k := 0; k < 5; k++ {
		lineNum := sourceLines[k].LineNum
		if sourceLines[k].LineNum >= 0 {
			if k == 2 {
				builder.WriteString("\n  ")
				colorer.ColoredText("→", color.BrightRed)
				builder.WriteByte(' ')
			} else {
				builder.WriteString("\n    ")
			}
			for l := 0; l < numDigits-util.DigitsInNumber(lineNum); l++ {
				builder.WriteByte(' ')
			}
			if k == 2 {
				colorer.ColoredInt(int64(lineNum), color.Bold)
			} else {
				builder.WriteInt(int64(lineNum))
			}
			builder.WriteString("│    ")
			source := sourceLines[k].Source
			// replace tabs with spaces
			source = strings.ReplaceAll(source, "\t", "    ")
			if k == 2 {
				builder.WriteString(source)
			} else {
				colorer.ColoredText(source, color.Faint)
			}
		}
	}
}
