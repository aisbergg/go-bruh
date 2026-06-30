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
	assertFormatMessage := func(name string, input error, exp string) {
		t.Run(name, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			assert.Equal(exp, bruh.StringFormat(input, nil))
			assert.Equal(exp, bruh.Message(input))
			if input != nil {
				assert.Equal(exp, fmt.Sprintf("%v", input))
			}
		})
	}

	assertFormatMessage("EmptyInput", nil, "")
	assertFormatMessage("BasicRootError", bruh.New("root error"), "root error")
	assertFormatMessage(
		"BasicWrappedError",
		bruh.Wrap(
			bruh.Wrap(bruh.New("root error"), "additional context"),
			"even more context",
		),
		"even more context: additional context: root error",
	)
	assertFormatMessage(
		"ExternalWrappedError",
		bruh.Wrap(errors.New("external error"), "additional context"),
		"additional context: external error",
	)
	assertFormatMessage(
		"WrappedPartiallyEmptyError",
		bruh.Wrap(
			bruh.Wrap(
				bruh.Wrap(bruh.Wrap(bruh.Wrap(bruh.New(""), "some context"), ""), ""),
				"even more context",
			),
			"",
		),
		"even more context: some context",
	)
	assertFormatMessage("ExternalError", errors.New("external error"), "external error")
	assertFormatMessage("EmptyError", bruh.New(""), "")
	assertFormatMessage(
		"EmptyWrappedExternalError",
		bruh.Wrap(errors.New(""), "additional context"),
		"additional context",
	)
	assertFormatMessage(
		"EmptyWrappedError",
		bruh.Wrap(bruh.New(""), "additional context"),
		"additional context",
	)
}

func TestAppendMessage(t *testing.T) {
	t.Parallel()
	assertAppendMessage := func(name, prefix string, input error, exp string) {
		t.Run(name, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			result := bruh.AppendMessage([]byte(prefix), input)
			assert.Equal(exp, string(result))
		})
	}

	assertAppendMessage("NilError", "prefix: ", nil, "prefix: ")
	assertAppendMessage("RootError", "prefix: ", singleRootError(), "prefix: root error")
	assertAppendMessage(
		"WrappedError",
		"prefix: ",
		wrappedError2(),
		"prefix: wrapped 2: wrapped 1: root error",
	)
}

func TestMessageLastN(t *testing.T) {
	t.Parallel()
	assertMessageLastN := func(name string, input error, n int, exp string) {
		t.Run(name, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			assert.Equal(exp, bruh.MessageLastN(input, n))
		})
	}

	assertMessageLastN("NilError", nil, 1, "")
	assertMessageLastN("NonPositiveN", wrappedError3(), 0, "")
	assertMessageLastN("LastOne", wrappedError3(), 1, "root error")
	assertMessageLastN("LastTwo", wrappedError3(), 2, "wrapped 1: root error")
	assertMessageLastN(
		"ChainShorterThanN",
		wrappedError3(),
		10,
		"wrapped 3: wrapped 2: wrapped 1: root error",
	)
}

func TestString(t *testing.T) {
	t.Parallel()
	assertString := func(name string, input error, exp string) {
		t.Run(name, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			assert.Equal(exp, bruh.String(input))
		})
	}

	assertString("NilError", nil, "")
	err := wrappedError1()
	assertString("UsesBruhFormatter", err, bruh.StringFormat(err, bruh.BruhFormatter))
}

func TestAppendString(t *testing.T) {
	t.Parallel()
	assertAppendString := func(name, prefix string, input error, exp string) {
		t.Run(name, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			result := bruh.AppendString([]byte(prefix), input)
			assert.Equal(exp, string(result))
		})
	}

	assertAppendString("NilError", "prefix", nil, "prefix")
	err := wrappedError1()
	assertAppendString(
		"UsesBruhFormatter",
		"prefix\n",
		err,
		"prefix\n"+bruh.String(err),
	)
}

func TestStringFormat(t *testing.T) {
	t.Parallel()
	customFormatter := func(b []byte, unpacker *bruh.Unpacker) []byte {
		return fmt.Appendf(b, "chain=%d unpacked=%d", unpacker.ChainLen(), len(unpacker.Unpack()))
	}
	multiExternalWrappedError := fmt.Errorf(
		"external 2: %w",
		fmt.Errorf("external 1: %w", singleRootError()),
	)
	assertStringFormat := func(name string, input error, formatter bruh.Formatter, unpackAll bool, exp string) {
		t.Run(name, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			assert.Equal(exp, bruh.StringFormat(input, formatter, unpackAll))
		})
	}

	assertStringFormat("NilError", nil, customFormatter, false, "")
	assertStringFormat("NilFormatterUsesMessage", wrappedError2(), nil, false, "wrapped 2: wrapped 1: root error")
	assertStringFormat(
		"FormatterReceivesUnpacker",
		multiExternalWrappedError,
		customFormatter,
		false,
		"chain=3 unpacked=2",
	)
	assertStringFormat("UnpackAllPropagates", multiExternalWrappedError, customFormatter, true, "chain=3 unpacked=3")
}

func TestAppendStringFormat(t *testing.T) {
	t.Parallel()
	customFormatter := func(b []byte, unpacker *bruh.Unpacker) []byte {
		return fmt.Appendf(b, "chain=%d unpacked=%d", unpacker.ChainLen(), len(unpacker.Unpack()))
	}
	multiExternalWrappedError := fmt.Errorf(
		"external 2: %w",
		fmt.Errorf("external 1: %w", singleRootError()),
	)
	assertAppendStringFormat := func(
		name, prefix string,
		input error,
		formatter bruh.Formatter,
		unpackAll bool,
		exp string,
	) {
		t.Run(name, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			result := bruh.AppendStringFormat([]byte(prefix), input, formatter, unpackAll)
			assert.Equal(exp, string(result))
		})
	}

	assertAppendStringFormat("NilError", "prefix", nil, customFormatter, false, "prefix")
	assertAppendStringFormat(
		"NilFormatterUsesMessage",
		"prefix: ",
		wrappedError2(),
		nil,
		false,
		"prefix: wrapped 2: wrapped 1: root error",
	)
	assertAppendStringFormat(
		"AppendsCustomFormattedValue",
		"prefix: ",
		multiExternalWrappedError,
		customFormatter,
		true,
		"prefix: chain=3 unpacked=3",
	)
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
