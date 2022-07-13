package errors

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

// framePC is a single program counter of a stack framePC.
type framePC uintptr

// pc returns the program counter for a frame.
func (f framePC) pc() uintptr {
	return uintptr(f) - 1
}

// get returns a human readable stack frame.
func (f framePC) get() StackFrame {
	pc := f.pc()
	frames := runtime.CallersFrames([]uintptr{pc})
	frame, _ := frames.Next()

	i := strings.LastIndex(frame.Function, "/")
	name := frame.Function[i+1:]

	return StackFrame{
		Name:           name,
		File:           frame.File,
		Line:           frame.Line,
		ProgramCounter: pc,
	}
}

// callers returns a stack trace. the argument skip is the number of stack
// frames to skip before recording in pc, with 0 identifying the frame for
// Callers itself and 1 identifying the caller of Callers.
func callers(skip uint) stackPC {
	pcs := make([]uintptr, 64)
	n := runtime.Callers(int(skip), pcs[:])
	var st stackPC = pcs[0:n]
	return st
}

// stackPC is an array of program counters.
type stackPC []uintptr

// RelativeTo returns new version of this stack relative to the other stack.
func (s stackPC) RelativeTo(other stackPC) stackPC {
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
func (s stackPC) toStack() (stack Stack) {
	if len(s) == 0 {
		return
	}
	frames := runtime.CallersFrames(s)
	for {
		frame, more := frames.Next()
		// exclude runtime calls
		if strings.Contains(frame.File, "runtime/") {
			break
		}
		stack = append(stack, StackFrame{
			Name:           frame.Function,
			File:           frame.File,
			Line:           frame.Line,
			ProgramCounter: frame.PC,
		})
		if !more {
			break
		}
	}
	return
}

// isGlobal determines if the stack trace represents a globally defined error.
func (s stackPC) isGlobal() bool {
	frames := runtime.CallersFrames(s)
	for {
		frame, more := frames.Next()
		if frame.Function == "runtime.doInit" {
			return true
		}
		if !more {
			break
		}
	}
	return false
}
