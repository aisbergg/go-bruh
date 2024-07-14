package bruh_test

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

var (
	ErrUnexpectedEOF          = bruh.New("unexpected EOF")
	FormattedErrUnexpectedEOF = bruh.Errorf("unexpected %v", "EOF")
)

// Demonstrates string formatting of wrapped errors that originate from local
// root errors (created at the source of the error via New).
func ExampleToString_local() {
	// examples functions that return some errors
	readFile := func(fname string) error {
		return bruh.New("file not found")
	}
	parseFile := func(fname string) error {
		err := readFile(fname)
		if err != nil {
			return bruh.Wrapf(err, "error reading file '%v'", fname)
		}
		return nil
	}

	// call the example functions
	err := parseFile("example.json")

	// print the error via fmt.Printf
	fmt.Printf("%v\n", err) // %v: omit stack trace

	// unpack and print the error via uerr.ToString(...)
	fmt.Println(rplPth(bruh.ToString(err, true))) // include stack trace; you can also use fmt.Printf("%+v\n", err)

	// Output:
	// error reading file 'example.json': file not found
	// error reading file 'example.json'
	//     .../examples_test.go:29 in github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleToString_local.func2
	//     .../examples_test.go:35 in github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleToString_local
	// file not found
	//     .../examples_test.go:24 in github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleToString_local.func1
	//     .../examples_test.go:27 in github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleToString_local.func2
}

// Demonstrates string formatting of wrapped errors that originate from external (non-bruh) error
// types.
func ExampleToString_external() {
	// example func that returns an IO error
	readFile := func(fname string) error {
		return io.ErrUnexpectedEOF
	}

	// unpack and print the error
	err := readFile("example.json")
	fmt.Println(bruh.ToString(err, false)) // false: omit stack trace

	// Output:
	// unexpected EOF
}

// Demonstrates string formatting of wrapped errors that originate from global root errors.
func ExampleToString_global() {
	// example func that wraps a global error value
	readFile := func(fname string) error {
		return bruh.Wrapf(FormattedErrUnexpectedEOF, "error reading file '%v'", fname)
	}

	// example func that catches and returns an error without modification
	parseFile := func(fname string) error {
		// read the file
		err := readFile(fname)
		if err != nil {
			return err
		}
		return nil
	}

	// example func that just catches and returns an error
	processFile := func(fname string) error {
		// parse the file
		err := parseFile(fname)
		if err != nil {
			return bruh.Wrapf(err, "error processing file '%v'", fname)
		}
		return nil
	}

	// call processFile and catch the error
	err := processFile("example.json")

	// print the error via fmt.Printf
	fmt.Printf("%v\n", err) // %v: omit stack trace

	// unpack and print the error via uerr.ToString(...)
	fmt.Printf("%v\n", rplPth(bruh.ToString(err, true))) // true: include stack trace

	// Output:
	// error processing file 'example.json': error reading file 'example.json': unexpected EOF
	// error processing file 'example.json'
	//     .../examples_test.go:91 in github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleToString_global.func3
	//     .../examples_test.go:97 in github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleToString_global
	// error reading file 'example.json'
	//     .../examples_test.go:73 in github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleToString_global.func1
	//     .../examples_test.go:79 in github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleToString_global.func2
	//     .../examples_test.go:89 in github.com/aisbergg/go-bruh/pkg/bruh_test.ExampleToString_global.func3
	// unexpected EOF
}

var (
	regexpRemoveTestMain  = regexp.MustCompile(` *_testmain\.go.*\n`)
	regexpRemoveTestingGo = regexp.MustCompile(` *.../(testing|example|run_example)\.go.*\n`)
)

// rplPth replaces the paths in the formatted error output to allow for consistent testing.
func rplPth(s string) string {
	_, thisFilePath, _, _ := runtime.Caller(0)
	_, goTestingPath, _, _ := runtime.Caller(2)
	goTestingPath = filepath.Dir(goTestingPath)
	s = strings.ReplaceAll(s, goTestingPath, "...")
	s = strings.ReplaceAll(s, thisFilePath, ".../examples_test.go")
	s = regexpRemoveTestMain.ReplaceAllLiteralString(s, "")
	s = regexpRemoveTestingGo.ReplaceAllLiteralString(s, "")
	return s
}
