package bruh_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func TestFormatGoPanic(t *testing.T) {
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

	assertGoPanic := func(name string, err error, exp string) {
		t.Run(name, func(t *testing.T) {
			result := goPanicReplacePath(bruh.StringFormat(err, bruh.GoPanicFormatter))
			if result != exp {
				t.Errorf("expected:\n|%s|\n\ngot:\n|%s|", exp, result)
			}
		})
	}

	assertGoPanic("Nil", nil, "")

	assertGoPanic("SingleRoot", singleRootError, `root error

github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError()
	/pkg/bruh/format_test.go:23 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatGoPanic()
	/pkg/bruh/format_go_panic_test.go:14 +0x012345
testing.tRunner()
	/testing/testing.go:1234 +0x012345`)

	assertGoPanic("EmptyMessage", emptyMessageError, `github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError()
	/pkg/bruh/format_test.go:28 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatGoPanic()
	/pkg/bruh/format_go_panic_test.go:15 +0x012345
testing.tRunner()
	/testing/testing.go:1234 +0x012345`)

	assertGoPanic("Wrapped", wrappedError, `wrapped 3: wrapped 2: wrapped 1: root error

github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError()
	/pkg/bruh/format_test.go:23 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError1()
	/pkg/bruh/format_test.go:33 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError1()
	/pkg/bruh/format_test.go:34 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError2()
	/pkg/bruh/format_test.go:41 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError2()
	/pkg/bruh/format_test.go:42 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3()
	/pkg/bruh/format_test.go:49 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedError3()
	/pkg/bruh/format_test.go:50 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatGoPanic()
	/pkg/bruh/format_go_panic_test.go:16 +0x012345
testing.tRunner()
	/testing/testing.go:1234 +0x012345`)

	assertGoPanic(
		"WrappedEmptyMessage",
		wrappedEmptyMessageError,
		`github.com/aisbergg/go-bruh/pkg/bruh_test.emptyMessageError()
	/pkg/bruh/format_test.go:28 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError()
	/pkg/bruh/format_test.go:57 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedEmptyMessageError()
	/pkg/bruh/format_test.go:58 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatGoPanic()
	/pkg/bruh/format_go_panic_test.go:17 +0x012345
testing.tRunner()
	/testing/testing.go:1234 +0x012345`,
	)

	assertGoPanic("External", externalError, `external error`)

	assertGoPanic("WrappedExternal", wrappedExternalError, `wrapped 1: external error

github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalError()
	/pkg/bruh/format_test.go:76 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatGoPanic()
	/pkg/bruh/format_go_panic_test.go:20 +0x012345
testing.tRunner()
	/testing/testing.go:1234 +0x012345`)

	assertGoPanic("ExternallyWrapped", externallyWrappedError, `external error: root error

github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError()
	/pkg/bruh/format_test.go:23 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedError()
	/pkg/bruh/format_test.go:70 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatGoPanic()
	/pkg/bruh/format_go_panic_test.go:19 +0x012345
testing.tRunner()
	/testing/testing.go:1234 +0x012345`)

	assertGoPanic("WrappedExternalInterleaved", wrappedExternalInterleavedError, `wrapped: external error: root error

github.com/aisbergg/go-bruh/pkg/bruh_test.singleRootError()
	/pkg/bruh/format_test.go:23 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedError()
	/pkg/bruh/format_test.go:70 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError()
	/pkg/bruh/format_test.go:83 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedExternalInterleavedError()
	/pkg/bruh/format_test.go:84 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatGoPanic()
	/pkg/bruh/format_go_panic_test.go:21 +0x012345
testing.tRunner()
	/testing/testing.go:1234 +0x012345`)

	assertGoPanic("ExternallyWrappedNil", externallyWrappedNilError, `wrapped: external error: %!w(<nil>)

github.com/aisbergg/go-bruh/pkg/bruh_test.externallyWrappedNilError()
	/pkg/bruh/format_test.go:92 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatGoPanic()
	/pkg/bruh/format_go_panic_test.go:22 +0x012345
testing.tRunner()
	/testing/testing.go:1234 +0x012345`)

	assertGoPanic("WrappedGlobal", wrappedGlobalError, `wrapped: globally wrapped: root error

github.com/aisbergg/go-bruh/pkg/bruh_test.wrappedGlobalError()
	/pkg/bruh/format_test.go:99 +0x012345
github.com/aisbergg/go-bruh/pkg/bruh_test.TestFormatGoPanic()
	/pkg/bruh/format_go_panic_test.go:23 +0x012345
testing.tRunner()
	/testing/testing.go:1234 +0x012345`)
}

var goPanicRegexpTestingGo = regexp.MustCompile(`testing\.go:\d+`)

func goPanicReplacePath(s string) string {
	s = strings.ReplaceAll(s, goTestingPath, "")
	s = strings.ReplaceAll(s, repoDir, "")
	s = goPanicRegexpTestingGo.ReplaceAllLiteralString(s, "testing.go:1234")
	s = regexpMemoryAddress.ReplaceAllLiteralString(s, "0x012345")
	return s
}
