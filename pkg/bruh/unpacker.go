package bruh

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/aisbergg/go-bruh/internal/stringbuilder"
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
		dispose4xStack(stack)
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

// CombinedMessage returns a combined message of all errors in the chain.
func (u *Unpacker) CombinedMessage() string {
	if _, ok := u.err.(messager); ok {
		buf := make([]byte, 0, u.ChainLen()*100)
		buf = u.AppendMessage(buf)
		return unsafe.String(unsafe.SliceData(buf), len(buf)) //nolint:gosec

	}
	return u.err.Error()
}

// AppendMessage appends the combined message of all errors in the chain to the
// provided byte slice.
func (u *Unpacker) AppendMessage(buf []byte) []byte {
	return appendMessageToBuffer(buf, u.err)
}

// AppendMessageBuilder appends the combined message of all errors in the chain
// to the provided [stringbuilder.StringBuilder].
func (u *Unpacker) AppendMessageBuilder(stringBuilder *stringbuilder.StringBuilder) {
	var messageWritten bool
	for err := u.err; err != nil; err = Unwrap(err) {
		var msg string
		var cont bool
		if e, ok := err.(messager); ok {
			msg = e.Message()
			cont = true
		} else {
			msg = err.Error()
		}

		if msg == "" {
			continue
		}

		if messageWritten {
			stringBuilder.WriteString(": ")
		}
		stringBuilder.WriteString(msg)
		messageWritten = true

		if !cont {
			break
		}
	}
}

