package bruh_test

import (
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func TestFormatBruhStacked(t *testing.T) {
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
			exp:  "",
		},
		{
			name: "SingleRoot",
			err:  singleRootError,
			exp: `root error
    at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStacked (/pkg/bruh/format_bruh_trace_stacked_test.go:12)
    at testing.tRunner (/testing/testing.go:1234)`,
		},
		{
			name: "EmptyMessage",
			err:  emptyMessageError,
			exp: `<no message>
    at github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError (/pkg/bruh/format_test.go:28)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStacked (/pkg/bruh/format_bruh_trace_stacked_test.go:13)
    at testing.tRunner (/testing/testing.go:1234)`,
		},
		{
			name: "Wrapped",
			err:  wrappedError,
			exp: `wrapped 3
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3 (/pkg/bruh/format_test.go:50)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStacked (/pkg/bruh/format_bruh_trace_stacked_test.go:14)
    at testing.tRunner (/testing/testing.go:1234)
wrapped 2
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError2 (/pkg/bruh/format_test.go:42)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3 (/pkg/bruh/format_test.go:49)
wrapped 1
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError1 (/pkg/bruh/format_test.go:34)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError2 (/pkg/bruh/format_test.go:41)
root error
    at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError1 (/pkg/bruh/format_test.go:33)`,
		},
		{
			name: "WrappedEmptyMessage",
			err:  wrappedEmptyMessageError,
			exp: `<no message>
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError (/pkg/bruh/format_test.go:58)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStacked (/pkg/bruh/format_bruh_trace_stacked_test.go:15)
    at testing.tRunner (/testing/testing.go:1234)
<no message>
    at github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError (/pkg/bruh/format_test.go:28)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError (/pkg/bruh/format_test.go:57)`,
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
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStacked (/pkg/bruh/format_bruh_trace_stacked_test.go:18)
    at testing.tRunner (/testing/testing.go:1234)
external error`,
		},
		{
			name: "ExternallyWrapped",
			err:  externallyWrappedError,
			exp: `external error
root error
    at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedError (/pkg/bruh/format_test.go:70)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStacked (/pkg/bruh/format_bruh_trace_stacked_test.go:17)
    at testing.tRunner (/testing/testing.go:1234)`,
		},
		{
			name: "WrappedExternalInterleaved",
			err:  wrappedExternalInterleavedError,
			exp: `wrapped
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError (/pkg/bruh/format_test.go:84)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStacked (/pkg/bruh/format_bruh_trace_stacked_test.go:19)
    at testing.tRunner (/testing/testing.go:1234)
external error
root error
    at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedError (/pkg/bruh/format_test.go:70)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError (/pkg/bruh/format_test.go:83)`,
		},
		{
			name: "ExternallyWrappedNil",
			err:  externallyWrappedNilError,
			exp: `wrapped
    at github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedNilError (/pkg/bruh/format_test.go:92)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStacked (/pkg/bruh/format_bruh_trace_stacked_test.go:20)
    at testing.tRunner (/testing/testing.go:1234)
external error: %!w(<nil>)`,
		},
		{
			name: "WrappedGlobal",
			err:  wrappedGlobalError,
			exp: `wrapped
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedGlobalError (/pkg/bruh/format_test.go:99)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatBruhStacked (/pkg/bruh/format_bruh_trace_stacked_test.go:21)
    at testing.tRunner (/testing/testing.go:1234)
globally wrapped
root error`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := bruhTraceReplacePath(bruh.StringFormat(tc.err, bruh.BruhStackedFormatter))
			if result != tc.exp {
				t.Errorf("expected:\n|%s|\n\ngot:\n|%s|", tc.exp, result)
			}
		})
	}
}
