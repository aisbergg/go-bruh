package bruh

import (
	"reflect"
	"strconv"
	"strings"
)

// ToString returns a default formatted string for a given error.
func ToString(err error, withTrace bool) string {
	if withTrace {
		return ToCustomString(err, FormatWithTrace)
	}
	return ToCustomString(err, nil)
}

// ToCustomString returns a formatted string for the given error and formatter.
func ToCustomString(err error, f Formatter) string {
	if err == nil {
		return ""
	}

	// Format error without a stack trace by simply combining all error
	// messages. (fast path)
	if f == nil {
		return formatWithoutTrace(err)
	}

	// unpack error and apply formatter
	upkErr := Unpack(err, false)
	return f(upkErr)
}

// Formatter turns an unpacked error into a formatted string.
type Formatter func(UnpackedError) string

// FormatWithoutTrace is an error formatter that produces a combined error
// message of all wrapped errors without a stack trace. (just kept for
// reference)
//
// Format:
//
//	<error1>: <error2>: <errorN>
func FormatWithoutTrace(upkErr UnpackedError) string {
	strBld := strings.Builder{}
	strBld.WriteString(upkErr[0].Msg)
	for i := 1; i < len(upkErr); i++ {
		strBld.WriteString(": ")
		strBld.WriteString(upkErr[i].Msg)
	}
	return strBld.String()
}

// formatWithoutTrace formats an error without a stack trace by simply combining
// all error messages.
//
// Format:
//
//	<error1>: <error2>: <errorN>
func formatWithoutTrace(err error) string {
	strBld := strings.Builder{}
	// pre-allocate a large buffer to avoid reallocations
	strBld.Grow(1024)
	if e, ok := err.(interface{ Message() string }); ok {
		strBld.WriteString(e.Message())
	} else {
		strBld.WriteString(err.Error())
		return strBld.String()
	}
	err = Unwrap(err)
	for err != nil {
		if e, ok := err.(interface{ Message() string }); ok {
			msg := e.Message()
			if msg != "" {
				strBld.WriteString(": ")
				strBld.WriteString(msg)
			}
		} else {
			msg := err.Error()
			if msg != "" {
				strBld.WriteString(": ")
				strBld.WriteString(msg)
			}
			break
		}
		err = Unwrap(err)
	}
	return strBld.String()
}

// FormatWithTrace is an error formatter that produces a trace containing a
// partial stack trace for each wrapped error.
//
// Format:
//
//	<error1>:
//	    <file1>:<line1> in <function1>
//	    <file2>:<line2> in <function2>
//	    <fileN>:<lineN> in <functionN>
//	<error2>:
//	    <file1>:<line1> in <function1>
//	    <file2>:<line2> in <function2>
//	    <fileN>:<lineN> in <functionN>
//	<errorN>:
//	    <file1>:<line1> in <function1>
//	    <file2>:<line2> in <function2>
//	    <fileN>:<lineN> in <functionN>
func FormatWithTrace(upkErr UnpackedError) string {
	strBld := strings.Builder{}
	// pre-allocate a large buffer to avoid reallocations; some guesswork here:
	// message: 80 per error
	// location: 160 per error
	guessCap := len(upkErr) * 80
	for _, upkElm := range upkErr {
		guessCap += len(upkElm.Stack) * 160
	}
	strBld.Grow(guessCap)

	for i, upkElm := range upkErr {
		if upkElm.Msg == "" {
			strBld.WriteString(`""`)
		} else {
			strBld.WriteString(upkElm.Msg)
		}
		if upkElm.Stack == nil {
			continue
		}
		strBld.WriteRune('\n')
		for j, s := range upkElm.Stack {
			strBld.WriteString("    ")
			strBld.WriteString(s.File)
			strBld.WriteRune(':')
			strBld.WriteString(strconv.Itoa(s.Line))
			strBld.WriteString(" in ")
			strBld.WriteString(s.Name)
			if j < len(upkElm.Stack)-1 {
				strBld.WriteRune('\n')
			}
		}
		if i < len(upkErr)-1 {
			strBld.WriteRune('\n')
		}
	}
	return strBld.String()
}

