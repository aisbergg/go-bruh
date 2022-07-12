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
		formatted = append(formatted, f.String())
	}
	return strings.Join(formatted, "\n")
}

// format returns an array of formatted stack frames.
func (s Stack) format(sep string, invert bool) []string {
	var str []string
	for _, f := range s {
		if invert {
			str = append(str, f.format(sep))
		} else {
			str = append([]string{f.format(sep)}, str...)
		}
	}
	return str
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

// String returns a string representation of the StackFrame formatted similar to
// Go's runtime stack.
func (f *StackFrame) String() string {
	return fmt.Sprintf("%s\n        %s:%d pc=0x%x", f.Name, f.File, f.Line, f.ProgramCounter)
}

// format returns a formatted stack frame.
func (f *StackFrame) format(sep string) string {
	return fmt.Sprintf("%v%v%v%v%v", f.Name, sep, f.File, sep, f.Line)
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
	const depth = 64
	var pcs [depth]uintptr
	n := runtime.Callers(int(skip), pcs[:])
	var st stackPC = pcs[0 : n-2]
	return st
}

// stackPC is an array of program counters.
type stackPC []uintptr

func relativeStack(prvStack, nxtStack Stack) Stack {
	prvInd := len(prvStack) - 1
	curInd := len(nxtStack) - 1

	for {
		if prvInd < 0 || curInd < 0 {
			break
		}
		if prvStack[prvInd] != nxtStack[curInd] {
			break
		}
		prvInd--
		curInd--
	}

	return nxtStack[:curInd+1]
}

// get returns a human readable stack trace.
func (s stackPC) get() Stack {
	var stackFrames Stack

	frames := runtime.CallersFrames(s)
	for {
		frame, more := frames.Next()
		i := strings.LastIndex(frame.Function, "/")
		name := frame.Function[i+1:]
		stackFrames = append(stackFrames, StackFrame{
			Name:           name,
			File:           frame.File,
			Line:           frame.Line,
			ProgramCounter: frame.PC,
		})
		if !more {
			break
		}
	}

	return stackFrames
}

// isGlobal determines if the stack trace represents a globally defined error.
func (s stackPC) isGlobal() bool {
	frames := s.get()
	for _, f := range frames {
		if strings.ToLower(f.Name) == "runtime.doinit" {
			return true
		}
	}
	return false
}
