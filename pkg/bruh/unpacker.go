package bruh

import (
	"strings"
	"sync"
)

// Unpacker holds information about an error chain and provides methods to
// unpack and process it.
//
// IMPORTANT: Instances of Unpacker and its produced objects are recycled and
// reused; do NOT retain references to them outside their intended scope.
type Unpacker struct {
	err       error          // The root error in the chain.
	upkErr    *UnpackedError // A pointer to the unpacked error representation.
	cbdStk    *Stack         // A pointer to the combined stack of all errors.
	chainLen  int            // The length of the error chain.
	unpackAll bool           // Indicates whether errors without a trace should get a separate entry in upkErr or shall be "pooled" together.
}

// unpackerPool is a sync.Pool for reusing Unpacker instances to reduce allocations.
var unpackerPool = sync.Pool{
	New: func() any { return &Unpacker{} },
}

// newUnpacker creates a new instance of *Unpacker. It calculates the length of the error chain
// and retrieves a previously allocated Unpacker from the pool if available, initializing it with
// the provided error and unpackAll flag.
func newUnpacker(err error, unpackAll bool) *Unpacker {
	chainLen := 0
	for uerr := err; uerr != nil; uerr = Unwrap(uerr) {
		chainLen++
	}
	unpacker := unpackerPool.Get().(*Unpacker) //nolint:revive
	*unpacker = Unpacker{
		err:       err,
		chainLen:  chainLen,
		unpackAll: unpackAll,
	}
	return unpacker
}

// disposeUnpacker disposes the [*Unpacker] instance and makes it available for
// reuse later. It resets the internal fields of the Unpacker to their zero values
// and returns the instance to the unpackerPool.
func disposeUnpacker(unpacker *Unpacker) {
	if unpacker.cbdStk != nil {
		stack := unpacker.cbdStk
		unpacker.cbdStk = nil
		disposeChainStack(stack)
	}
	if unpacker.upkErr != nil {
		upkErr := unpacker.upkErr
		unpacker.upkErr = nil
		disposeUnpackedError(upkErr)
	}
	unpackerPool.Put(unpacker)
}

// Error returns the root error in the chain.
func (u *Unpacker) Error() error {
	return u.err
}

// ChainLen returns the length of the error chain (number of wrapped errors).
func (u *Unpacker) ChainLen() int {
	return u.chainLen
}

// Unpack processes and returns the unpacked error representation.
func (u *Unpacker) Unpack() UnpackedError {
	if u.upkErr != nil {
		return *u.upkErr
	}
	if u.err == nil {
		return UnpackedError{}
	}
	err := u.err
	upkErrPtr := newUnpackedError(u.chainLen)
	upkErr := *upkErrPtr
	prvStack := Stack{}

	i := 0
	for err != nil {
		// If the error provides a list of callers, we can use that to build a
		// stack. This includes [*bruh.Err], but also other compatible errors.
		if e, ok := err.(callerser); ok {
			callers := stackPC(e.Callers())
			stack := Stack(upkErr[i].stackStore[:])
			stack = stack[:callers.toStack(stack)]
			var message string
			if m, ok := err.(messager); ok {
				message = m.Message()
			} else {
				message = err.Error()
				nerr := Unwrap(err)
				if nerr != nil {
					nextMessage := nerr.Error()
					if strings.HasSuffix(message, nextMessage) {
						message = message[:len(message)-len(nextMessage)]
						message = strings.TrimSuffix(message, ": ")
					}
				}
			}
			upkErr[i].Err = err
			upkErr[i].Msg = message
			upkErr[i].Stack = stack
			upkErr[i].PartialStack = stack.RelativeTo(prvStack)
			prvStack = stack
			err = Unwrap(err)
			i++
			continue
		}

		// external error without a stack
		extErr := err
		if u.unpackAll {
			err = Unwrap(err)
		} else {
			// keep errors without a stack trace as is
			for {
				err = Unwrap(err)
				if err == nil {
					break
				}
				if _, isCallersErrorer := err.(callerser); isCallersErrorer {
					break
				}
			}
		}
		message := extErr.Error()
		if err != nil {
			nextMessage := err.Error()
			if strings.HasSuffix(message, nextMessage) {
				message = message[:len(message)-len(nextMessage)]
				message = strings.TrimSuffix(message, ": ")
			}
		}
		upkErr[i].Err = extErr
		upkErr[i].Msg = message
		upkErr[i].Stack = Stack{}
		upkErr[i].PartialStack = Stack{}
		i++
	}

	// At this point len(upkErr) == cap(upkErr), thus we have to trim the slice
	// down to its actual size. Furthermore, we have to copy our slice header to
	// he already heap allocated upkErrPtr, which we can later reuse.
	upkErr = upkErr[:i]
	*upkErrPtr = upkErr
	u.upkErr = upkErrPtr

	return upkErr
}

