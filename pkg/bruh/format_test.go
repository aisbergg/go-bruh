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

func TestFormatWithTrace(t *testing.T) {
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
			exp: `root error
    {{.file}}:189 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"basic wrapped error": {
			input: Wrap(Wrap(New("root error"), "additional context"), "even more context"),
			exp: `even more context
    {{.file}}:195 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner
additional context
    {{.file}}:195 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
root error
    {{.file}}:195 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace`,
		},
		"external wrapped error": {
			input: Wrap(errors.New("external error"), "additional context"),
			exp: `additional context
    {{.file}}:205 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner
external error`,
		},
		"external error": {
			input: errors.New("external error"),
			exp:   `external error`,
		},
		"empty error": {
			input: New(""),
			exp: `""
    {{.file}}:216 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"empty wrapped external error": {
			input: Wrap(errors.New(""), "additional context"),
			exp: `additional context
    {{.file}}:222 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner
""`,
		},
		"empty wrapped error": {
			input: Wrap(New(""), "additional context"),
			exp: `additional context
    {{.file}}:229 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner
""
    {{.file}}:229 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithTrace`,
		},
	}
	_, testingFile, testingLine, _ := runtime.Caller(1)
	_, file, _, _ := runtime.Caller(0)

	print(file)
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			expTpl := template.Must(template.New("").Parse(tt.exp))
			strBld := strings.Builder{}
			if err := expTpl.Execute(&strBld, map[string]any{
				"file":        file,
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
		exp   string
	}{
		"empty input": {
			input: nil,
			exp:   "",
		},
		"basic root error": {
			input: New("root error"),
			exp: `root error
    {{.file}}:273 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"basic wrapped error": {
			input: Wrap(Wrap(New("root error"), "additional context"), "even more context"),
			exp: `even more context: additional context: root error
    {{.file}}:279 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.file}}:279 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.file}}:279 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"external wrapped error": {
			input: Wrap(errors.New("external error"), "additional context"),
			exp: `additional context: external error
    {{.file}}:287 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"external error": {
			input: errors.New("external error"),
			exp:   `external error`,
		},
		"empty error": {
			input: New(""),
			exp: `""
    {{.file}}:297 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"empty wrapped external error": {
			input: Wrap(errors.New(""), "additional context"),
			exp: `additional context:` + ` ` + `
    {{.file}}:303 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
		"empty wrapped error": {
			input: Wrap(New(""), "additional context"),
			exp: `additional context:` + ` ` + `
    {{.file}}:309 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.file}}:309 in github.com/aisbergg/go-bruh/pkg/bruh.TestFormatWithCombinedTrace
    {{.testingFile}}:{{.testingLine}} in testing.tRunner`,
		},
	}
	_, testingFile, testingLine, _ := runtime.Caller(1)
	_, file, _, _ := runtime.Caller(0)

	print(file)
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			expTpl := template.Must(template.New("").Parse(tt.exp))
			strBld := strings.Builder{}
			if err := expTpl.Execute(&strBld, map[string]any{
				"file":        file,
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
