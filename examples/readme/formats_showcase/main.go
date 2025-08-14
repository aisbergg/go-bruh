package main

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	err := configure()

	formats := []bruh.Formatter{
		bruh.BruhFormatter,
		bruh.BruhFancyFormatter(false, true),
		bruh.BruhStackedFormatter,
		bruh.BruhStackedFancyFormatter(false, true, false),
		bruh.GoPanicFormatter,
		bruh.JavaStackTraceFormatter,
		bruh.PythonTracebackFormatter,
	}

	for i, format := range formats {
		fmt.Println(clean(bruh.StringFormat(err, format)))
		if i < len(formats)-1 {
			fmt.Println(
				"\n\n--------------------------------------------------------------------------------",
			)
		}
	}
}

// clean replaces the paths in the formatted error output to allow for consistent testing.
func clean(s string) string {
	_, repoDir, _, _ := runtime.Caller(0)
	repoDir = filepath.Dir(filepath.Dir(filepath.Dir(repoDir))) + "/"
	s = strings.ReplaceAll(s, repoDir, "")
	return s
}

// -----------------------------------------------------------------------------
// Just some functions that return errors
// -----------------------------------------------------------------------------

func configure() error {
	err := decodingData()
	if err != nil {
		return bruh.Wrap(err, "configuring application")
	}
	return nil
}

func decodingData() error {
	err := readFile("example.json")
	if err != nil {
		return bruh.Wrap(err, "decoding data")
	}
	return nil
}

func readFile(path string) error {
	err := io.ErrUnexpectedEOF
	if err != nil {
		return bruh.Wrapf(err, "reading file '%s'", path)
	}
	return nil
}
