package main

import (
	"fmt"
	"runtime"

	"github.com/aisbergg/go-bruh/pkg/bruh"
)

type MyError struct {
	msg     string
	callers []uintptr
}

func New(msg string) error {
	callers := make([]uintptr, 32)
	return &MyError{
		msg:     msg,
		callers: callers[:runtime.Callers(1, callers)],
	}
}

func (e *MyError) Error() string {
	return e.msg
}

func (e *MyError) Callers() []uintptr {
	return e.callers
}

func main() {
	err := New("root error")
	err = bruh.Wrap(err, "wrapped")
	fmt.Println(bruh.StringFormat(err, bruh.BruhStackedFormatter))
}
