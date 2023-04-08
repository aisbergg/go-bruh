package main

import (
	"encoding/json"
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	err := foo()
	ftdErr := bruh.ToCustomString(err, FormatJSON)
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

func FormatJSON(upkErr bruh.UnpackedError) string {
	// convert unpacked error to a list of ErrorStruct
	errStructs := make([]ErrorStruct, len(upkErr))
	for i, err := range upkErr {
		stack := make([]Location, len(err.Stack))
		for j, frame := range err.Stack {
			stack[j] = Location{
				Name:           frame.Name,
				File:           frame.File,
				Line:           frame.Line,
				ProgramCounter: frame.ProgramCounter,
			}
		}
		errStructs[i] = ErrorStruct{
			Message: err.Msg,
			Stack:   stack,
		}
	}
	// serialize to JSON
	json, err := json.MarshalIndent(errStructs, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println("json: ", string(json))
	return string(json)
}

// -----------------------------------------------------------------------------
// Just some functions that return errors
// -----------------------------------------------------------------------------

func foo() error {
	err := bar()
	if err != nil {
		return bruh.Wrapf(err, "foo: failed to read config file")
	}
	return nil
}

func bar() error {
	err := baz()
	if err != nil {
		return bruh.Wrapf(err, "bar: failed to parse")
	}
	return nil
}

func baz() error {
	// external error
	return fmt.Errorf("oh no")
}
