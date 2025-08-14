package bruh

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/aisbergg/go-bruh/internal/testutils"
)

const (
	file        = "bruh/stack_test.go"
	readFunc    = "bruh.ReadFile"
	parseFunc   = "bruh.ParseFile"
	processFunc = "bruh.ProcessFile"
	prefix      = "bruh."
)

var (
	errEOF = New("unexpected EOF")
	errExt = errors.New("external error")
)

// example func that either returns a wrapped global or creates/wraps a new local error
func ReadFile(fname string, global, external bool) error {
	var err error
	if !external && !global { // local errors
		err = New("unexpected EOF")
	} else if !external && global { // global errors
		err = errEOF
	} else if external && !global { // local external
		err = fmt.Errorf("external context: %w", errors.New("external error"))
	} else { // global external
		err = fmt.Errorf("external context: %w", errExt)
	}
	return Wrapf(err, "error reading file '%v'", fname)
}

// example func that just catches and returns an error
func ParseFile(fname string, global, external bool) error {
	err := ReadFile(fname, global, external)
	if err != nil {
		return err
	}
	return nil
}

// example func that wraps an error with additional context
func ProcessFile(fname string, global, external bool) error {
	// parse the file
	err := ParseFile(fname, global, external)
	if err != nil {
		return Wrapf(err, "error processing file '%v'", fname)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

func TestCombinedStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedCombinedStack0 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 89},
	}
	expectedCombinedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 89},
	}
	expectedCombinedStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 89},
	}

	err := ProcessFile("example.json", false, false)
	uerr := newUnpacker(err, true).Unpack()
	validateStack(t, expectedCombinedStack0, uerr[0].Err.(*Err).Stack())
	validateStack(t, expectedCombinedStack1, uerr[1].Err.(*Err).Stack())
	validateStack(t, expectedCombinedStack2, uerr[2].Err.(*Err).Stack())
}

func TestCombinedStackGlobal(t *testing.T) {
	assert := testutils.NewAssert(t)
	tfname := prefix + t.Name()
	expectedCombinedStack0 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 114},
	}
	expectedCombinedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 114},
	}
	expectedCombinedStack2 := []StackFrame{}

	err := ProcessFile("example.json", true, false)
	uerr := newUnpacker(err, true).Unpack()
	validateStack(t, expectedCombinedStack0, uerr[0].Err.(*Err).Stack())
	validateStack(t, expectedCombinedStack1, uerr[1].Err.(*Err).Stack())
	validateStack(t, expectedCombinedStack2, uerr[2].Err.(*Err).Stack())

	// error with call stack that is larger than MAX_STACK_DEPTH
	bruhErr := (errorFn1()).(*Err)
	stack := bruhErr.Stack()
	assert.Equal(101, len(stack))
	if !strings.HasSuffix(stack[0].Name, "errorFn50") {
		t.Errorf("Expected name to end with '%s', name was '%s'", "errorFn50", stack[0].Name)
	}
	for i := 1; i < 99; i++ {
		expectedName := fmt.Sprintf("errorFn%d", 50-(i+1)/2)
		if !strings.HasSuffix(stack[i].Name, expectedName) {
			t.Errorf("Expected name to end with '%s', name was '%s'", expectedName, stack[i].Name)
		}
	}
}

func TestPartialStack(t *testing.T) {
	assert := testutils.NewAssert(t)
	tfname := prefix + t.Name()
	expectedPartialStack0 := []StackFrame{
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 151},
	}
	expectedPartialStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
	}
	expectedPartialStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
	}

	err := ProcessFile("example.json", false, false)
	uerr := newUnpacker(err, true).Unpack()
	validateStack(t, expectedPartialStack0, uerr[0].PartialStack)
	validateStack(t, expectedPartialStack1, uerr[1].PartialStack)
	validateStack(t, expectedPartialStack2, uerr[2].PartialStack)

	// error with call stack that is larger than MAX_STACK_DEPTH
	err = errorFn1()
	uerr = newUnpacker(err, true).Unpack()
	assert.Equal(len(uerr), 50)
	assert.Equal(len(uerr[0].PartialStack), 3) // same as the original stack
	for i := len(uerr) - 1; i > 0; i-- {
		assert.Equal(
			len(uerr[i].PartialStack),
			2,
			"Expected a partial stack length of %d, got %d; The Stack:\n: %s",
			2,
			len(uerr[i].PartialStack),
			uerr[i].PartialStack,
		)
	}
}

func TestLocalStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedStack0 := []StackFrame{
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 193},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 193},
	}
	expectedStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 193},
	}

	err := ProcessFile("example.json", false, false)
	uerr := newUnpacker(err, true).Unpack()
	validateStack(t, expectedStack0, uerr[0].Stack)
	validateStack(t, expectedStack1, uerr[1].Stack)
	validateStack(t, expectedStack2, uerr[2].Stack)
}

func TestExtLocalStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedStack0 := []StackFrame{
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 214},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 214},
	}
	expectedStack2 := []StackFrame{}

	err := ProcessFile("example.json", false, true)
	uerr := newUnpacker(err, true).Unpack()
	validateStack(t, expectedStack0, uerr[0].Stack)
	validateStack(t, expectedStack1, uerr[1].Stack)
	validateStack(t, expectedStack2, uerr[2].Stack)
}

func TestGlobalStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedStack0 := []StackFrame{
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 235},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 235},
	}
	expectedStack2 := []StackFrame{}

	err := ProcessFile("example.json", true, false)
	uerr := newUnpacker(err, true).Unpack()
	validateStack(t, expectedStack0, uerr[0].Stack)
	validateStack(t, expectedStack1, uerr[1].Stack)
	validateStack(t, expectedStack2, uerr[2].Stack)
}

func TestExtGlobalStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedStack0 := []StackFrame{
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 256},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 256},
	}
	expectedStack2 := []StackFrame{}

	err := ProcessFile("example.json", true, true)
	uerr := newUnpacker(err, true).Unpack()
	validateStack(t, expectedStack0, uerr[0].Stack)
	validateStack(t, expectedStack1, uerr[1].Stack)
	validateStack(t, expectedStack2, uerr[2].Stack)
}

func validateStack(t *testing.T, expected, actual []StackFrame) {
	t.Helper()
	// remove irrelevant frames
	filtered := []StackFrame{}
	for _, f := range actual {
		f.Name = filepath.Base(f.Name)
		if !strings.HasPrefix(f.Name, "bruh.") {
			break
		}
		filtered = append(filtered, f)
	}
	actual = filtered

	if len(expected) != len(filtered) {
		t.Fatalf("expected %v stack frames, got %v", len(expected), len(actual))
	}
	for i := 0; i < len(expected); i++ {
		if expected[i].Name != actual[i].Name {
			t.Errorf("expected func name %v, got %v", expected[i].Name, actual[i].Name)
		}
		if !strings.Contains(actual[i].File, expected[i].File) {
			t.Errorf("expected file name to contain %v, got %v", expected[i].File, actual[i].File)
		}
		if expected[i].Line != actual[i].Line {
			t.Errorf("expected line number %v, got %v", expected[i].Line, actual[i].Line)
		}
	}
}

func TestGoRoutines(t *testing.T) {
	tfname := prefix + t.Name() + ".func1"
	expectedCombinedStack0 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 322},
	}
	expectedCombinedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 322},
	}
	expectedCombinedStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 322},
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			err := ProcessFile("example.json", false, false)
			uerr := newUnpacker(err, true).Unpack()
			validateStack(t, expectedCombinedStack0, uerr[0].Err.(*Err).Stack())
			validateStack(t, expectedCombinedStack1, uerr[1].Err.(*Err).Stack())
			validateStack(t, expectedCombinedStack2, uerr[2].Err.(*Err).Stack())
		}(i)
	}
	wg.Wait()
}

func TestStackString(t *testing.T) {
	stack := Stack{
		StackFrame{Name: "main", File: "main.go", Line: 10, ProgramCounter2: 0x123},
		StackFrame{Name: "foo", File: "foo.go", Line: 20, ProgramCounter2: 0x456},
		StackFrame{Name: "bar", File: "bar.go", Line: 30, ProgramCounter2: 0x789},
	}
	expectedString := "main\n\tmain.go:10 pc=0x123\nfoo\n\tfoo.go:20 pc=0x456\nbar\n\tbar.go:30 pc=0x789"

	result := stack.String()
	if result != expectedString {
		t.Errorf("Expected stack string:\n'%s'\n\nGot:\n'%s'", expectedString, result)
	}
}