// CombinedStack returns a combined stack trace of all errors in the chain.
func (u *Unpacker) CombinedStack() Stack {
	if u.cbdStk != nil {
		return *u.cbdStk
	}
	if u.err == nil {
		return Stack{}
	}
	stackPtr := newChainStack()
	stack := *stackPtr
	stack = stack[:combinedStack(u.err, stack)]
	*stackPtr = stack
	u.cbdStk = stackPtr
	return stack
}

// GetSourceLines returns the source lines for the given unpacked error. The
// output is in the same order as the unpacked errors and stack frames
// (`[upkErrIdx][partialStackIdx]SourceLines`). If the source code is not
// available, an error is returned. ctxLines is the number of lines before and
// after the requested lines that should be included. colCap is the maximum
// number of characters per line. If unindent is true, the source lines are
// unindented.
func (u *Unpacker) GetSourceLines(ctxLines, colCap int, unindent bool) ([][]SourceLines, error) {
	upkErr := *u.upkErr

	// create a list of files and the source lines we want to read from them
	linesInFiles := make(map[string][]int, len(upkErr))
	for i := range upkErr {
		for _, sf := range upkErr[i].PartialStack {
			if _, ok := linesInFiles[sf.File]; !ok {
				linesInFiles[sf.File] = make([]int, 0, 8) // guess 8 lines per file
			}
			linesInFiles[sf.File] = append(linesInFiles[sf.File], sf.Line)
		}
	}

	type fileLine struct {
		file string
		line int
	}

	// create a mapping from (file, line) combination to source lines
	sourcesInFileLine := make(map[fileLine]SourceLines, len(upkErr))
	for file, lineNums := range linesInFiles {
		sourceLines, err := getSourceLines(osFS{}, file, lineNums, ctxLines, colCap, unindent)
		if err != nil {
			return nil, err
		}
		for i, l := range lineNums {
			sourcesInFileLine[fileLine{file, l}] = sourceLines[i]
		}
	}

	// create the source lines data structure in the same order as the errors and stack frames
	sourceLines := make([][]SourceLines, 0, len(upkErr))
	for i := range upkErr {
		sl := make([]SourceLines, 0, len(upkErr[i].PartialStack))
		for _, sf := range upkErr[i].PartialStack {
			sl = append(sl, sourcesInFileLine[fileLine{sf.File, sf.Line}])
		}
		sourceLines = append(sourceLines, sl)
	}

	return sourceLines, nil
}

var callerserErrorPool = sync.Pool{
	New: func() any {
		return &[]callerser{}
	},
}

// newCallerserErrors allocates a new callerser slice or grabs a cached one.
func newCallerserErrors(size int) *[]callerser {
	callerserErrsPtr := callerserErrorPool.Get().(*[]callerser) //nolint:revive
	if size > cap(*callerserErrsPtr) {
		callerserErrs := make([]callerser, 0, ((size/10)+1)*10)
		*callerserErrsPtr = callerserErrs
	}
	return callerserErrsPtr
}

// disposeCallerserErrors puts the callerser slice back into the pool.
func disposeCallerserErrors(callerserErrs *[]callerser) {
	*callerserErrs = (*callerserErrs)[:0]
	callerserErrorPool.Put(callerserErrs)
}

var combinedStackPCPool = sync.Pool{
	New: func() any {
		stack := make(stackPC, MaxChainStackDepth)
		return &stack
	},
}

