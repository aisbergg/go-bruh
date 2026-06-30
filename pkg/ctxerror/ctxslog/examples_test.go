package ctxslog_test

import (
	"context"
	"log/slog"
	"os"

	"github.com/aisbergg/go-bruh/pkg/ctxerror"
	"github.com/aisbergg/go-bruh/pkg/ctxerror/ctxslog"
)

func initLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		}}),
	)
}

func ExampleAsAttributes() {
	logger := initLogger()

	err := ctxerror.New("request failed").
		SetContext("user", map[string]any{"id": "u1"}).
		SetTag("env", "prod")

	// equivalent to logger.Error(...)
	logger.LogAttrs(
		context.Background(),
		slog.LevelError,
		"error occurred",
		ctxslog.AsAttributes(err)...,
	)

	// Output:
	// level=ERROR msg="error occurred" error="request failed" user.id=u1 env=prod
}
