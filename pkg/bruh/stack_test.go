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

func TestFullStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedFullStack0 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 89},
	}
	expectedFullStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 89},
	}
	expectedFullStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 89},
	}

	err := ProcessFile("example.json", false, false)
	uerr := Unpack(err, true)
	validateStack(t, expectedFullStack0, uerr[0].Err.(*TraceableError).FullStack())
	validateStack(t, expectedFullStack1, uerr[1].Err.(*TraceableError).FullStack())
	validateStack(t, expectedFullStack2, uerr[2].Err.(*TraceableError).FullStack())
}

func TestPartialStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedPartialStack0 := []StackFrame{
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 111},
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
	uerr := Unpack(err, true)
	validateStack(t, expectedPartialStack0, uerr[0].PartialStack)
	validateStack(t, expectedPartialStack1, uerr[1].PartialStack)
	validateStack(t, expectedPartialStack2, uerr[2].PartialStack)
}

func TestLocalStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedStack0 := []StackFrame{
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 137},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 137},
	}
	expectedStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 137},
	}

	err := ProcessFile("example.json", false, false)
	uerr := Unpack(err, true)
	validateStack(t, expectedStack0, uerr[0].Stack)
	validateStack(t, expectedStack1, uerr[1].Stack)
	validateStack(t, expectedStack2, uerr[2].Stack)
}

func TestExtLocalStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedStack0 := []StackFrame{
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 158},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 158},
	}
	expectedStack2 := []StackFrame{}

	err := ProcessFile("example.json", false, true)
	uerr := Unpack(err, true)
	validateStack(t, expectedStack0, uerr[0].Stack)
	validateStack(t, expectedStack1, uerr[1].Stack)
	validateStack(t, expectedStack2, uerr[2].Stack)
}

func TestGlobalStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedStack0 := []StackFrame{
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 179},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 179},
	}
	expectedStack2 := []StackFrame{}

	err := ProcessFile("example.json", true, false)
	uerr := Unpack(err, true)
	validateStack(t, expectedStack0, uerr[0].Stack)
	validateStack(t, expectedStack1, uerr[1].Stack)
	validateStack(t, expectedStack2, uerr[2].Stack)
}

func TestExtGlobalStack(t *testing.T) {
	tfname := prefix + t.Name()
	expectedStack0 := []StackFrame{
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 200},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 200},
	}
	expectedStack2 := []StackFrame{}

	err := ProcessFile("example.json", true, true)
	uerr := Unpack(err, true)
	validateStack(t, expectedStack0, uerr[0].Stack)
	validateStack(t, expectedStack1, uerr[1].Stack)
	validateStack(t, expectedStack2, uerr[2].Stack)
}

func validateStack(t *testing.T, expected, actual []StackFrame) {
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
	expectedFullStack0 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: processFunc, File: file, Line: 56},
		{Name: tfname, File: file, Line: 265},
	}
	expectedFullStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: readFunc, File: file, Line: 39},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 265},
	}
	expectedFullStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 31},
		{Name: parseFunc, File: file, Line: 44},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 265},
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			err := ProcessFile("example.json", false, false)
			uerr := Unpack(err, true)
			validateStack(t, expectedFullStack0, uerr[0].Err.(*TraceableError).FullStack())
			validateStack(t, expectedFullStack1, uerr[1].Err.(*TraceableError).FullStack())
			validateStack(t, expectedFullStack2, uerr[2].Err.(*TraceableError).FullStack())
		}(i)
	}
	wg.Wait()
}

func TestStackString(t *testing.T) {
	stack := Stack{
		StackFrame{Name: "main", File: "main.go", Line: 10, ProgramCounter: 0x123},
		StackFrame{Name: "foo", File: "foo.go", Line: 20, ProgramCounter: 0x456},
		StackFrame{Name: "bar", File: "bar.go", Line: 30, ProgramCounter: 0x789},
	}
	expectedString := "main\n        main.go:10 pc=0x123\nfoo\n        foo.go:20 pc=0x456\nbar\n        bar.go:30 pc=0x789\n"

	result := stack.String()
	if result != expectedString {
		t.Errorf("Expected stack string:\n%s\n\nGot:\n%s", expectedString, result)
	}
}

func TestStackFirst(t *testing.T) {
	stack := Stack{
		StackFrame{Name: "main", File: "main.go", Line: 10, ProgramCounter: 0x123},
		StackFrame{Name: "foo", File: "foo.go", Line: 20, ProgramCounter: 0x456},
		StackFrame{Name: "bar", File: "bar.go", Line: 30, ProgramCounter: 0x789},
	}
	expectedFirstTwo := Stack{
		StackFrame{Name: "main", File: "main.go", Line: 10, ProgramCounter: 0x123},
		StackFrame{Name: "foo", File: "foo.go", Line: 20, ProgramCounter: 0x456},
	}
	expectedFirstOne := Stack{
		StackFrame{Name: "main", File: "main.go", Line: 10, ProgramCounter: 0x123},
	}

	assert := testutils.NewAssert(t)
	assert.Equal(expectedFirstTwo, stack.First(2))
	assert.Equal(expectedFirstOne, stack.First(1))
}

func TestStackLast(t *testing.T) {
	stack := Stack{
		StackFrame{Name: "main", File: "main.go", Line: 10, ProgramCounter: 0x123},
		StackFrame{Name: "foo", File: "foo.go", Line: 20, ProgramCounter: 0x456},
		StackFrame{Name: "bar", File: "bar.go", Line: 30, ProgramCounter: 0x789},
	}
	expectedLastTwo := Stack{
		StackFrame{Name: "foo", File: "foo.go", Line: 20, ProgramCounter: 0x456},
		StackFrame{Name: "bar", File: "bar.go", Line: 30, ProgramCounter: 0x789},
	}
	expectedLastOne := Stack{
		StackFrame{Name: "bar", File: "bar.go", Line: 30, ProgramCounter: 0x789},
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
