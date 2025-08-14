package bruh

import (
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"github.com/aisbergg/go-bruh/internal/stringbuilder"
)

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

// Stack is an array of stack frames stored in a human readable format.
type Stack []StackFrame

func (s Stack) String() string {
	// allocate a large buffer to avoid later reallocations
	// Name: 40 per error
	// Location: 160 per location
	builder := stringbuilder.New([]byte{})
	builder.Grow(len(s) * (40 + 160))
	for i, f := range s {
		builder.WriteString(f.Name)
		builder.WriteString("\n\t")
		builder.WriteString(f.File)
		builder.WriteByte(':')
		builder.WriteInt(int64(f.Line))
		builder.WriteString(" pc=0x")
		builder.WriteIntAsHex(int64(f.ProgramCounter2))
		if i < len(s)-1 {
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}

// RelativeTo returns the stack that is relative to the other stack. It uses the
// original underlying buffer, so do not change its contents!
func (s Stack) RelativeTo(other Stack) Stack {
	if len(s) == 0 || len(other) == 0 {
		return s
	}
	othIdx := len(other) - 1
	curIdx := len(s) - 1
	// find first common index in case the captured stack got truncated when its
	// size exceeded MAX_STACK_DEPTH
	for othIdx >= 0 && curIdx >= 0 &&
		other[othIdx].ProgramCounter2 != s[curIdx].ProgramCounter2 {
		othIdx--
	}
	// find last common index
	for othIdx >= 0 && curIdx >= 0 &&
		other[othIdx].ProgramCounter2 == s[curIdx].ProgramCounter2 {
		othIdx--
		curIdx--
	}
	// Due to optimizations, the frames of the two stacks may be identical. In
	// that case, we would end up with an empty stack, which is not what we
	// want.
	if curIdx < 0 {
		return s[:1]
	}
	return s[:curIdx+1]
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

// GetSourceLines returns the source lines for the given stack. The output is in
// the same order as the stack frames (`[stackIdx]SourceLines`). If the source
// code is not available, an error is returned. ctxLines is the number of lines
// before and after the requested lines that should be included. colCap is the
// maximum number of characters per line. If unindent is true, the source lines
// are unindented.
func (s Stack) GetSourceLines(ctxLines, colCap int, unindent bool) ([]SourceLines, error) {
	// create a list of files and the source lines we want to read from them
	linesInFiles := make(map[string][]int, len(s))
	for _, sf := range s {
		if _, ok := linesInFiles[sf.File]; !ok {
			linesInFiles[sf.File] = make([]int, 0, 8) // guess 8 lines per file
		}
		linesInFiles[sf.File] = append(linesInFiles[sf.File], sf.Line)
	}

	type fileLine struct {
		file string
		line int
	}

	// create a mapping from (file, line) combination to source lines
	sourcesInFileLine := make(map[fileLine]SourceLines, len(s))
	for file, lineNums := range linesInFiles {
		sourceLines, err := getSourceLinesFromFile(file, lineNums, ctxLines, colCap, unindent)
		if err != nil {
			return nil, err
		}
		for i, l := range lineNums {
			sourcesInFileLine[fileLine{file, l}] = sourceLines[i]
		}
	}

	// create the source lines data structure in the same order as the errors and stack frames
	sourceLines := make([]SourceLines, 0, len(s))
	for _, sf := range s {
		sourceLines = append(sourceLines, sourcesInFileLine[fileLine{sf.File, sf.Line}])
	}

	return sourceLines, nil
}

var stack4xPool = sync.Pool{
	New: func() any {
		stack := Stack(make([]StackFrame, 4*MAX_STACK_DEPTH))
		return &stack
	},
}

// allocStack allocates a new [Stack] or grabs a cached one.
func new4xStack() *Stack {
	return stack4xPool.Get().(*Stack) //nolint:revive
}

// disposeStack puts the stack back into the pool.
func dispose4xStack(stack *Stack) {
	*stack = (*stack)[:cap(*stack)]
	stack4xPool.Put(stack)
}

// -----------------------------------------------------------------------------

// framesDouble acts as a double for [runtime.Frames] and allows us to reset the
// private fields of [runtime.Frames].
type framesDouble struct {
	callers    []uintptr
	nextPC     uintptr
	frames     []runtime.Frame
	frameStore [2]runtime.Frame
}

func asRuntimeFrames(f *framesDouble) *runtime.Frames {
	return (*runtime.Frames)(unsafe.Pointer(f)) //nolint:gosec
}

func asFramesDouble(f *runtime.Frames) *framesDouble {
	return (*framesDouble)(unsafe.Pointer(f)) //nolint:gosec
}

var framesDoublePool = sync.Pool{
	New: func() any { return &framesDouble{} },
}

func allocFrames(callers []uintptr) *runtime.Frames {
	f := framesDoublePool.Get().(*framesDouble) //nolint:revive
	f.callers = callers
	f.frames = f.frameStore[:0]
	return asRuntimeFrames(f)
}

func disposeFrames(f *runtime.Frames) {
	framesDoublePool.Put(asFramesDouble(f))
}

// -----------------------------------------------------------------------------

// stackPC is an array of program counters.
type stackPC []uintptr

// relativeTo returns new version of this stack relative to the other stack.
func (s stackPC) relativeTo(other stackPC) stackPC {
	if len(s) == 0 || len(other) == 0 {
		return s
	}
	othIdx := len(other) - 1
	curIdx := len(s) - 1
	// find first common index (only necessary when the stack depth is larger
	// than the configured MAX_STACK_DEPTH)
	for othIdx >= 0 && curIdx >= 0 &&
		other[othIdx] != s[curIdx] {
		othIdx--
	}
	// find last common index
	for othIdx >= 0 && curIdx >= 0 &&
		other[othIdx] == s[curIdx] {
		othIdx--
		curIdx--
	}
	// Due to optimizations, the frames of the two stacks may be identical. In
	// that case, we would end up with an empty stack, which is not what we
	// want.
	if curIdx < 0 {
		return s[:1]
	}
	return s[:curIdx+1]
}

// toStack fills the slice with information details about the program counters.
// It returns the number of entries written to stack.
func (s stackPC) toStack(stack Stack) int {
	if len(s) == 0 {
		return 0
	}
	frames := allocFrames(s)
	var i int
	// we don't want to exceed the number of allocated stack frames
	for i < len(stack) {
		frame, more := frames.Next()
		// discard stack for globally defined errors
		if isGloballyDefinedError(frame.Function) {
			disposeFrames(frames)
			return 0
		}
		// exclude runtime calls
		if strings.Contains(frame.File, "runtime/") {
			break
		}
		stack[i] = StackFrame{
			Name:            frame.Function,
			File:            frame.File,
			Line:            frame.Line,
			ProgramCounter2: frame.PC,
			// CallersFrames reduces the program counter by 1. Using the
			// reduced program counter in subsequent calls of CallersFrames
			// would lead to wrong frames being returned, which happens in
			// Sentry for example. Therefore we use the original program counter
			// without the reduction.
			ProgramCounter: s[i],
		}
		i++
		if !more {
			break
		}
	}
	disposeFrames(frames)
	return i
}

// isGloballyDefinedError returns true if the function is a globally defined,
// meaning it is not created on the spot.
func isGloballyDefinedError(f string) bool {
	return strings.HasPrefix(f, "runtime.doInit") || strings.HasSuffix(f, ".init")
}