func TestStackFirst(t *testing.T) {
	stack := Stack{
		StackFrame{Name: "main", File: "main.go", Line: 10, ProgramCounter2: 0x123},
		StackFrame{Name: "foo", File: "foo.go", Line: 20, ProgramCounter2: 0x456},
		StackFrame{Name: "bar", File: "bar.go", Line: 30, ProgramCounter2: 0x789},
	}
	expectedFirstTwo := Stack{
		StackFrame{Name: "main", File: "main.go", Line: 10, ProgramCounter2: 0x123},
		StackFrame{Name: "foo", File: "foo.go", Line: 20, ProgramCounter2: 0x456},
	}
	expectedFirstOne := Stack{
		StackFrame{Name: "main", File: "main.go", Line: 10, ProgramCounter2: 0x123},
	}

	assert := testutils.NewAssert(t)
	assert.Equal(expectedFirstTwo, stack.First(2))
	assert.Equal(expectedFirstOne, stack.First(1))
}

func TestStackLast(t *testing.T) {
	stack := Stack{
		StackFrame{Name: "main", File: "main.go", Line: 10, ProgramCounter2: 0x123},
		StackFrame{Name: "foo", File: "foo.go", Line: 20, ProgramCounter2: 0x456},
		StackFrame{Name: "bar", File: "bar.go", Line: 30, ProgramCounter2: 0x789},
	}
	expectedLastTwo := Stack{
		StackFrame{Name: "foo", File: "foo.go", Line: 20, ProgramCounter2: 0x456},
		StackFrame{Name: "bar", File: "bar.go", Line: 30, ProgramCounter2: 0x789},
	}
	expectedLastOne := Stack{
		StackFrame{Name: "bar", File: "bar.go", Line: 30, ProgramCounter2: 0x789},
	}

	assert := testutils.NewAssert(t)
	assert.Equal(expectedLastTwo, stack.Last(2))
	assert.Equal(expectedLastOne, stack.Last(1))
}

func TestStackPCRelativeTo(t *testing.T) {
	tests := []struct {
		name          string
		stack         stackPC
		otherStack    stackPC
		expectedStack stackPC
	}{
		{
			name:          "empty stacks",
			stack:         stackPC{},
			otherStack:    stackPC{},
			expectedStack: stackPC{},
		},
		{
			name: "identical stacks",
			stack: stackPC{
				0x123,
				0x456,
				0x789,
			},
			otherStack: stackPC{
				0x123,
				0x456,
				0x789,
			},
			// we keep at least one frame
			expectedStack: stackPC{0x123},
		},
		{
			name: "similar stacks",
			stack: stackPC{
				0x111,
				0x222,
				0x333,
				0x444,
				0x555,
			},
			otherStack: stackPC{
				0x888,
				0x999,
				0x333,
				0x444,
				0x555,
			},
			expectedStack: stackPC{
				0x111,
				0x222,
			},
		},
		{
			name: "different stacks",
			stack: stackPC{
				0x111,
				0x222,
				0x333,
				0x444,
				0x555,
			},
			otherStack: stackPC{
				0x888,
				0x999,
			},
			expectedStack: stackPC{
				0x111,
				0x222,
				0x333,
				0x444,
				0x555,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			relStk := test.stack.relativeTo(test.otherStack)
			assert.Equal(test.expectedStack, relStk)
		})
	}
}

// error with a call stack depth larger than MAX_STACK_DEPTH

//go:noinline
func errorFn1() error {
	return Wrap(errorFn2(), "wrapped")
}

//go:noinline
func errorFn2() error {
	return Wrap(errorFn3(), "wrapped")
}

//go:noinline
func errorFn3() error {
	return Wrap(errorFn4(), "wrapped")
}

//go:noinline
func errorFn4() error {
	return Wrap(errorFn5(), "wrapped")
}

//go:noinline
func errorFn5() error {
	return Wrap(errorFn6(), "wrapped")
}

//go:noinline
func errorFn6() error {
	return Wrap(errorFn7(), "wrapped")
}

//go:noinline
func errorFn7() error {
	return Wrap(errorFn8(), "wrapped")
}

