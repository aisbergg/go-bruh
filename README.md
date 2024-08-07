<a name="readme-top"></a>

[![GoDoc](https://pkg.go.dev/badge/github.com/aisbergg/go-bruh)](https://pkg.go.dev/github.com/aisbergg/go-bruh/pkg/bruh)
[![GoReport](https://goreportcard.com/badge/github.com/aisbergg/go-bruh)](https://goreportcard.com/report/github.com/aisbergg/go-bruh)
[![Coverage Status](https://codecov.io/gh/aisbergg/go-bruh/branch/main/graph/badge.svg)](https://codecov.io/gh/aisbergg/go-bruh)
[![CodeQL](https://github.com/aisbergg/go-bruh/actions/workflows/codeql.yml/badge.svg
)](https://github.com/aisbergg/go-bruh/actions/workflows/codeql.yml)
[![License](https://img.shields.io/github/license/aisbergg/go-bruh)](https://pkg.go.dev/github.com/aisbergg/go-bruh)
[![LinkedIn](https://img.shields.io/badge/-LinkedIn-green.svg?logo=linkedin&colorB=555)](https://www.linkedin.com/in/andre-lehmann-97408221a/)

<br />
<br />
<div align="center">
  <a href="https://github.com/aisbergg/go-bruh">
    <img src="assets/logo.svg" alt="Logo" width="160" height="160">
  </a>

  <h2 align="center"><b>bruh - Error Handling</b></h2>

  <p align="center">
    Having a bruh moment? No problem! Handle errors like a pro with <i>bruh</i> - the Go error handling library that simplifies error management and beautifies stack traces.
    <br />
    <br />
    <a href="https://pkg.go.dev/github.com/aisbergg/go-bruh/pkg/bruh">View Docs</a>
    ·
    <a href="https://github.com/aisbergg/go-bruh/issues">Report Bug</a>
    ·
    <a href="https://github.com/aisbergg/go-bruh/issues">Request Feature</a>
  </p>
</div>

<details>
  <summary>Table of Contents</summary>

- [About](#about)
- [Installation](#installation)
- [Synopsis](#synopsis)
  - [Creating Errors](#creating-errors)
  - [Wrapping Errors](#wrapping-errors)
  - [Formatting Errors](#formatting-errors)
  - [Creating Custom Errors](#creating-custom-errors)
- [Benchmark](#benchmark)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)
- [Acknowledgments](#acknowledgments)

</details>



## About

Having a bruh moment? Don't worry, _bruh_ is here to help! _bruh_ is a Go error handling library that makes it easy to deal with errors and print pretty stack traces. Simply create, wrap, and format your errors with all the details you need. Since it is designed as a drop-in replacement for Go's standard library `errors` package, you don't even have to worry about making major changes to your code in order to get it working.

**Features:**

- includes stack traces (also included when captured by [Sentry](https://sentry.io/))
- offers custom error formatting
- allows to create custom errors
- acts as a drop-in replacement for Go's standard library `errors` package

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Installation

```sh
go get github.com/aisbergg/go-bruh
```

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Synopsis

### Creating Errors

Creating new errors with stack traces is done by calling [`bruh.New(msg string)`](https://pkg.go.dev/github.com/aisbergg/go-bruh/pkg/bruh#New) or [`Errorf(format string, args ...any)`](https://pkg.go.dev/github.com/aisbergg/go-bruh/pkg/bruh#Errorf):

```golang
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

// create a global error
var ErrInternalServer = bruh.New("error internal server")

func main() {
	url := "https://foo-bar.local"
	if _, err := Get(url); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func Get(url string) (*http.Response, error) {
	res, err := http.Get(url)
	if err != nil {
		// create a local error
		return nil, bruh.New("request failed")
	}
	if res.StatusCode != http.StatusOK {
		// create a local error using a format
		return nil, bruh.Errorf("request failed with status code %d", res.StatusCode)
	}
	return res, nil
}
```

Outputs:

```plaintext
request failed
    path/to/main.go:26 in main.Get
    path/to/main.go:16 in main.main
```

### Wrapping Errors

Wrapping errors is not different than creating entirely new errors. You can use [`Wrap(err error, msg string)`](https://pkg.go.dev/github.com/aisbergg/go-bruh/pkg/bruh#Wrap) or [`Wrapf(err error, format string, args ...any)`](https://pkg.go.dev/github.com/aisbergg/go-bruh/pkg/bruh#Wrapf) for this:

```golang
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	url := "https://foo-bar.local"
	if _, err := Get(url); err != nil {
		err = bruh.Wrapf(err, "failed to fetch %s", url)
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func Get(url string) (*http.Response, error) {
	res, err := http.Get(url)
	if err != nil {
		// wrap previous error
		return nil, bruh.Wrap(err, "request failed")
	}
	return res, nil
}
```

Outputs:

```plaintext
failed to fetch https://foo-bar.local
    path/to/main.go:17 in main.main
request failed
    path/to/main.go:27 in main.Get
    path/to/main.go:16 in main.main
Get "https://foo-bar.local": dial tcp: lookup foo-bar.local: no such host
```

### Formatting Errors

Errors can be formatted using the built-in formats or customized by creating a custom formatter. No matter what format you use, the basic usage is as such:

```go
package main

import (
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	err := foo()

	// only message
	ftdErr := fmt.Sprintf("%v", err)

	// with trace
	ftdErr = fmt.Sprintf("%+v", err)

	// with custom formatter
	ftdErr = bruh.ToCustomString(err, bruh.FormatWithCombinedTrace)

	fmt.Println(ftdErr)
}

// external error
var extErr = fmt.Errorf("external")

func foo() error {
	return bruh.Wrapf(bar(), "foo")
}

func bar() error {
	return bruh.Wrapf(extErr, "bar")
}
```

Following formats are built-in:

- `bruh.FormatWithTrace`: FormatWithTrace is an error formatter that produces a trace containing a partial stack trace for each wrapped error. E.g.:
    ```plaintext
    <error1>:
        <file1>:<line1> in <function1>
        <file2>:<line2> in <function2>
        <fileN>:<lineN> in <functionN>
    <error2>:
        <file1>:<line1> in <function1>
        <file2>:<line2> in <function2>
        <fileN>:<lineN> in <functionN>
    <errorN>:
        <file1>:<line1> in <function1>
        <file2>:<line2> in <function2>
        <fileN>:<lineN> in <functionN>
    ```

- `bruh.FormatWithCombinedTrace`: FormatWithCombinedTrace is an error formatter that produces a single combined stack trace. E.g.:
    ```plaintext
    <error1>: <error2>: <errorN>
        <file1>:<line1> in <function1>
        <file2>:<line2> in <function2>
        <fileN>:<lineN> in <functionN>
    ```

- `bruh.FormatPythonTraceback`: FormatPythonTraceback is an error formatter that produces error traces similar to Python's tracebacks. E.g.:
    ```plaintext
    Traceback (most recent call last):
      File "<file3>", line <line3>, in <function3>
      File "<file2>", line <line2>, in <function2>
      File "<file1>", line <line1>, in <function1>
    <typeName1>: <error1>

    The above exception was the direct cause of the following exception:

    Traceback (most recent call last):
      File "<file3>", line <line3>, in <function3>
      File "<file2>", line <line2>, in <function2>
      File "<file1>", line <line1>, in <function1>
    <typeName2>: <error2>
    ```

You can customize the format by creating your own formatter. Check the [json example](examples/formatting/json.go) on how to accomplish that.

### Creating Custom Errors

Custom errors can be created based on the bruh standard error. The custom error will inherit the properties of the bruh error, such as the trace backs. Here is a short example:

> Note: You can find a more detailed example in the [examples directory](examples/customizing/timestamped_error.go)

```go
package main

import (
	"fmt"
	"time"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	err := fmt.Errorf("external error")
	err = TEWrapf(err, "%s, what a day", "bruh")
	if terr, ok := err.(*TimestampedError); ok {
		fmt.Printf("%s: %s\n", terr.Timestamp().Format(time.RFC3339), terr.Error())
	}
}

// TimestampedError represents the custom error. It embeds bruh.TraceableError
// and "inherits" its properties. This way, you can create your own custom types
// and add more properties as needed.
type TimestampedError struct {
	bruh.TraceableError
	timestamp time.Time
}

// TEWrapf creates the custom error. As you can see, it initializes the
// underlying bruh error using the `WrapfSkip` function. For each of the
// standard error creation function, there is a `Skip` equivalent. These
// initialize the bruh error and make sure, that this very function does not
// appear in the stack trace.
func TEWrapf(err error, format string, args ...any) error {
	return &TimestampedError{
		// skip is required to skip the current function and thus exclude this
		// function from the stack trace
		TraceableError: *bruh.WrapfSkip(err, 1, format, args...),
		timestamp:      time.Now(),
	}
}

func (te *TimestampedError) Timestamp() time.Time {
	return te.timestamp
}
```

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Benchmark

Inside the `benchmark` directory reside some comparable benchmarks that allow some performance comparison of bruh with other error handling libraries. The benchmarks can be executed by running `make bench`. Here are my results:

```
cpu: Intel(R) Core(TM) i7-8550U CPU @ 1.80GHz
BenchmarkWrap/pkg=bruh/layers=1-4                 752203              1395 ns/op             648 B/op          5 allocs/op
BenchmarkWrap/pkg=pkgerrors/layers=1-4            814128              1465 ns/op             648 B/op          8 allocs/op
BenchmarkWrap/pkg=eris/layers=1-4                 276560              3665 ns/op            2048 B/op         18 allocs/op
BenchmarkWrap/pkg=bruh/layers=10-4                148750              7318 ns/op            3600 B/op         32 allocs/op
BenchmarkWrap/pkg=pkgerrors/layers=10-4           125960              9040 ns/op            3744 B/op         53 allocs/op
BenchmarkWrap/pkg=eris/layers=10-4                 60559             18084 ns/op            9537 B/op         81 allocs/op
BenchmarkWrap/pkg=bruh/layers=100-4                19662             62340 ns/op           33126 B/op        302 allocs/op
BenchmarkWrap/pkg=pkgerrors/layers=100-4           14648             84568 ns/op           34710 B/op        503 allocs/op
BenchmarkWrap/pkg=eris/layers=100-4                 5730            178413 ns/op           84431 B/op        711 allocs/op
BenchmarkFormatWithoutTrace/pkg=bruh/layers=1-4                  6034773               179.4 ns/op          1024 B/op          1 allocs/op
BenchmarkFormatWithoutTrace/pkg=pkgerrors/layers=1-4             7391492               153.3 ns/op            32 B/op          2 allocs/op
BenchmarkFormatWithoutTrace/pkg=eris/layers=1-4                   779587              1530 ns/op             832 B/op          9 allocs/op
BenchmarkFormatWithoutTrace/pkg=bruh/layers=10-4                 4275469               278.8 ns/op          1024 B/op          1 allocs/op
BenchmarkFormatWithoutTrace/pkg=pkgerrors/layers=10-4            1630375               700.7 ns/op           648 B/op         11 allocs/op
BenchmarkFormatWithoutTrace/pkg=eris/layers=10-4                  183391              6237 ns/op            7136 B/op         54 allocs/op
BenchmarkFormatWithoutTrace/pkg=bruh/layers=100-4                 822076              1259 ns/op            1024 B/op          1 allocs/op
BenchmarkFormatWithoutTrace/pkg=pkgerrors/layers=100-4             81950             14245 ns/op           48432 B/op        101 allocs/op
BenchmarkFormatWithoutTrace/pkg=eris/layers=100-4                  10794            107690 ns/op          382554 B/op        504 allocs/op
BenchmarkFormatWithTrace/pkg=bruh/layers=1-4                      348889              3600 ns/op            4745 B/op          9 allocs/op
BenchmarkFormatWithTrace/pkg=pkgerrors/layers=1-4                 162422              6887 ns/op            1168 B/op         19 allocs/op
BenchmarkFormatWithTrace/pkg=eris/layers=1-4                      198182              5363 ns/op            4457 B/op         51 allocs/op
BenchmarkFormatWithTrace/pkg=bruh/layers=10-4                      69140             16209 ns/op           11529 B/op         27 allocs/op
BenchmarkFormatWithTrace/pkg=pkgerrors/layers=10-4                 30410             46021 ns/op            6951 B/op        100 allocs/op
BenchmarkFormatWithTrace/pkg=eris/layers=10-4                      64896             16679 ns/op           21166 B/op        159 allocs/op
BenchmarkFormatWithTrace/pkg=bruh/layers=100-4                      7196            145970 ns/op           97930 B/op        209 allocs/op
BenchmarkFormatWithTrace/pkg=pkgerrors/layers=100-4                 2107            568167 ns/op           64751 B/op        910 allocs/op
BenchmarkFormatWithTrace/pkg=eris/layers=100-4                      3900            302192 ns/op         1044113 B/op       1241 allocs/op
```

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Contributing

If you have any suggestions, want to file a bug report or want to contribute to this project in some other way, please read the [contribution guideline](CONTRIBUTING.md).

And don't forget to give this project a star 🌟! Thanks again!

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Contact

André Lehmann

- Email: aisberg@posteo.de
- [GitHub](https://github.com/aisbergg)
- [LinkedIn](https://www.linkedin.com/in/andre-lehmann-97408221a/)

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Acknowledgments

_bruh_ was originally inspired by [eris](https://github.com/rotisserie/eris). I started of with eris's code base, but heavily modified it. _bruh_ still borrows some tests and other places of the code might still resemble pieces of eris, so shoutout to the maintainers of eris!

The logo is a derivative of the [logo](https://github.com/rfyiamcool/golang_logo/blob/master/svg/golang_3.svg) by [rfyiamcool](https://github.com/rfyiamcool). Sorry rfyiamcool that I butchered your gopher.

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>
