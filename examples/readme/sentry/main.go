package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aisbergg/go-bruh/pkg/ctxerror"
	"github.com/getsentry/sentry-go"
)

var dsn string

func init() {
	flag.StringVar(&dsn, "dsn", "", "Sentry DSN for logging stack traces")
}

func initSentry() {
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
}

func main() {
	initSentry()
	defer sentry.Flush(5 * time.Second)

	err := ctxerror.New("something went wrong").
		SetContext("user", map[string]any{"id": "123", "name": "Alice"}).
		SetTag("module", "example")

	sentry.WithScope(func(scope *sentry.Scope) {
		// if any error context is available, add it to the Sentry scope
		scope.SetContexts(ctxerror.GetContext(err))
		scope.SetTags(ctxerror.GetTags(err))
		scope.SetLevel(sentry.LevelError)
		// Sentry extracts the stack trace from the error automatically
		sentry.CaptureException(err)
	})
}