//go:noinline
func errorFn8() error {
	return Wrap(errorFn9(), "wrapped")
}

//go:noinline
func errorFn9() error {
	return Wrap(errorFn10(), "wrapped")
}

//go:noinline
func errorFn10() error {
	return Wrap(errorFn11(), "wrapped")
}

//go:noinline
func errorFn11() error {
	return Wrap(errorFn12(), "wrapped")
}

//go:noinline
func errorFn12() error {
	return Wrap(errorFn13(), "wrapped")
}

//go:noinline
func errorFn13() error {
	return Wrap(errorFn14(), "wrapped")
}

//go:noinline
func errorFn14() error {
	return Wrap(errorFn15(), "wrapped")
}

//go:noinline
func errorFn15() error {
	return Wrap(errorFn16(), "wrapped")
}

//go:noinline
func errorFn16() error {
	return Wrap(errorFn17(), "wrapped")
}

//go:noinline
func errorFn17() error {
	return Wrap(errorFn18(), "wrapped")
}

//go:noinline
func errorFn18() error {
	return Wrap(errorFn19(), "wrapped")
}

//go:noinline
func errorFn19() error {
	return Wrap(errorFn20(), "wrapped")
}

//go:noinline
func errorFn20() error {
	return Wrap(errorFn21(), "wrapped")
}

//go:noinline
func errorFn21() error {
	return Wrap(errorFn22(), "wrapped")
}

//go:noinline
func errorFn22() error {
	return Wrap(errorFn23(), "wrapped")
}

//go:noinline
func errorFn23() error {
	return Wrap(errorFn24(), "wrapped")
}

//go:noinline
func errorFn24() error {
	return Wrap(errorFn25(), "wrapped")
}

//go:noinline
func errorFn25() error {
	return Wrap(errorFn26(), "wrapped")
}

//go:noinline
func errorFn26() error {
	return Wrap(errorFn27(), "wrapped")
}

//go:noinline
func errorFn27() error {
	return Wrap(errorFn28(), "wrapped")
}

//go:noinline
func errorFn28() error {
	return Wrap(errorFn29(), "wrapped")
}

//go:noinline
func errorFn29() error {
	return Wrap(errorFn30(), "wrapped")
}

//go:noinline
func errorFn30() error {
	return Wrap(errorFn31(), "wrapped")
}

//go:noinline
func errorFn31() error {
	return Wrap(errorFn32(), "wrapped")
}

//go:noinline
func errorFn32() error {
	return Wrap(errorFn33(), "wrapped")
}

//go:noinline
func errorFn33() error {
	return Wrap(errorFn34(), "wrapped")
}

//go:noinline
func errorFn34() error {
	return Wrap(errorFn35(), "wrapped")
}

//go:noinline
func errorFn35() error {
	return Wrap(errorFn36(), "wrapped")
}

//go:noinline
func errorFn36() error {
	return Wrap(errorFn37(), "wrapped")
}

//go:noinline
func errorFn37() error {
	return Wrap(errorFn38(), "wrapped")
}

//go:noinline
func errorFn38() error {
	return Wrap(errorFn39(), "wrapped")
}

//go:noinline
func errorFn39() error {
	return Wrap(errorFn40(), "wrapped")
}

//go:noinline
func errorFn40() error {
	return Wrap(errorFn41(), "wrapped")
}

//go:noinline
func errorFn41() error {
	return Wrap(errorFn42(), "wrapped")
}

//go:noinline
func errorFn42() error {
	return Wrap(errorFn43(), "wrapped")
}

//go:noinline
func errorFn43() error {
	return Wrap(errorFn44(), "wrapped")
}

//go:noinline
func errorFn44() error {
	return Wrap(errorFn45(), "wrapped")
}

//go:noinline
func errorFn45() error {
	return Wrap(errorFn46(), "wrapped")
}

//go:noinline
func errorFn46() error {
	return Wrap(errorFn47(), "wrapped")
}

//go:noinline
func errorFn47() error {
	return Wrap(errorFn48(), "wrapped")
}

//go:noinline
func errorFn48() error {
	return Wrap(errorFn49(), "wrapped")
}

//go:noinline
func errorFn49() error {
	return Wrap(errorFn50(), "wrapped")
}

//go:noinline
func errorFn50() error {
	return New("root cause")
}