// newCombinedStackPC allocates a new stackPC or grabs a cached one.
func newCombinedStackPC() *stackPC {
	return combinedStackPCPool.Get().(*stackPC) //nolint:revive
}

// disposeCombinedStackPC puts the stackPC back into the pool.
func disposeCombinedStackPC(stack *stackPC) {
	*stack = (*stack)[:cap(*stack)]
	combinedStackPCPool.Put(stack)
}

// combinedStack returns a combined stack trace of all errors in the chain.
func combinedStack(err error, stack Stack) int {
	chainLen := 0
	for uerr := err; uerr != nil; uerr = Unwrap(uerr) {
		chainLen++
	}
	errsPtr := newCallerserErrors(chainLen)
	errs := *errsPtr

	// unwrap the errors
	for uerr := err; uerr != nil; uerr = Unwrap(uerr) {
		if cerr, ok := uerr.(callerser); ok {
			callers := cerr.Callers()
			if len(callers) == 0 {
				continue
			}
			frames := allocFrames(callers)
			frame, _ := frames.Next()
			disposeFrames(frames)
			// globally defined errors are no much use to use, so we stop here
			if isGloballyDefinedError(frame.Function) {
				break
			}
			errs = append(errs, cerr)
		}
	}
	if len(errs) == 0 {
		return 0
	}

	// combine the stack traces
	combinedPtr := newCombinedStackPC()
	combined := *combinedPtr
	combined = combined[:copy(combined, errs[len(errs)-1].Callers())]
	for i := len(errs) - 2; i >= 0; i-- {
		current := errs[i].Callers()
		relative := combined.relativeTo(current)
		capacityLeft := cap(combined) - len(relative)
		if capacityLeft-len(current) < 0 {
			combined = append(relative, current[:capacityLeft]...) //nolint:gocritic
			break
		}
		combined = append(relative, current...) //nolint:gocritic
	}
	disposeCallerserErrors(errsPtr)
	n := combined.toStack(stack)
	disposeCombinedStackPC(combinedPtr)

	return n
}

// -----------------------------------------------------------------------------

// UnpackedElement represents a single error frame and the accompanying message.
type UnpackedElement struct {
	// Err is the error instance.
	Err error
	// Msg is the message contained in the error.
	Msg string
	// stackStore is the backing store for Stack and PartialStack.
	stackStore [MaxErrorStackDepth]StackFrame
	// Stack is the error stack for this particular error instance.
	Stack Stack
	// PartialStack is the error stack with parts cut off that are already in
	// the previous error stack.
	PartialStack Stack
}

// UnpackedError represents an unpacked error which is quite useful for
// formatting purposes and other error processing. Use [Unpack] to unpack any
// kind of error that supports it.
type UnpackedError []UnpackedElement

var unpackedErrorPool = sync.Pool{
	New: func() any {
		return &UnpackedError{}
	},
}

// newUnpackedError allocates a new [UnpackedError] or grabs a cached one.
func newUnpackedError(size int) *UnpackedError {
	upkErrPtr := unpackedErrorPool.Get().(*UnpackedError) //nolint:revive
	if size > cap(*upkErrPtr) {
		upkErr := make(UnpackedError, ((size/10)+1)*10)
		*upkErrPtr = upkErr
	}
	return upkErrPtr
}

// disposeUnpackedError puts the [UnpackedError] object back into the pool.
func disposeUnpackedError(upkErr *UnpackedError) {
	*upkErr = (*upkErr)[:cap(*upkErr)]
	unpackedErrorPool.Put(upkErr)
}

// CombinedStack returns a combined stack trace of all errors in the chain.
func (upkErr UnpackedError) CombinedStack() Stack {
	if len(upkErr) == 0 {
		return Stack{}
	}
	numFrames := 0
	for i := 0; i < len(upkErr); i++ {
		numFrames += len(upkErr[i].PartialStack)
	}
	combinedStack := make(Stack, 0, numFrames)
	for i := len(upkErr) - 1; i >= 0; i-- {
		combinedStack = append(combinedStack, upkErr[i].PartialStack...)
	}
	return combinedStack
}
