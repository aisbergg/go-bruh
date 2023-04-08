<a name="readme-top"></a>

[![GoDoc](https://pkg.go.dev/badge/github.com/aisbergg/go-bruh)](https://pkg.go.dev/github.com/aisbergg/go-bruh/pkg/bruh) 
[![GoReport](https://goreportcard.com/badge/github.com/aisbergg/go-bruh)](https://goreportcard.com/report/github.com/aisbergg/go-bruh) 
[![Coverage Status](https://codecov.io/gh/aisbergg/go-bruh/branch/main/graph/badge.svg)](https://codecov.io/gh/aisbergg/go-bruh)
[![License](https://img.shields.io/github/license/aisbergg/go-bruh)](https://pkg.go.dev/github.com/aisbergg/go-bruh) 
[![LinkedIn](https://img.shields.io/badge/-LinkedIn-green.svg?logo=linkedin&colorB=555)](https://www.linkedin.com/in/andre-lehmann-97408221a/)

<br />
<br />
<div align="center">
  <a href="https://github.com/aisbergg/go-bruh">
    <img src="assets/logo.svg" alt="Logo" width="160" height="160">
  </a>

  <h2 align="center"><b>bruh</b></h2>

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

- [About The Project](#about-the-project)
- [Installation](#installation)
- [Synopsis](#synopsis)
  - [Creating Errors](#creating-errors)
  - [Wrapping Errors](#wrapping-errors)
  - [Formatting Errors](#formatting-errors)
  - [Creating Custom Errors](#creating-custom-errors)
- [Benchmark](#benchmark)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)
- [Acknowledgments](#acknowledgments)

</details>



<!-- ABOUT THE PROJECT -->
## About The Project

Having a bruh moment? Don't worry, _bruh_ is here to help! _bruh_ is a Go error handling library that makes it easy to deal with errors and print pretty stack traces. Simply create, wrap, and format your errors with all the details you need. Since it is designed as a drop-in replacement for Go's standard library `errors` package, you don't even have to worry about making major changes to your code in order to get it working.

**Features:**

- includes stack traces
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
// create a global error
var ErrInternalServer = bruh.New("error internal server")

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

### Wrapping Errors

Wrapping errors is not different than creating entirely new errors. You can use [`Wrap(err error, msg string)`](https://pkg.go.dev/github.com/aisbergg/go-bruh/pkg/bruh#Wrap) or [`Wrapf(err error, format string, args ...any)`](https://pkg.go.dev/github.com/aisbergg/go-bruh/pkg/bruh#Wrapf) for this:

```golang
func main() {
	url := "https://example.com"
	_, err := Get(url)
	if err != nil {
		err = bruh.Wrapf(err, "failed to fetch %s", url)
		panic(err)
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

The formatting can be customized by creating a formatter. Check the [json example](examples/formatting/json.go) on how create your very own formatter.

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
// and "inherits" is properties. This way, you can create your own custom types
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
func TEWrapf(err error, format string, args ...interface{}) error {
	return &TimestampedError{
		// skip is required to skip the current function and thus exclude this
		// function from the stack trace
		TraceableError: *bruh.WrapfSkip(err, 1, format, args...).(*bruh.TraceableError),
		timestamp:      time.Now(),
	}
}

func (te *TimestampedError) Timestamp() time.Time {
	return te.timestamp
}
```

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>


## Benchmark

Inside the `benchmark` directory reside some comparable benchmarks that allow some performance comparison of bruh with other error handling libraries. The benchmarks can be executed as follows:

```sh
cd benchmark
go mod download
go test -run '^$' -bench=. -benchmem ./bench_test.go
```

Here are my results:

```
cpu: AMD Ryzen 5 5600X 6-Core Processor             
BenchmarkWrap/std_errors_1_layers-12             7593969               159.7 ns/op            72 B/op          3 allocs/op
BenchmarkWrap/pkg_errors_1_layers-12             1000000              1123 ns/op             648 B/op          8 allocs/op
BenchmarkWrap/eris_1_layers-12                    486070              2476 ns/op            2072 B/op         18 allocs/op
BenchmarkWrap/bruh_1_layers-12                    449115              2425 ns/op            1096 B/op          7 allocs/op
BenchmarkWrap/std_errors_10_layers-12             792187              1455 ns/op             968 B/op         21 allocs/op
BenchmarkWrap/pkg_errors_10_layers-12             175286              6755 ns/op            3745 B/op         53 allocs/op
BenchmarkWrap/eris_10_layers-12                    90980             12604 ns/op            9564 B/op         81 allocs/op
BenchmarkWrap/bruh_10_layers-12                    86458             13995 ns/op            6066 B/op         43 allocs/op
BenchmarkWrap/std_errors_100_layers-12             66074             18258 ns/op           51265 B/op        201 allocs/op
BenchmarkWrap/pkg_errors_100_layers-12             19021             63471 ns/op           34720 B/op        503 allocs/op
BenchmarkWrap/eris_100_layers-12                    8143            137947 ns/op           84480 B/op        711 allocs/op
BenchmarkWrap/bruh_100_layers-12                    8653            128263 ns/op           55770 B/op        403 allocs/op
BenchmarkFormatWithoutTrace/std_errors_1_layers-12              18859053                62.58 ns/op           24 B/op          1 allocs/op
BenchmarkFormatWithoutTrace/pkg_errors_1_layers-12              11473681               103.5 ns/op            48 B/op          2 allocs/op
BenchmarkFormatWithoutTrace/eris_1_layers-12                     1374807               871.0 ns/op           840 B/op          9 allocs/op
BenchmarkFormatWithoutTrace/bruh_1_layers-12                    11404887               104.4 ns/op          1024 B/op          1 allocs/op
BenchmarkFormatWithoutTrace/std_errors_10_layers-12             16593740                69.05 ns/op           96 B/op          1 allocs/op
BenchmarkFormatWithoutTrace/pkg_errors_10_layers-12              2795234               430.7 ns/op           728 B/op         11 allocs/op
BenchmarkFormatWithoutTrace/eris_10_layers-12                     329241              3676 ns/op            7136 B/op         54 allocs/op
BenchmarkFormatWithoutTrace/bruh_10_layers-12                    4943726               243.3 ns/op          1024 B/op          1 allocs/op
BenchmarkFormatWithoutTrace/std_errors_100_layers-12             8941819               133.0 ns/op          1024 B/op          1 allocs/op
BenchmarkFormatWithoutTrace/pkg_errors_100_layers-12              154872              7749 ns/op           49070 B/op        101 allocs/op
BenchmarkFormatWithoutTrace/eris_100_layers-12                     20685             57930 ns/op          381146 B/op        504 allocs/op
BenchmarkFormatWithoutTrace/bruh_100_layers-12                    775707              1479 ns/op            1024 B/op          1 allocs/op
BenchmarkFormatWithTrace/pkg_errors_1_layers-12                   265171              4246 ns/op            1168 B/op         19 allocs/op
BenchmarkFormatWithTrace/eris_1_layers-12                         411547              2648 ns/op            4658 B/op         51 allocs/op
BenchmarkFormatWithTrace/bruh_1_layers-12                         420058              2512 ns/op            5363 B/op         16 allocs/op
BenchmarkFormatWithTrace/pkg_errors_10_layers-12                   48283             25116 ns/op            6178 B/op        100 allocs/op
BenchmarkFormatWithTrace/eris_10_layers-12                        139959              8799 ns/op           22284 B/op        159 allocs/op
BenchmarkFormatWithTrace/bruh_10_layers-12                         99267             11999 ns/op           17145 B/op         79 allocs/op
BenchmarkFormatWithTrace/pkg_errors_100_layers-12                   3628            313512 ns/op           57854 B/op        910 allocs/op
BenchmarkFormatWithTrace/eris_100_layers-12                         8454            133356 ns/op         1079221 B/op       1240 allocs/op
BenchmarkFormatWithTrace/bruh_100_layers-12                        10000            107189 ns/op          150202 B/op        711 allocs/op
```

## Roadmap

- [ ] Add more Tests
- [ ] Add pre-commit hooks
- [ ] Add some automation
- [x] Write more documentation

See the [open issues](https://github.com/aisbergg/go-bruh/issues) for a full list of proposed features (and known issues).

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Contributing

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue and tell me what you have in mind.
Don't forget to give the project a star 🌟! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat: add AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Contact

André Lehmann - aisberg@posteo.de

Project Link: [https://github.com/aisbergg/go-bruh](https://github.com/aisbergg/go-bruh)

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Acknowledgments

_bruh_ was originally inspired by [eris](https://github.com/rotisserie/eris). I started of with eris's code base, but heavily modified it. _bruh_ still borrows some tests and other places of the code might still resemble pieces of eris, so shoutout to the maintainers of eris!

The logo is a derivative of the [logo](https://github.com/rfyiamcool/golang_logo/blob/master/svg/golang_3.svg) by [rfyiamcool](https://github.com/rfyiamcool). Sorry rfyiamcool that I butchered your gopher.

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>
