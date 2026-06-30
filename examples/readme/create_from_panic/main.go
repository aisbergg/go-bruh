package main

import (
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			err := bruh.NewFromPanic(r)
			fmt.Printf("%s\n", bruh.StringFormat(err, bruh.BruhFormatter))
		}
	}()

	fmt.Printf("result: %d\n", Divide(10, 0))

	// this will never be reached because the panic will be caught by the defer
	fmt.Println("done")
}

func Divide(a, b int) int {
	if b == 0 {
		panic("division by zero")
	}
	return a / b
}
