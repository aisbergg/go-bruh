package bruh

import (
	"runtime"
	"strconv"
	"strings"
)

// Stack is an array of stack frames stored in a human readable format.
type Stack []StackFrame

func (s Stack) String() string {
	strBld := strings.Builder{}
	// pre-allocate a large buffer to avoid reallocations; some guesswork here:
	// Name: 40 per error
	// Location: 160 per error
	strBld.Grow(len(s) * (40 + 160))
	for _, f := range s {
		strBld.WriteString(f.Name)
		strBld.WriteString("\n        ")
		strBld.WriteString(f.File)
		strBld.WriteRune(':')
		strBld.WriteString(strconv.Itoa(f.Line))
		strBld.WriteString(" pc=0x")
		strBld.WriteString(strconv.FormatInt(int64(f.ProgramCounter2), 16)) // format as hex
		strBld.WriteRune('\n')
	}
	return strBld.String()
}

// RelativeTo returns new version of this stack relative to the other stack.
func (s Stack) RelativeTo(other Stack) Stack {
	if len(s) == 0 || len(other) == 0 {
		return s
	}
	othInd := len(other) - 1
	curInd := len(s) - 1
	for {
		if othInd < 0 || curInd < 0 {
			break
		}
		if other[othInd] != s[curInd] {
			break
		}
		othInd--
		curInd--
	}
	// Due to optimizations, the frames of the two stacks may be identical. In
	// that case, we would end up with an empty stack, which is not what we
	// want.
	if curInd < 0 {
		return s[:1]
	}
	return s[:curInd+1]
}

// First returns the first x stack frames in the stack.
func (s Stack) First(x int) Stack {
	if len(s) <= x {
		return s
	}
	return s[:x]
}

// Last returns the last x stack frames in the stack.
func (s Stack) Last(x int) Stack {
	if len(s) <= x {
		return s
	}
	return s[len(s)-x:]
}

// StackFrame stores a frame's runtime information in a human readable format.
type StackFrame struct {
	// Name of the function.
	Name string
	// File path where the function is defined.
	File string
	// Line number where the function is defined.
	Line int
	// ProgramCounter, obtained from [runtime.Callers], indicates the starting
	// point of the previous instruction before our instruction of interest.
	// Apperently this is done for historical reasons. You can use its value
	// with [runtime.CallersFrames] to look up the corresponding symbolic
	// information of the function (as done by Sentry for example).
	ProgramCounter uintptr
	// ProgramCounter2, obtained from [runtime.CallersFrames], indicates the
	// starting point of the instruction in question. However, if your goal is to
	// retrieve the associated function, it is recommended to utilize
	// ProgramCounter instead. ProgramCounter appears to offer greater
	// reliability in conjunction with [runtime.CallersFrames].
	ProgramCounter2 uintptr
}

// callers returns a stack trace. the argument skip is the number of stack
// frames to skip before recording in pc, with 0 identifying the frame for
// Callers itself and 1 identifying the caller of Callers.
func callers(skip uint) stackPC {
	var pcs [32]uintptr
	n := runtime.Callers(int(skip), pcs[:])
	var st stackPC = pcs[0:n]
	return st
}

// stackPC is an array of program counters.
type stackPC []uintptr

// relativeTo returns new version of this stack relative to the other stack.
func (s stackPC) relativeTo(other stackPC) stackPC {
	if len(s) == 0 || len(other) == 0 {
		return s
	}
	othInd := len(other) - 1
	curInd := len(s) - 1
	for {
		if othInd < 0 || curInd < 0 {
			break
		}
		if other[othInd] != s[curInd] {
			break
		}
		othInd--
		curInd--
	}
	// Due to optimizations, the frames of the two stacks may be identical. In
	// that case, we would end up with an empty stack, which is not what we
	// want.
	if curInd < 0 {
		return s[:1]
	}
	return s[:curInd+1]
}

// toStack returns a Stack object with information details about the PCs.
func (s stackPC) toStack() Stack {
	if len(s) == 0 {
		return Stack{}
	}
	stack := make(Stack, 0, len(s))
	frames := runtime.CallersFrames(s)
	var i int
	for {
		frame, more := frames.Next()
		// discard stack for globally defined errors
		if frame.Function == "runtime.doInit" {
			return Stack{}
		}
		// exclude runtime calls
		if strings.Contains(frame.File, "runtime/") {
			break
		}
		stack = append(stack, StackFrame{
			Name:            frame.Function,
			File:            frame.File,
			Line:            frame.Line,
			ProgramCounter2: frame.PC,
			// CallersFrames.Next() reduces the program counter by 1. Using the
			// reduced program counter in subsequent calls of CallersFrames
			// would lead to wrong frames being returned, which happens in
			// Sentry for example. Therefore we use the original program counter
			// without the reduction.
			ProgramCounter: s[i],
		})
		if !more {
			break
		}
		i++
	}
	return stack
}
