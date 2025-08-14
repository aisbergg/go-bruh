package main

import (
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	err := foo()

	// message only
	f := bruh.Message(err)
	// or simply
	f = err.Error()

	// with trace (default format)
	f = bruh.String(err)

	// with custom format
	f = bruh.StringFormat(err, bruh.BruhStackedFormatter)

	fmt.Println(f)
}

// external error
var extErr = fmt.Errorf("external")

func foo() error {
	return bruh.Wrapf(bar(), "foo")
}

func bar() error {
	return bruh.Wrapf(extErr, "bar")
}
