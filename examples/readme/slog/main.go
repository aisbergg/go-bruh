package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aisbergg/go-bruh/pkg/ctxerror"
	"github.com/aisbergg/go-bruh/pkg/ctxerror/ctxslog"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

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
}
