package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aisbergg/go-bruh/pkg/bruh"
	"github.com/aisbergg/go-bruh/pkg/ctxerror"
	"github.com/aisbergg/go-bruh/pkg/multierror"
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
		os.Exit(1)
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	}); err != nil {
		fmt.Printf("failed to initialize Sentry: %v\n", err)
		os.Exit(1)
	}
	defer sentry.Flush(time.Second * 5)

	trackWithStackTrace()
	trackWithStackTraceAndContext()
	trackWithMultiError()

	fmt.Println("done")
}

func trackWithStackTrace() {
	fmt.Println("capturing error with stack trace")

	err := loadConfig("config.yml")
	if err != nil {
		// Sentry extracts the stack trace from the error automatically
		sentry.CaptureException(err)
	}
}

func trackWithStackTraceAndContext() {
	fmt.Println("capturing error with stack trace and context")

	err := loadConfig("config.yml")
	if err != nil {
		sentry.WithScope(func(scope *sentry.Scope) {
			// if any error context is available, add it to the Sentry scope.
			scope.SetContexts(ctxerror.GetContext(err))
			scope.SetTags(ctxerror.GetTags(err))
			scope.SetLevel(sentry.LevelError)
			// Sentry extracts the stack trace from the error automatically
			sentry.CaptureException(err)
		})
	}
}

func trackWithMultiError() {
	fmt.Println("capturing multi error with stack trace")

	errs := []error{
		loadConfig("config.yml"),
		produceError(),
	}
	err := multierror.New("multiple errors", multierror.Options{UnwrapBehavior: multierror.UnwrapFirst})
	err.Add(errs...)
	sentry.CaptureException(err)
}

// -----------------------------------------------------------------------------
// Just some functions that return errors
// -----------------------------------------------------------------------------

//go:noinline
func loadConfig(file string) error {
	err := parseConfig("")
	if err != nil {
		// error with extra context
		return ctxerror.Wrap(err, "reading config file").
			SetContext("config", map[string]any{"filename": file})
	}
	return nil
}

//go:noinline
func parseConfig(contents string) error {
	err := produceError()
	if err != nil {
		return bruh.Wrap(err, "parsing config")
	}
	return nil
}

//go:noinline
func produceError() error {
	// external error
	return fmt.Errorf("oh no")
}