// FormatWithCombinedTrace is an error formatter that produces a single
// combined stack trace.
//
// Format:
//
//	<error1>: <error2>: <errorN>
//	    <file1>:<line1> in <function1>
//	    <file2>:<line2> in <function2>
//	    <fileN>:<lineN> in <functionN>
func FormatWithCombinedTrace(upkErr UnpackedError) string {
	strBld := strings.Builder{}
	if len(upkErr) == 1 && upkErr[0].Msg == "" {
		strBld.WriteString(`""`)
	} else {
		strBld.WriteString(upkErr[0].Msg)
		for i := 1; i < len(upkErr); i++ {
			strBld.WriteString(": ")
			strBld.WriteString(upkErr[i].Msg)
		}
	}
	upkElm := upkErr[0]
	if terr, ok := upkElm.Err.(interface{ FullStack() Stack }); ok {
		fStk := terr.FullStack()
		strBld.WriteRune('\n')
		for i, s := range fStk {
			strBld.WriteString("    ")
			strBld.WriteString(s.File)
			strBld.WriteRune(':')
			strBld.WriteString(strconv.Itoa(s.Line))
			strBld.WriteString(" in ")
			strBld.WriteString(s.Name)
			if i < len(fStk)-1 {
				strBld.WriteRune('\n')
			}
		}
	}
	return strBld.String()
}

// FormatPythonTraceback is an error formatter that produces error traces
// similar to Python's tracebacks.
//
// Format:
//
//	Traceback (most recent call last):
//	  File "<file3>", line <line3>, in <function3>
//	  File "<file2>", line <line2>, in <function2>
//	  File "<file1>", line <line1>, in <function1>
//	<typeName1>: <error1>
//
//	The above exception was the direct cause of the following exception:
//
//	Traceback (most recent call last):
//	  File "<file3>", line <line3>, in <function3>
//	  File "<file2>", line <line2>, in <function2>
//	  File "<file1>", line <line1>, in <function1>
//	<typeName2>: <error2>
func FormatPythonTraceback(upkErr UnpackedError) string {
	strBld := strings.Builder{}
	for i := len(upkErr) - 1; i >= 0; i-- {
		upkElm := upkErr[i]
		if upkElm.PartialStack != nil {
			strBld.WriteString("Traceback (most recent call last):\n")
			for j := len(upkElm.PartialStack) - 1; j >= 0; j-- {
				s := upkElm.PartialStack[j]
				strBld.WriteString("  File \"")
				strBld.WriteString(s.File)
				strBld.WriteString("\", line ")
				strBld.WriteString(strconv.Itoa(s.Line))
				strBld.WriteString(", in ")
				strBld.WriteString(s.Name)
				strBld.WriteRune('\n')
			}
		}
		strBld.WriteString(reflect.TypeOf(upkElm.Err).String())
		if upkElm.Msg != "" {
			strBld.WriteString(": ")
			strBld.WriteString(upkElm.Msg)
		}

		if i > 0 {
			strBld.WriteString("\n\nThe above exception was the direct cause of the following exception:\n\n")
		}
	}
	return strBld.String()
}

// Unpack returns a human-readable UnpackedError type for a given error.
func Unpack(err error, unwrapExternal bool) UnpackedError {
	upkErr := make([]UnpackedElement, 0, 32)
	prvStack := Stack{}

	for err != nil {
		// check if it behaves like a base error
		if e, ok := err.(interface {
			Message() string
			Stack() Stack
		}); ok {
			stack := e.Stack()
			upkErr = append(upkErr, UnpackedElement{
				Err:          err,
				Msg:          e.Message(),
				Stack:        stack,
				PartialStack: stack.RelativeTo(prvStack),
			})
			prvStack = stack

		} else {
			// continue unwrapping external errors
			upkErr = append(upkErr, UnpackedElement{
				Err: err,
				Msg: err.Error(),
			})
			if !unwrapExternal {
				break
			}
		}

		err = Unwrap(err)
	}

	return upkErr
}

// UnpackedError represents an unpacked error which is quite useful for
// formatting purposes and other error processing. Use Unpack() to unpack any
// kind of error that supports it.
type UnpackedError []UnpackedElement

// UnpackedElement represents a single error frame and the accompanying message.
type UnpackedElement struct {
	// Err is the error instance.
	Err error
	// Msg is the message contained in the error.
	Msg string
	// Stack is the error stack for this particular error instance.
	Stack Stack
	// PartialStack is the error stack with parts cut off that are already in
	// the previous error stack.
	PartialStack Stack
}
