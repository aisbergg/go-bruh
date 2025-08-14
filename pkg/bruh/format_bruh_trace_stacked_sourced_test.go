package bruh_test

import (
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func TestFormatBruhStackedSourced(t *testing.T) {
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

	tests := []struct {
		name string
		err  error
		exp  string
	}{
		{
			name: "Nil",
			err:  nil,
			exp:  ``,
		},
		{
			name: "SingleRoot",
			err:  singleRootError,
			exp: `root error

at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    21│    //go:noinline
    22│    func singleRootError() error {
  → 23│        return bruh.New("root error")
    24│    }
    25│
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStackedSourced (/pkg/bruh/format_bruh_trace_stacked_sourced_test.go:12)
    10│    t.Parallel()
    11│
  → 12│    singleRootError := singleRootError()
    13│    emptyMessageError := emptyMessageError()
    14│    wrappedError := wrappedError3()
at testing.tRunner (/testing/testing.go:1234)`,
		},
		{
			name: "EmptyMessage",
			err:  emptyMessageError,
			exp: `<no message>

at github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError (/pkg/bruh/format_test.go:28)
    26│    //go:noinline
    27│    func emptyMessageError() error {
  → 28│        return bruh.New("")
    29│    }
    30│
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStackedSourced (/pkg/bruh/format_bruh_trace_stacked_sourced_test.go:13)
    11│
    12│    singleRootError := singleRootError()
  → 13│    emptyMessageError := emptyMessageError()
    14│    wrappedError := wrappedError3()
    15│    wrappedEmptyMessageError := wrappedEmptyMessageError()
at testing.tRunner (/testing/testing.go:1234)`,
		},
		{
			name: "Wrapped",
			err:  wrappedError,
			exp: `wrapped 3

at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3 (/pkg/bruh/format_test.go:50)
    48│    func wrappedError3() error {
    49│        if err := wrappedError2(); err != nil {
  → 50│            return bruh.Wrap(err, "wrapped 3")
    51│        }
    52│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStackedSourced (/pkg/bruh/format_bruh_trace_stacked_sourced_test.go:14)
    12│    singleRootError := singleRootError()
    13│    emptyMessageError := emptyMessageError()
  → 14│    wrappedError := wrappedError3()
    15│    wrappedEmptyMessageError := wrappedEmptyMessageError()
    16│    externalError := externalError()
at testing.tRunner (/testing/testing.go:1234)
wrapped 2

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
wrapped 1

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
root error

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
    35│        }`,
		},
		{
			name: "WrappedEmptyMessage",
			err:  wrappedEmptyMessageError,
			exp: `<no message>

at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError (/pkg/bruh/format_test.go:58)
    56│    func wrappedEmptyMessageError() error {
    57│        if err := emptyMessageError(); err != nil {
  → 58│            return bruh.Wrap(err, "")
    59│        }
    60│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStackedSourced (/pkg/bruh/format_bruh_trace_stacked_sourced_test.go:15)
    13│    emptyMessageError := emptyMessageError()
    14│    wrappedError := wrappedError3()
  → 15│    wrappedEmptyMessageError := wrappedEmptyMessageError()
    16│    externalError := externalError()
    17│    externallyWrappedError := externallyWrappedError()
at testing.tRunner (/testing/testing.go:1234)
<no message>

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
    59│        }`,
		},
		{
			name: "External",
			err:  externalError,
			exp:  `external error`,
		},
		{
			name: "WrappedExternal",
			err:  wrappedExternalError,
			exp: `wrapped 1

at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalError (/pkg/bruh/format_test.go:76)
    74│    func wrappedExternalError() error {
    75│        if err := externalError(); err != nil {
  → 76│            return bruh.Wrap(err, "wrapped 1")
    77│        }
    78│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStackedSourced (/pkg/bruh/format_bruh_trace_stacked_sourced_test.go:18)
    16│    externalError := externalError()
    17│    externallyWrappedError := externallyWrappedError()
  → 18│    wrappedExternalError := wrappedExternalError()
    19│    wrappedExternalInterleavedError := wrappedExternalInterleavedError()
    20│    externallyWrappedNilError := externallyWrappedNilError()
at testing.tRunner (/testing/testing.go:1234)
external error`,
		},
		{
			name: "ExternallyWrapped",
			err:  externallyWrappedError,
			exp: `external error
root error

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
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStackedSourced (/pkg/bruh/format_bruh_trace_stacked_sourced_test.go:17)
    15│    wrappedEmptyMessageError := wrappedEmptyMessageError()
    16│    externalError := externalError()
  → 17│    externallyWrappedError := externallyWrappedError()
    18│    wrappedExternalError := wrappedExternalError()
    19│    wrappedExternalInterleavedError := wrappedExternalInterleavedError()
at testing.tRunner (/testing/testing.go:1234)`,
		},
		{
			name: "WrappedExternalInterleaved",
			err:  wrappedExternalInterleavedError,
			exp: `wrapped

at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError (/pkg/bruh/format_test.go:84)
    82│    func wrappedExternalInterleavedError() error {
    83│        if err := externallyWrappedError(); err != nil {
  → 84│            return bruh.Wrap(err, "wrapped")
    85│        }
    86│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStackedSourced (/pkg/bruh/format_bruh_trace_stacked_sourced_test.go:19)
    17│    externallyWrappedError := externallyWrappedError()
    18│    wrappedExternalError := wrappedExternalError()
  → 19│    wrappedExternalInterleavedError := wrappedExternalInterleavedError()
    20│    externallyWrappedNilError := externallyWrappedNilError()
    21│    wrappedGlobalError := wrappedGlobalError()
at testing.tRunner (/testing/testing.go:1234)
external error
root error

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
    85│        }`,
		},
		{
			name: "ExternallyWrappedNil",
			err:  externallyWrappedNilError,
			exp: `wrapped

at github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedNilError (/pkg/bruh/format_test.go:92)
    90│    func externallyWrappedNilError() error {
    91│        if err := fmt.Errorf("external error: %w", nil); err != nil {
  → 92│            return bruh.Wrap(err, "wrapped")
    93│        }
    94│        return nil
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStackedSourced (/pkg/bruh/format_bruh_trace_stacked_sourced_test.go:20)
    18│    wrappedExternalError := wrappedExternalError()
    19│    wrappedExternalInterleavedError := wrappedExternalInterleavedError()
  → 20│    externallyWrappedNilError := externallyWrappedNilError()
    21│    wrappedGlobalError := wrappedGlobalError()
    22│
at testing.tRunner (/testing/testing.go:1234)
external error: %!w(<nil>)`,
		},
		{
			name: "WrappedGlobal",
			err:  wrappedGlobalError,
			exp: `wrapped

at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedGlobalError (/pkg/bruh/format_test.go:99)
     97│    //go:noinline
     98│    func wrappedGlobalError() error {
  →  99│        return bruh.Wrap(globalWrappedError, "wrapped")
    100│    }
    101│
at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStackedSourced (/pkg/bruh/format_bruh_trace_stacked_sourced_test.go:21)
    19│    wrappedExternalInterleavedError := wrappedExternalInterleavedError()
    20│    externallyWrappedNilError := externallyWrappedNilError()
  → 21│    wrappedGlobalError := wrappedGlobalError()
    22│
    23│    tests := []struct {
at testing.tRunner (/testing/testing.go:1234)
globally wrapped
root error`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := bruhTraceSourcedReplacePath(bruh.StringFormat(tc.err, bruh.BruhStackedFancyFormatter(false, true, false)))
			if result != tc.exp {
				t.Errorf("%s, expected:\n|%s|\n\ngot:\n|%s|", tc.name, tc.exp, result)
			}
		})
	}
}
