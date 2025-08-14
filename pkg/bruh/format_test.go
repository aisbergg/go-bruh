package bruh_test

import (
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/aisbergg/go-bruh/internal/testutils"
	"github.com/aisbergg/go-bruh/pkg/bruh"
)

var (
	globalSingleError  = bruh.New("root error")
	globalWrappedError = bruh.Wrap(globalSingleError, "globally wrapped")
)

//go:noinline
func singleRootError() error {
	return bruh.New("root error")
}

//go:noinline
func emptyMessageError() error {
	return bruh.New("")
}

//go:noinline
func wrappedError1() error {
	if err := singleRootError(); err != nil {
		return bruh.Wrap(err, "wrapped 1")
	}
	return nil
}

//go:noinline
func wrappedError2() error {
	if err := wrappedError1(); err != nil {
		return bruh.Wrap(err, "wrapped 2")
	}
	return nil
}

//go:noinline
func wrappedError3() error {
	if err := wrappedError2(); err != nil {
		return bruh.Wrap(err, "wrapped 3")
	}
	return nil
}

//go:noinline
func wrappedEmptyMessageError() error {
	if err := emptyMessageError(); err != nil {
		return bruh.Wrap(err, "")
	}
	return nil
}

//go:noinline
func externalError() error {
	return errors.New("external error")
}

//go:noinline
func externallyWrappedError() error {
	return fmt.Errorf("external error: %w", singleRootError())
}

//go:noinline
func wrappedExternalError() error {
	if err := externalError(); err != nil {
		return bruh.Wrap(err, "wrapped 1")
	}
	return nil
}

//go:noinline
func wrappedExternalInterleavedError() error {
	if err := externallyWrappedError(); err != nil {
		return bruh.Wrap(err, "wrapped")
	}
	return nil
}

//go:noinline
func externallyWrappedNilError() error {
	if err := fmt.Errorf("external error: %w", nil); err != nil {
		return bruh.Wrap(err, "wrapped")
	}
	return nil
}

//go:noinline
func wrappedGlobalError() error {
	return bruh.Wrap(globalWrappedError, "wrapped")
}

var (
	repoDir             string
	goTestingPath       string
	regexpMemoryAddress = regexp.MustCompile(`0x[0-9a-fA-F]+`)
)

func init() {
	_, repoDir, _, _ = runtime.Caller(0)
	repoDir = filepath.Dir(filepath.Dir(filepath.Dir(repoDir)))
	_, goTestingPath, _, _ = runtime.Caller(2)
	goTestingPath = filepath.Dir(filepath.Dir(goTestingPath))
}

func TestFormatMessageOnly(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input error
		exp   string
	}{
		"empty input": {
			input: nil,
			exp:   "",
		},
		"basic root error": {
			input: bruh.New("root error"),
			exp:   "root error",
		},
		"basic wrapped error": {
			input: bruh.Wrap(
				bruh.Wrap(bruh.New("root error"), "additional context"),
				"even more context",
			),
			exp: "even more context: additional context: root error",
		},
		"external wrapped error": {
			input: bruh.Wrap(errors.New("external error"), "additional context"),
			exp:   "additional context: external error",
		},
		"wrapped partially empty error": {
			input: bruh.Wrap(
				bruh.Wrap(
					bruh.Wrap(bruh.Wrap(bruh.Wrap(bruh.New(""), "some context"), ""), ""),
					"even more context",
				),
				"",
			),
			exp: "even more context: some context",
		},
		"external error": {
			input: errors.New("external error"),
			exp:   "external error",
		},
		"empty error": {
			input: bruh.New(""),
			exp:   "",
		},
		"empty wrapped external error": {
			input: bruh.Wrap(errors.New(""), "additional context"),
			exp:   "additional context",
		},
		"empty wrapped error": {
			input: bruh.Wrap(bruh.New(""), "additional context"),
			exp:   "additional context",
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			assert.Equal(tt.exp, bruh.StringFormat(tt.input, nil))
			assert.Equal(tt.exp, bruh.Message(tt.input))
			if tt.input != nil {
				assert.Equal(tt.exp, fmt.Sprintf("%v", tt.input))
			}
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

func BenchmarkFormatters(b *testing.B) {
	for _, tc := range []struct {
		name string
		fmt  bruh.Formatter
	}{
		{"WithoutTrace", nil},
		{"Bruh", bruh.BruhFormatter},
		{"BruhColored", bruh.BruhFancyFormatter(true, false)},
		{"BruhSourced", bruh.BruhFancyFormatter(false, true)},
		{"BruhSourcedColored", bruh.BruhFancyFormatter(true, true)},
		{"BruhStacked", bruh.BruhStackedFormatter},
		{"BruhStackedColored", bruh.BruhStackedFancyFormatter(true, false, false)},
		{"BruhStackedTyped", bruh.BruhStackedFancyFormatter(false, false, true)},
		{"BruhStackedTypedColored", bruh.BruhStackedFancyFormatter(true, false, true)},
		{"BruhStackedSourced", bruh.BruhStackedFancyFormatter(false, true, false)},
		{"BruhStackedSourcedColored", bruh.BruhStackedFancyFormatter(true, true, false)},
		{"BruhStackedSourcedTypedColored", bruh.BruhStackedFancyFormatter(true, true, true)},
		{"GoPanic", bruh.GoPanicFormatter},
		{"JavaStackTrace", bruh.JavaStackTraceFormatter},
		{"PythonTraceback", bruh.PythonTracebackFormatter},
	} {
		b.Run(fmt.Sprintf("%v", tc.name), func(b *testing.B) {
			err := wrappedError(20)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = bruh.StringFormat(err, tc.fmt)
			}
			_ = str
		})
	}
}

func wrappedError(layers int) error {
	err := bruh.New("error")
	for i := 0; i < layers; i++ {
		err = bruh.Wrap(err, "wrap")
	}
	return err
}
