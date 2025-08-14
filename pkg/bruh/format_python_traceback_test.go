package bruh_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func TestFormatPythonTraceback(t *testing.T) {
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
			exp: `Traceback (most recent call last):
  File "/testing/testing.go", line 1234, in testing.tRunner
  File "/pkg/bruh/format_python_traceback_test.go", line 14, in github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatPythonTraceback
  File "/pkg/bruh/format_test.go", line 23, in github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError
*bruh.Err: root error`,
		},
		{
			name: "EmptyMessage",
			err:  emptyMessageError,
			exp: `Traceback (most recent call last):
  File "/testing/testing.go", line 1234, in testing.tRunner
  File "/pkg/bruh/format_python_traceback_test.go", line 15, in github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatPythonTraceback
  File "/pkg/bruh/format_test.go", line 28, in github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError
*bruh.Err`,
		},
		{
			name: "Wrapped",
			err:  wrappedError,
			exp: `Traceback (most recent call last):
  File "/pkg/bruh/format_test.go", line 33, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError1
  File "/pkg/bruh/format_test.go", line 23, in github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError
*bruh.Err: root error

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/pkg/bruh/format_test.go", line 41, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError2
  File "/pkg/bruh/format_test.go", line 34, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError1
*bruh.Err: wrapped 1

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/pkg/bruh/format_test.go", line 49, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3
  File "/pkg/bruh/format_test.go", line 42, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError2
*bruh.Err: wrapped 2

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/testing/testing.go", line 1234, in testing.tRunner
  File "/pkg/bruh/format_python_traceback_test.go", line 16, in github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatPythonTraceback
  File "/pkg/bruh/format_test.go", line 50, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3
*bruh.Err: wrapped 3`,
		},
		{
			name: "WrappedEmptyMessage",
			err:  wrappedEmptyMessageError,
			exp: `Traceback (most recent call last):
  File "/pkg/bruh/format_test.go", line 57, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError
  File "/pkg/bruh/format_test.go", line 28, in github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError
*bruh.Err

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/testing/testing.go", line 1234, in testing.tRunner
  File "/pkg/bruh/format_python_traceback_test.go", line 17, in github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatPythonTraceback
  File "/pkg/bruh/format_test.go", line 58, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError
*bruh.Err`,
		},
		{
			name: "External",
			err:  externalError,
			exp:  `*errors.errorString: external error`,
		},
		{
			name: "WrappedExternal",
			err:  wrappedExternalError,
			exp: `*errors.errorString: external error

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/testing/testing.go", line 1234, in testing.tRunner
  File "/pkg/bruh/format_python_traceback_test.go", line 20, in github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatPythonTraceback
  File "/pkg/bruh/format_test.go", line 76, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalError
*bruh.Err: wrapped 1`,
		},
		{
			name: "ExternallyWrapped",
			err:  externallyWrappedError,
			exp: `Traceback (most recent call last):
  File "/testing/testing.go", line 1234, in testing.tRunner
  File "/pkg/bruh/format_python_traceback_test.go", line 19, in github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatPythonTraceback
  File "/pkg/bruh/format_test.go", line 70, in github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedError
  File "/pkg/bruh/format_test.go", line 23, in github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError
*bruh.Err: root error

The above exception was the direct cause of the following exception:

*fmt.wrapError: external error`,
		},
		{
			name: "WrappedExternalInterleaved",
			err:  wrappedExternalInterleavedError,
			exp: `Traceback (most recent call last):
  File "/pkg/bruh/format_test.go", line 83, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError
  File "/pkg/bruh/format_test.go", line 70, in github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedError
  File "/pkg/bruh/format_test.go", line 23, in github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError
*bruh.Err: root error

The above exception was the direct cause of the following exception:

*fmt.wrapError: external error

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/testing/testing.go", line 1234, in testing.tRunner
  File "/pkg/bruh/format_python_traceback_test.go", line 21, in github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatPythonTraceback
  File "/pkg/bruh/format_test.go", line 84, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError
*bruh.Err: wrapped`,
		},
		{
			name: "ExternallyWrappedNil",
			err:  externallyWrappedNilError,
			exp: `*fmt.wrapError: external error: %!w(<nil>)

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/testing/testing.go", line 1234, in testing.tRunner
  File "/pkg/bruh/format_python_traceback_test.go", line 22, in github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatPythonTraceback
  File "/pkg/bruh/format_test.go", line 92, in github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedNilError
*bruh.Err: wrapped`,
		},
		{
			name: "WrappedGlobal",
			err:  wrappedGlobalError,
			exp: `*bruh.Err: root error

The above exception was the direct cause of the following exception:

*bruh.Err: globally wrapped

The above exception was the direct cause of the following exception:

Traceback (most recent call last):
  File "/testing/testing.go", line 1234, in testing.tRunner
  File "/pkg/bruh/format_python_traceback_test.go", line 23, in github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatPythonTraceback
  File "/pkg/bruh/format_test.go", line 99, in github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedGlobalError
*bruh.Err: wrapped`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := pythonTracebackReplacePath(bruh.StringFormat(tc.err, bruh.PythonTracebackFormatter))
			if result != tc.exp {
				t.Errorf("expected:\n|%s|\n\ngot:\n|%s|", tc.exp, result)
			}
		})
	}
}

var pythonTracebackRegexpTestingGo = regexp.MustCompile(`testing.go", line \d+`)

func pythonTracebackReplacePath(s string) string {
	s = strings.ReplaceAll(s, goTestingPath, "")
	s = strings.ReplaceAll(s, repoDir, "")
	s = pythonTracebackRegexpTestingGo.ReplaceAllLiteralString(s, `testing.go", line 1234`)
	s = regexpMemoryAddress.ReplaceAllLiteralString(s, "0x012345")
	return s
}