// CombinedStack returns a combined stack trace of all errors in the chain.
func (u *Unpacker) CombinedStack() Stack {
	if u.cbdStk != nil {
		return *u.cbdStk
	}
	if u.err == nil {
		return Stack{}
	}
	stackPtr := new4xStack()
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
		sourceLines, err := getSourceLinesFromFile(file, lineNums, ctxLines, colCap, unindent)
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

// newUnpackedError allocates a new [UnpackedError] or grabs a cached one.
func newCallerserErrors(size int) *[]callerser {
	callerserErrsPtr := callerserErrorPool.Get().(*[]callerser) //nolint:revive
	if size > cap(*callerserErrsPtr) {
		callerserErrs := make([]callerser, 0, ((size/10)+1)*10)
		*callerserErrsPtr = callerserErrs
	}
	return callerserErrsPtr
}

// dispose[]callerser puts the [[]callerser] object back into the pool.
func disposeCallerserErrors(callerserErrs *[]callerser) {
	*callerserErrs = (*callerserErrs)[:0]
	callerserErrorPool.Put(callerserErrs)
}

var stackPCPool = sync.Pool{
	New: func() any {
		stack := make(stackPC, 4*MAX_STACK_DEPTH)
		return &stack
	},
}

// combinedStack returns a combined stack trace of all errors in the chain.
func combinedStack(err error, stack Stack) int {
	chainLen := 0
	for uerr := err; uerr != nil; uerr = Unwrap(uerr) {
		chainLen++
	}
	// errs := make([]callerser, 0, chainLen)
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
	combinedPtr := stackPCPool.Get().(*stackPC) //nolint:revive
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
	stackPCPool.Put(combinedPtr)

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
	stackStore [MAX_STACK_DEPTH]StackFrame
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

// -----------------------------------------------------------------------------

// SourceLines represents a collection of SourceLine elements used for
// generating snippets of source code in error messages.
type SourceLines []SourceLine

// SourceLine represents a single line of source code along with its line number.
// LineNum specifies the line number in the source file.
// Source contains the actual content of the line.
type SourceLine struct {
	LineNum int
	Source  string
}

type lineToIndex struct {
	lineNum int
	index1  int
	index2  int
	isLine  bool
}

// getSourceLinesFromFile reads the given lines (index starting at 1) of source
// code from a file. ctxLines are the number of lines before and after the
// requested lines that should be included. colCap is the maximum number of
// characters per line. If unindent is true, the source lines are unindented.
func getSourceLinesFromFile(
	file string,
	lines []int,
	ctxLines, colCap int,
	unindent bool,
) ([]SourceLines, error) {
	ctxLines = max(0, ctxLines)
	colCap = max(0, colCap)

	// initialize the source lines data structure
	numLines := len(lines) * (2*ctxLines + 1)
	sourceLines := make([]SourceLines, len(lines))
	for i := range lines {
		sourceLines[i] = make([]SourceLine, 2*ctxLines+1)
	}

	// create a mapping from line numbers to indices in the sourceLines data
	// structure for efficient access to the file
	linesToIndex := make([]lineToIndex, 0, numLines)
	for i, l := range lines {
		index2 := 0
		for j := ctxLines - 1; j >= 0; j-- {
			linesToIndex = append(
				linesToIndex,
				lineToIndex{lineNum: l - j - 1, index1: i, index2: index2, isLine: false},
			)
			index2++
		}
		linesToIndex = append(
			linesToIndex,
			lineToIndex{lineNum: l, index1: i, index2: index2, isLine: true},
		)
		index2++
		for j := range ctxLines {
			linesToIndex = append(
				linesToIndex,
				lineToIndex{lineNum: l + j + 1, index1: i, index2: index2, isLine: false},
			)
			index2++
		}
	}
	sort.Slice(linesToIndex, func(i, j int) bool {
		return linesToIndex[i].lineNum < linesToIndex[j].lineNum
	})

	// ensure we only read files in the current working directory
	if filepath.IsAbs(file) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, Wrap(err, "getting current working directory")
		}
		file, err = filepath.Rel(wd, file)
		if err != nil {
			return nil, Wrap(err, "getting relative path to source file")
		}
	}
	// ensure '.go' file extension, so we reduce the risk of reading anything
	// that is not supposed to be read
	if !strings.HasSuffix(file, ".go") {
		return nil, New("source file must have a .go extension")
	}
	f, err := os.Open(file) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	// try to find the source lines in the file and store them in the sourceLines data structure
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	moreLinesInFile := true
	currentLine := 1
outer:
	for _, lti := range linesToIndex {
		if lti.lineNum < 1 {
			continue
		}
		if !moreLinesInFile {
			sourceLines[lti.index1][lti.index2] = SourceLine{LineNum: lti.lineNum}
			continue
		}
		for currentLine < lti.lineNum {
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return nil, err
				}
				if lti.isLine {
					return nil, New("source file too short")
				}
				moreLinesInFile = false
				sourceLines[lti.index1][lti.index2] = SourceLine{LineNum: lti.lineNum}
				continue outer
			}
			currentLine++
		}
		sourceLines[lti.index1][lti.index2] = SourceLine{
			LineNum: lti.lineNum,
			Source:  scanner.Text(),
		}
	}

	// unindent the source lines
	if unindent {
		for i := range sourceLines {
			// count the number of leading tabs
			minTabIndents := int(^uint(0) >> 1)
			for _, sl := range sourceLines[i] {
				// skip lines that are not in the file or are empty
				if sl.LineNum <= 0 || sl.Source == "" {
					continue
				}
				lineTabIndents := 0
				for _, c := range sl.Source {
					if c != '\t' {
						break
					}
					lineTabIndents++
				}
				if lineTabIndents < minTabIndents {
					minTabIndents = lineTabIndents
				}
			}
			// strip leading tabs
			for j, sl := range sourceLines[i] {
				// skip lines that are not in the file or are empty
				if sl.LineNum <= 0 || sl.Source == "" {
					continue
				}
				sourceLines[i][j].Source = sourceLines[i][j].Source[minTabIndents:]
			}
		}
	}

	// trim the source lines to the given column capacity
	if colCap > 0 {
		for i := range sourceLines {
			for j := range sourceLines[i] {
				if len(sourceLines[i][j].Source) > colCap {
					sourceLines[i][j].Source = sourceLines[i][j].Source[:colCap]
				}
			}
		}
	}

	return sourceLines, nil
}
