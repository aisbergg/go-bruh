package bruh

import (
	"errors"
	"fmt"
	"html/template"
	"runtime"
	"strings"
	"testing"

	"github.com/aisbergg/go-bruh/internal/testutils"
)

func TestUnpack(t *testing.T) {
	tests := map[string]struct {
		cause          error
		input          []string
		unwrapExternal bool
		exp            UnpackedError
	}{
		"nil error": {
			cause: nil,
			input: nil,
			exp:   UnpackedError{},
		},
		"nil root error": {
			cause: nil,
			input: []string{"additional context"},
			exp: UnpackedError{
				{
					Msg: "additional context",
				},
			},
		},
		"standard error wrapping with internal root cause (New)": {
			cause: New("root error"),
			input: []string{"additional context", "even more context"},
			exp: UnpackedError{
				{
					Msg: "even more context",
				},
				{
					Msg: "additional context",
				},
				{
					Msg: "root error",
				},
			},
		},
		"standard error wrapping with external root cause (errors.New)": {
			cause: errors.New("external error"),
			input: []string{"additional context", "even more context"},
			exp: UnpackedError{
				{
					Msg: "even more context",
				},
				{
					Msg: "additional context",
				},
				{
					Msg: "external error",
				},
			},
		},
		"no error wrapping with internal root cause (Errorf)": {
			cause: Errorf("%v", "root error"),
			exp: UnpackedError{
				{
					Msg: "root error",
				},
			},
		},
		"external wrapped error cause, no unwrap (fmt.Errorf)": {
			cause: fmt.Errorf("external wrapped: %w", New("external root error")),
			exp: UnpackedError{
				{
					Msg: "external wrapped: external root error",
				},
			},
		},
		"external wrapped error cause, with unwrap (fmt.Errorf)": {
			cause: fmt.Errorf("external wrapped: %w", New("external root error")),
			exp: UnpackedError{
				{
					Msg: "external wrapped: external root error",
				},
				{
					Msg: "external root error",
				},
			},
			unwrapExternal: true,
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tt.cause, tt.input)
			upk := Unpack(err, tt.unwrapExternal)
			// assert.Equal(tt.exp, upk)
			if !isUnpackedErrorEqual(tt.exp, upk) {
				expMsgs := make([]string, 0, len(tt.exp))
				for _, msg := range tt.exp {
					expMsgs = append(expMsgs, msg.Msg)
				}
				actMsgs := make([]string, 0, len(upk))
				for _, msg := range upk {
					actMsgs = append(actMsgs, msg.Msg)
				}
				t.Errorf("expected %#v, got %#v", expMsgs, actMsgs)
			}
		})
	}
}

func isUnpackedErrorEqual(a, b []UnpackedElement) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Msg != b[i].Msg {
			return false
		}
	}

	return true
}

func TestFormatWithoutTrace(t *testing.T) {
	tests := map[string]struct {
		input error
		exp   string
	}{
		"empty input": {
			input: nil,
			exp:   "",
		},
		"basic root error": {
			input: New("root error"),
			exp:   "root error",
		},
		"basic wrapped error": {
			input: Wrap(Wrap(New("root error"), "additional context"), "even more context"),
			exp:   "even more context: additional context: root error",
		},
		"external wrapped error": {
			input: Wrap(errors.New("external error"), "additional context"),
			exp:   "additional context: external error",
		},
		"external error": {
			input: errors.New("external error"),
			exp:   "external error",
		},
		"empty error": {
			input: New(""),
			exp:   "",
		},
		"empty wrapped external error": {
			input: Wrap(errors.New(""), "additional context"),
			exp:   "additional context",
		},
		"empty wrapped error": {
			input: Wrap(New(""), "additional context"),
			exp:   "additional context",
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			assert.Equal(tt.exp, ToCustomString(tt.input, nil))
		})
	}
}

var funcMaps = template.FuncMap{
	"add": func(x, y int) int {
		return x + y
	},
}

func lineNum() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}

