package bruh_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func TestFormatBruhSourced(t *testing.T) {
	t.Parallel()

	singleRootError := singleRootError()
	emptyMessageError := emptyMessageError()
	wrappedError := wrappedError3()
	wrappedEmptyMessageError := wrappedEmptyMessageError()
	externalError := externalError()
	externallyWrappedError := externallyWrappedError()
	wrappedExternalError := wrappedExternalError()
	wrappedExternalInterleavedError := wrappedExternalInterleavedError()
	externallyWrappedNilError := externallyWrappedNilError()
	wrappedGlobalError := wrappedGlobalError()

	assertBruhSourced := func(name string, err error, exp string) {
		t.Run(name, func(t *testing.T) {
			result := bruhTraceSourcedReplacePath(bruh.StringFormat(err, bruh.BruhFancyFormatter(false, true)))
			if result != exp {
				t.Errorf("expected:\n|%s|\n\ngot:\n|%s|", exp, result)
			}
		})
	}

	assertBruhSourced("Nil", nil, "")
	assertBruhSourced("SingleRoot", singleRootError, `root error

at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    21│    //go:noinline
    22│    func singleRootError() error {
  → 23│        return bruh.New("root error")
    24│    }
    25│
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhSourced (/pkg/bruh/format_bruh_trace_sourced_test.go:14)
    12│    t.Parallel()
    13│
  → 14│    singleRootError := singleRootError()
    15│    emptyMessageError := emptyMessageError()
    16│    wrappedError := wrappedError3()
at testing.tRunner (/testing/testing.go:1234)`)
	assertBruhSourced("EmptyMessage", emptyMessageError, `<no message>

at github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError (/pkg/bruh/format_test.go:28)
    26│    //go:noinline
    27│    func emptyMessageError() error {
  → 28│        return bruh.New("")
    29│    }
    30│
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhSourced (/pkg/bruh/format_bruh_trace_sourced_test.go:15)
    13│
    14│    singleRootError := singleRootError()
  → 15│    emptyMessageError := emptyMessageError()
    16│    wrappedError := wrappedError3()
    17│    wrappedEmptyMessageError := wrappedEmptyMessageError()
at testing.tRunner (/testing/testing.go:1234)`)
	assertBruhSourced("Wrapped", wrappedError, `wrapped 3: wrapped 2: wrapped 1: root error

at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    21│    //go:noinline
    22│    func singleRootError() error {
  → 23│        return bruh.New("root error")
    24│    }
    25│
at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError1 (/pkg/bruh/format_test.go:33)
    31│    //go:noinline
    32│    func wrappedError1() error {
  → 33│        if err := singleRootError(); err != nil {
    34│            return bruh.Wrap(err, "wrapped 1")
    35│        }
at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError1 (/pkg/bruh/format_test.go:34)
    32│    func wrappedError1() error {
    33│        if err := singleRootError(); err != nil {
  → 34│            return bruh.Wrap(err, "wrapped 1")
    35│        }
    36│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError2 (/pkg/bruh/format_test.go:41)
    39│    //go:noinline
    40│    func wrappedError2() error {
  → 41│        if err := wrappedError1(); err != nil {
    42│            return bruh.Wrap(err, "wrapped 2")
    43│        }
at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError2 (/pkg/bruh/format_test.go:42)
    40│    func wrappedError2() error {
    41│        if err := wrappedError1(); err != nil {
  → 42│            return bruh.Wrap(err, "wrapped 2")
    43│        }
    44│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3 (/pkg/bruh/format_test.go:49)
    47│    //go:noinline
    48│    func wrappedError3() error {
  → 49│        if err := wrappedError2(); err != nil {
    50│            return bruh.Wrap(err, "wrapped 3")
    51│        }
at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3 (/pkg/bruh/format_test.go:50)
    48│    func wrappedError3() error {
    49│        if err := wrappedError2(); err != nil {
  → 50│            return bruh.Wrap(err, "wrapped 3")
    51│        }
    52│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhSourced (/pkg/bruh/format_bruh_trace_sourced_test.go:16)
    14│    singleRootError := singleRootError()
    15│    emptyMessageError := emptyMessageError()
  → 16│    wrappedError := wrappedError3()
    17│    wrappedEmptyMessageError := wrappedEmptyMessageError()
    18│    externalError := externalError()
at testing.tRunner (/testing/testing.go:1234)`)
	assertBruhSourced("WrappedEmptyMessage", wrappedEmptyMessageError, `<no message>

at github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError (/pkg/bruh/format_test.go:28)
    26│    //go:noinline
    27│    func emptyMessageError() error {
  → 28│        return bruh.New("")
    29│    }
    30│
at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError (/pkg/bruh/format_test.go:57)
    55│    //go:noinline
    56│    func wrappedEmptyMessageError() error {
  → 57│        if err := emptyMessageError(); err != nil {
    58│            return bruh.Wrap(err, "")
    59│        }
at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError (/pkg/bruh/format_test.go:58)
    56│    func wrappedEmptyMessageError() error {
    57│        if err := emptyMessageError(); err != nil {
  → 58│            return bruh.Wrap(err, "")
    59│        }
    60│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhSourced (/pkg/bruh/format_bruh_trace_sourced_test.go:17)
    15│    emptyMessageError := emptyMessageError()
    16│    wrappedError := wrappedError3()
  → 17│    wrappedEmptyMessageError := wrappedEmptyMessageError()
    18│    externalError := externalError()
    19│    externallyWrappedError := externallyWrappedError()
at testing.tRunner (/testing/testing.go:1234)`)
	assertBruhSourced("External", externalError, `external error`)
	assertBruhSourced("WrappedExternal", wrappedExternalError, `wrapped 1: external error

at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalError (/pkg/bruh/format_test.go:76)
    74│    func wrappedExternalError() error {
    75│        if err := externalError(); err != nil {
  → 76│            return bruh.Wrap(err, "wrapped 1")
    77│        }
    78│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhSourced (/pkg/bruh/format_bruh_trace_sourced_test.go:20)
    18│    externalError := externalError()
    19│    externallyWrappedError := externallyWrappedError()
  → 20│    wrappedExternalError := wrappedExternalError()
    21│    wrappedExternalInterleavedError := wrappedExternalInterleavedError()
    22│    externallyWrappedNilError := externallyWrappedNilError()
at testing.tRunner (/testing/testing.go:1234)`)
	assertBruhSourced("ExternallyWrapped", externallyWrappedError, `external error: root error

at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    21│    //go:noinline
    22│    func singleRootError() error {
  → 23│        return bruh.New("root error")
    24│    }
    25│
at github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedError (/pkg/bruh/format_test.go:70)
    68│    //go:noinline
    69│    func externallyWrappedError() error {
  → 70│        return fmt.Errorf("external error: %w", singleRootError())
    71│    }
    72│
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhSourced (/pkg/bruh/format_bruh_trace_sourced_test.go:19)
    17│    wrappedEmptyMessageError := wrappedEmptyMessageError()
    18│    externalError := externalError()
  → 19│    externallyWrappedError := externallyWrappedError()
    20│    wrappedExternalError := wrappedExternalError()
    21│    wrappedExternalInterleavedError := wrappedExternalInterleavedError()
at testing.tRunner (/testing/testing.go:1234)`)
	assertBruhSourced(
		"WrappedExternalInterleaved",
		wrappedExternalInterleavedError,
		`wrapped: external error: root error

at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    21│    //go:noinline
    22│    func singleRootError() error {
  → 23│        return bruh.New("root error")
    24│    }
    25│
at github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedError (/pkg/bruh/format_test.go:70)
    68│    //go:noinline
    69│    func externallyWrappedError() error {
  → 70│        return fmt.Errorf("external error: %w", singleRootError())
    71│    }
    72│
at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError (/pkg/bruh/format_test.go:83)
    81│    //go:noinline
    82│    func wrappedExternalInterleavedError() error {
  → 83│        if err := externallyWrappedError(); err != nil {
    84│            return bruh.Wrap(err, "wrapped")
    85│        }
at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError (/pkg/bruh/format_test.go:84)
    82│    func wrappedExternalInterleavedError() error {
    83│        if err := externallyWrappedError(); err != nil {
  → 84│            return bruh.Wrap(err, "wrapped")
    85│        }
    86│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhSourced (/pkg/bruh/format_bruh_trace_sourced_test.go:21)
    19│    externallyWrappedError := externallyWrappedError()
    20│    wrappedExternalError := wrappedExternalError()
  → 21│    wrappedExternalInterleavedError := wrappedExternalInterleavedError()
    22│    externallyWrappedNilError := externallyWrappedNilError()
    23│    wrappedGlobalError := wrappedGlobalError()
at testing.tRunner (/testing/testing.go:1234)`,
	)
	assertBruhSourced("ExternallyWrappedNil", externallyWrappedNilError, `wrapped: external error: %!w(<nil>)

at github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedNilError (/pkg/bruh/format_test.go:92)
    90│    func externallyWrappedNilError() error {
    91│        if err := fmt.Errorf("external error: %w", nil); err != nil {
  → 92│            return bruh.Wrap(err, "wrapped")
    93│        }
    94│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhSourced (/pkg/bruh/format_bruh_trace_sourced_test.go:22)
    20│    wrappedExternalError := wrappedExternalError()
    21│    wrappedExternalInterleavedError := wrappedExternalInterleavedError()
  → 22│    externallyWrappedNilError := externallyWrappedNilError()
    23│    wrappedGlobalError := wrappedGlobalError()
    24│
at testing.tRunner (/testing/testing.go:1234)`)
	assertBruhSourced("WrappedGlobal", wrappedGlobalError, `wrapped: globally wrapped: root error

at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedGlobalError (/pkg/bruh/format_test.go:99)
     97│    //go:noinline
     98│    func wrappedGlobalError() error {
  →  99│        return bruh.Wrap(globalWrappedError, "wrapped")
    100│    }
    101│
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhSourced (/pkg/bruh/format_bruh_trace_sourced_test.go:23)
    21│    wrappedExternalInterleavedError := wrappedExternalInterleavedError()
    22│    externallyWrappedNilError := externallyWrappedNilError()
  → 23│    wrappedGlobalError := wrappedGlobalError()
    24│
    25│    assertBruhSourced := func(name string, err error, exp string) {
at testing.tRunner (/testing/testing.go:1234)`)
}

var (
	bruhTraceSourcedRegexpTestingGo        = regexp.MustCompile(`(?m)testing\.go:\d+\)\n  (.+(\n  )?)+`)
	bruhTraceTrailingSpacesRegexpTestingGo = regexp.MustCompile(`(?m) *$`)
)

func bruhTraceSourcedReplacePath(s string) string {
	s = strings.ReplaceAll(s, goTestingPath, "")
	s = strings.ReplaceAll(s, repoDir, "")
	s = bruhTraceSourcedRegexpTestingGo.ReplaceAllLiteralString(s, `testing.go:1234)`)
	s = bruhTraceTrailingSpacesRegexpTestingGo.ReplaceAllLiteralString(s, "")
	s = regexpMemoryAddress.ReplaceAllLiteralString(s, "0x012345")
	return s
}
