package main

import (
	"fmt"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func main() {
	// limits the number of stack frames for serialization to 10
	bruh.MaxChainStackDepth = 10

	err := recursiveError(50)
	fmt.Println(bruh.StringFormat(err, bruh.BruhFormatter))
}

func recursiveError(depth int) error {
	if depth <= 0 {
		return bruh.New("base error")
	}
	return bruh.Wrap(recursiveError(depth-1), "")
}
