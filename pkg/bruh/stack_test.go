package bruh

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
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
func ReadFile(fname string, global bool, external bool) error {
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
func ParseFile(fname string, global bool, external bool) error {
	err := ReadFile(fname, global, external)
	if err != nil {
		return err
	}
	return nil
}

// example func that wraps an error with additional context
func ProcessFile(fname string, global bool, external bool) error {
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
		{Name: readFunc, File: file, Line: 29},
		{Name: readFunc, File: file, Line: 37},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 87},
	}
	expectedFullStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 29},
		{Name: readFunc, File: file, Line: 37},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: tfname, File: file, Line: 87},
	}
	expectedFullStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 29},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: tfname, File: file, Line: 87},
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
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 109},
	}
	expectedPartialStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 37},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
	}
	expectedPartialStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 29},
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
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 135},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 37},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: tfname, File: file, Line: 135},
	}
	expectedStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 29},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: tfname, File: file, Line: 135},
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
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 156},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 37},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: tfname, File: file, Line: 156},
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
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 177},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 37},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: tfname, File: file, Line: 177},
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
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 198},
	}
	expectedStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 37},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: tfname, File: file, Line: 198},
	}
	expectedStack2 := []StackFrame{}

	err := ProcessFile("example.json", true, true)
	uerr := Unpack(err, true)
	validateStack(t, expectedStack0, uerr[0].Stack)
	validateStack(t, expectedStack1, uerr[1].Stack)
	validateStack(t, expectedStack2, uerr[2].Stack)
}

func validateStack(t *testing.T, expected []StackFrame, actual []StackFrame) {
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
		{Name: readFunc, File: file, Line: 29},
		{Name: readFunc, File: file, Line: 37},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: processFunc, File: file, Line: 54},
		{Name: tfname, File: file, Line: 263},
	}
	expectedFullStack1 := []StackFrame{
		{Name: readFunc, File: file, Line: 29},
		{Name: readFunc, File: file, Line: 37},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: tfname, File: file, Line: 263},
	}
	expectedFullStack2 := []StackFrame{
		{Name: readFunc, File: file, Line: 29},
		{Name: parseFunc, File: file, Line: 42},
		{Name: processFunc, File: file, Line: 52},
		{Name: tfname, File: file, Line: 263},
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

func dummyStack() error {
	return New("unexpected EOF")
}
