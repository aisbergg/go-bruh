package bruh

import (
	"fmt"
	"runtime"
	"strings"
)

// Stack is an array of stack frames stored in a human readable format.
type Stack []StackFrame

func (s Stack) String() string {
	formatted := make([]string, 0, len(s))
	for _, f := range s {
		ftdFrm := fmt.Sprintf("%s\n        %s:%d pc=0x%x", f.Name, f.File, f.Line, f.ProgramCounter)
		formatted = append(formatted, ftdFrm)
	}
	return strings.Join(formatted, "\n")
}

// RelativeTo returns new version of this stack relative to the other stack.
func (s Stack) RelativeTo(other Stack) Stack {
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
	// ProgramCounter is the underlying program counter for the function.
	ProgramCounter uintptr
}

// callers returns a stack trace. the argument skip is the number of stack
// frames to skip before recording in pc, with 0 identifying the frame for
// Callers itself and 1 identifying the caller of Callers.
func callers(skip uint) stackPC {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(int(skip), pcs[:])
	var st stackPC = pcs[0:n]
	return st
}

// stackPC is an array of program counters.
type stackPC []uintptr

// relativeTo returns new version of this stack relative to the other stack.
func (s stackPC) relativeTo(other stackPC) stackPC {
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
			Name: frame.Function,
			File: frame.File,
			Line: frame.Line,
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