func TestFormatWithTrace(t *testing.T) {
	tests := map[string]struct {
		input error
		line  int
		exp   string
	}{
		"empty input": {
			input: nil,
			exp:   "",
		},
		"basic root error": {
			input: New("root error"),
			line:  lineNum() - 1,
			exp: `root error
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"basic wrapped error": {
			input: Wrap(
				Wrap(
					New("root error"),
					"additional context",
				),
				"even more context",
			),
			line: lineNum() - 7,
			exp: `even more context
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner
additional context
    {{.file}}:{{add .line 1}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
root error
    {{.file}}:{{add .line 2}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace`,
		},
		"external wrapped error": {
			input: Wrap(errors.New("external error"), "additional context"),
			line:  lineNum() - 1,
			exp: `additional context
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner
external error`,
		},
		"external error": {
			input: errors.New("external error"),
			exp:   `external error`,
		},
		"empty error": {
			input: New(""),
			line:  lineNum() - 1,
			exp: `""
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"empty wrapped external error": {
			input: Wrap(errors.New(""), "additional context"),
			line:  lineNum() - 1,
			exp: `additional context
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner
""`,
		},
		"empty wrapped error": {
			input: Wrap(
				New(""),
				"additional context",
			),
			line: lineNum() - 4,
			exp: `additional context
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner
""
    {{.file}}:{{add .line 1}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace`,
		},
	}
	_, testingFile, testingLine, _ := runtime.Caller(1)
	_, file, _, _ := runtime.Caller(0)

	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			expTpl := template.Must(template.New("").Funcs(funcMaps).Parse(tt.exp))
			strBld := strings.Builder{}
			if err := expTpl.Execute(&strBld, map[string]any{
				"file":        file,
				"line":        tt.line,
				"testingFile": testingFile,
				"testingLine": testingLine,
			}); err != nil {
				panic(err)
			}
			exp := strBld.String()
			act := ToCustomString(tt.input, FormatWithTrace)
			if tt.input == nil && len(act) != 0 {
				t.Errorf("expected empty string, got '%s'", act)
			}
			assert.Equal(exp, ToCustomString(tt.input, FormatWithTrace))
		})
	}
}

func TestFormatWithCombinedTrace(t *testing.T) {
	tests := map[string]struct {
		input error
		line  int
		exp   string
	}{
		"empty input": {
			input: nil,
			exp:   "",
		},
		"basic root error": {
			input: New("root error"),
			line:  lineNum() - 1,
			exp: `root error
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"basic wrapped error": {
			input: Wrap(
				Wrap(
					New("root error"),
					"additional context",
				),
				"even more context",
			),
			line: lineNum() - 7,
			exp: `even more context: additional context: root error
    {{.file}}:{{add .line 2}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.file}}:{{add .line 1}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"external wrapped error": {
			input: Wrap(errors.New("external error"), "additional context"),
			line:  lineNum() - 1,
			exp: `additional context: external error
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"external error": {
			input: errors.New("external error"),
			line:  lineNum() - 1,
			exp:   `external error`,
		},
		"empty error": {
			input: New(""),
			line:  lineNum() - 1,
			exp: `""
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"empty wrapped external error": {
			input: Wrap(errors.New(""), "additional context"),
			line:  lineNum() - 1,
			exp: `additional context:` + ` ` + `
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"empty wrapped error": {
			input: Wrap(
				New(""),
				"additional context",
			),
			line: lineNum() - 4,
			exp: `additional context:` + ` ` + `
    {{.file}}:{{add .line 1}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.file}}:{{.line}} in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
	}
	_, testingFile, testingLine, _ := runtime.Caller(1)
	_, file, _, _ := runtime.Caller(0)

	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			expTpl := template.Must(template.New("").Funcs(funcMaps).Parse(tt.exp))
			strBld := strings.Builder{}
			if err := expTpl.Execute(&strBld, map[string]any{
				"file":        file,
				"line":        tt.line,
				"testingFile": testingFile,
				"testingLine": testingLine,
			}); err != nil {
				panic(err)
			}
			exp := strBld.String()
			act := ToCustomString(tt.input, FormatWithTrace)
			if tt.input == nil && len(act) != 0 {
				t.Errorf("expected empty string, got '%s'", act)
			}
			assert.Equal(exp, ToCustomString(tt.input, FormatWithCombinedTrace))
		})
	}
}

