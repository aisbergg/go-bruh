package bruh_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func TestFormatJavaStackTrace(t *testing.T) {
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
			exp: `*bruh.Err: root error
    at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatJavaStackTrace (/pkg/bruh/format_java_stacktrace_test.go:14)
    at testing.tRunner (/testing/testing.go:1234)`,
		},
		{
			name: "EmptyMessage",
			err:  emptyMessageError,
			exp: `*bruh.Err: _
    at github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError (/pkg/bruh/format_test.go:28)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatJavaStackTrace (/pkg/bruh/format_java_stacktrace_test.go:15)
    at testing.tRunner (/testing/testing.go:1234)`,
		},
		{
			name: "Wrapped",
			err:  wrappedError,
			exp: `*bruh.Err: wrapped 3
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3 (/pkg/bruh/format_test.go:50)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatJavaStackTrace (/pkg/bruh/format_java_stacktrace_test.go:16)
    at testing.tRunner (/testing/testing.go:1234)
Caused by: *bruh.Err: wrapped 2
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError2 (/pkg/bruh/format_test.go:42)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3 (/pkg/bruh/format_test.go:49)
Caused by: *bruh.Err: wrapped 1
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError1 (/pkg/bruh/format_test.go:34)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError2 (/pkg/bruh/format_test.go:41)
Caused by: *bruh.Err: root error
    at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError1 (/pkg/bruh/format_test.go:33)`,
		},
		{
			name: "WrappedEmptyMessage",
			err:  wrappedEmptyMessageError,
			exp: `*bruh.Err: _
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError (/pkg/bruh/format_test.go:58)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatJavaStackTrace (/pkg/bruh/format_java_stacktrace_test.go:17)
    at testing.tRunner (/testing/testing.go:1234)
Caused by: *bruh.Err: _
    at github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError (/pkg/bruh/format_test.go:28)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError (/pkg/bruh/format_test.go:57)`,
		},
		{
			name: "External",
			err:  externalError,
			exp:  `*errors.errorString: external error`,
		},
		{
			name: "WrappedExternal",
			err:  wrappedExternalError,
			exp: `*bruh.Err: wrapped 1
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalError (/pkg/bruh/format_test.go:76)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatJavaStackTrace (/pkg/bruh/format_java_stacktrace_test.go:20)
    at testing.tRunner (/testing/testing.go:1234)
Caused by: *errors.errorString: external error`,
		},
		{
			name: "ExternallyWrapped",
			err:  externallyWrappedError,
			exp: `*fmt.wrapError: external error
Caused by: *bruh.Err: root error
    at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedError (/pkg/bruh/format_test.go:70)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatJavaStackTrace (/pkg/bruh/format_java_stacktrace_test.go:19)
    at testing.tRunner (/testing/testing.go:1234)`,
		},
		{
			name: "WrappedExternalInterleaved",
			err:  wrappedExternalInterleavedError,
			exp: `*bruh.Err: wrapped
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError (/pkg/bruh/format_test.go:84)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatJavaStackTrace (/pkg/bruh/format_java_stacktrace_test.go:21)
    at testing.tRunner (/testing/testing.go:1234)
Caused by: *fmt.wrapError: external error
Caused by: *bruh.Err: root error
    at github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError (/pkg/bruh/format_test.go:23)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedError (/pkg/bruh/format_test.go:70)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError (/pkg/bruh/format_test.go:83)`,
		},
		{
			name: "ExternallyWrappedNil",
			err:  externallyWrappedNilError,
			exp: `*bruh.Err: wrapped
    at github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedNilError (/pkg/bruh/format_test.go:92)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatJavaStackTrace (/pkg/bruh/format_java_stacktrace_test.go:22)
    at testing.tRunner (/testing/testing.go:1234)
Caused by: *fmt.wrapError: external error: %!w(<nil>)`,
		},
		{
			name: "WrappedGlobal",
			err:  wrappedGlobalError,
			exp: `*bruh.Err: wrapped
    at github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedGlobalError (/pkg/bruh/format_test.go:99)
    at github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatJavaStackTrace (/pkg/bruh/format_java_stacktrace_test.go:23)
    at testing.tRunner (/testing/testing.go:1234)
Caused by: *bruh.Err: globally wrapped
Caused by: *bruh.Err: root error`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := javaStackTraceReplacePath(bruh.StringFormat(tc.err, bruh.JavaStackTraceFormatter))
			if result != tc.exp {
				t.Errorf("expected:\n|%s|\n\ngot:\n|%s|", tc.exp, result)
			}
		})
	}
}

var javaStackTraceRegexpTestingGo = regexp.MustCompile(`testing\.go:\d+`)

func javaStackTraceReplacePath(s string) string {
	s = strings.ReplaceAll(s, goTestingPath, "")
	s = strings.ReplaceAll(s, repoDir, "")
	s = javaStackTraceRegexpTestingGo.ReplaceAllLiteralString(s, "testing.go:1234")
	s = regexpMemoryAddress.ReplaceAllLiteralString(s, "0x012345")
	return s
}
