package main

import (
	"encoding/json"
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	err := foo()
	ftdErr := bruh.StringFormat(err, JsonFormatter)
	fmt.Println(ftdErr)
}

// -----------------------------------------------------------------------------
// Custom error formatter
// -----------------------------------------------------------------------------

type ErrorStruct struct {
	Message string     `json:"message"`
	Stack   []Location `json:"stack"`
}

type Location struct {
	Name           string  `json:"name"`
	File           string  `json:"file"`
	Line           int     `json:"line"`
	ProgramCounter uintptr `json:"pc"`
}

func JsonFormatter(b []byte, unpacker *bruh.Unpacker) []byte {
	// convert unpacked error to a list of ErrorStruct
	upkErr := unpacker.Unpack()
	errStructs := make([]ErrorStruct, len(upkErr))
	for i, err := range upkErr {
		stack := make([]Location, len(err.Stack))
		for j, frame := range err.Stack {
			stack[j] = Location{
				Name:           frame.Name,
				File:           frame.File,
				Line:           frame.Line,
				ProgramCounter: frame.ProgramCounter2,
			}
		}
		errStructs[i] = ErrorStruct{
			Message: err.Msg,
			Stack:   stack,
		}
	}
	// serialize to JSON
	serialized, err := json.MarshalIndent(errStructs, "", "  ")
	if err != nil {
		panic(err)
	}
	if b == nil {
		return serialized
	}
	b = append(b, serialized...)
	return b
}

// -----------------------------------------------------------------------------
// Just some functions that return errors
// -----------------------------------------------------------------------------

func foo() error {
	err := bar()
	if err != nil {
		return bruh.Wrapf(err, "foo: reading config file")
	}
	return nil
}

func bar() error {
	err := baz()
	if err != nil {
		return bruh.Wrapf(err, "bar: parsing stuff")
	}
	return nil
}

func baz() error {
	// external error
	return fmt.Errorf("oh no")
}