func TestFormatPythonTraceback(t *testing.T) {
	tests := map[string]struct {
		input UnpackedError
		exp   string
	}{
		"empty input": {
			input: nil,
			exp:   "",
		},
		"basic root error": {
			input: UnpackedError{
				{
					PartialStack: []StackFrame{
						{File: "<file1>", Line: 10, Name: "<function1>"},
						{File: "<file2>", Line: 20, Name: "<function2>"},
						{File: "<file3>", Line: 30, Name: "<function3>"},
					},
					Err: New("root error"),
					Msg: "root error",
				},
			},
			exp: `Traceback (most recent call last):
  File "<file3>", line 30, in <function3>
  File "<file2>", line 20, in <function2>
  File "<file1>", line 10, in <function1>
*bruh.TraceableError: root error`,
		},
		"basic wrapped error": {
			input: UnpackedError{
				{
					PartialStack: []StackFrame{
						{File: "<file1>", Line: 10, Name: "<function1>"},
						{File: "<file2>", Line: 20, Name: "<function2>"},
						{File: "<file3>", Line: 30, Name: "<function3>"},
					},
					Err: New("foo"),
					Msg: "foo",
				},
				{
					PartialStack: []StackFrame{
						{File: "<file4>", Line: 40, Name: "<function4>"},
						{File: "<file5>", Line: 50, Name: "<function5>"},
						{File: "<file6>", Line: 60, Name: "<function6>"},
					},
					Err: New("bar"),
					Msg: "bar",
				},
				{
					PartialStack: []StackFrame{
						{File: "<file7>", Line: 70, Name: "<function7>"},
						{File: "<file8>", Line: 80, Name: "<function8>"},
						{File: "<file9>", Line: 90, Name: "<function9>"},
					},
					Err: errors.New("bar"),
					Msg: "root error",
				},
			},
			exp: `Traceback (most recent call last):
  File "<file9>", line 90, in <function9>
  File "<file8>", line 80, in <function8>
  File "<file7>", line 70, in <function7>
*errors.errorString: root error

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "<file6>", line 60, in <function6>
  File "<file5>", line 50, in <function5>
  File "<file4>", line 40, in <function4>
*bruh.TraceableError: bar

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "<file3>", line 30, in <function3>
  File "<file2>", line 20, in <function2>
  File "<file1>", line 10, in <function1>
*bruh.TraceableError: foo`,
		},
		"external error": {
			input: UnpackedError{
				{
					PartialStack: []StackFrame{
						{File: "<file1>", Line: 10, Name: "<function1>"},
						{File: "<file2>", Line: 20, Name: "<function2>"},
					},
					Err: errors.New("external error"),
					Msg: "external error",
				},
			},
			exp: `Traceback (most recent call last):
  File "<file2>", line 20, in <function2>
  File "<file1>", line 10, in <function1>
*errors.errorString: external error`,
		},
		"empty error": {
			input: UnpackedError{
				{
					PartialStack: []StackFrame{
						{File: "<file1>", Line: 10, Name: "<function1>"},
					},
					Err: New(""),
					Msg: "",
				},
			},
			exp: `Traceback (most recent call last):
  File "<file1>", line 10, in <function1>
*bruh.TraceableError`,
		},
		"empty wrapped external error": {
			input: UnpackedError{
				{
					PartialStack: []StackFrame{
						{File: "<file1>", Line: 10, Name: "<function1>"},
					},
					Err: New("foo"),
					Msg: "foo",
				},
				{
					PartialStack: []StackFrame{
						{File: "<file2>", Line: 20, Name: "<function2>"},
					},
					Err: errors.New("external error"),
				},
			},
			exp: `Traceback (most recent call last):
  File "<file2>", line 20, in <function2>
*errors.errorString

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "<file1>", line 10, in <function1>
*bruh.TraceableError: foo`,
		},
		"empty wrapped error": {
			input: UnpackedError{
				{
					PartialStack: []StackFrame{
						{File: "<file1>", Line: 10, Name: "<function1>"},
					},
					Err: New("foo"),
					Msg: "foo",
				},
				{
					PartialStack: []StackFrame{
						{File: "<file2>", Line: 20, Name: "<function2>"},
					},
					Err: errors.New("root error"),
					Msg: "root error",
				},
			},
			exp: `Traceback (most recent call last):
  File "<file2>", line 20, in <function2>
*errors.errorString: root error

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "<file1>", line 10, in <function1>
*bruh.TraceableError: foo`,
		},
	}

	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			result := FormatPythonTraceback(tt.input)
			if result != tt.exp {
				t.Errorf("unexpected result:\n%s\nexpected:\n%s", result, tt.exp)
			}
		})
	}
}
