package main

import (
	"flag"
	"fmt"
	"syscall"
	"time"

	"github.com/aisbergg/go-bruh/pkg/bruh"
	"github.com/getsentry/sentry-go"
)

var dsn string

func init() {
	flag.StringVar(&dsn, "dsn", "", "Sentry DSN for logging stack traces")
}

func main() {
	flag.Parse()
	if dsn == "" {
		fmt.Println("missing Sentry DSN, use -dsn flag to set it")
		syscall.Exit(1)
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
	if err != nil {
		fmt.Printf("failed to initialize Sentry: %v\n", err)
		syscall.Exit(1)
	}

	err = foo()
	if err != nil {
		fmt.Println("capturing error")
		sentry.CaptureException(err)
	}

	sentry.Flush(time.Second * 5)
	fmt.Println("done")
}

func foo() error {
	err := bar()
	if err != nil {
		return bruh.Wrap(err, "foo: failed to read config file")
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
