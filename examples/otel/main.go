package main

import (
	"context"
	"flag"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/aisbergg/go-bruh/pkg/bruh"
	"github.com/aisbergg/go-bruh/pkg/ctxerror"
	"github.com/aisbergg/go-bruh/pkg/ctxerror/ctxotel"
)

func recordError(span trace.Span, err error) {
	stackTrace := bruh.StringFormat(err, bruh.BruhStackedFancyFormatter(false, false, true))
	span.RecordError(
		err,
		trace.WithAttributes(attribute.String("exception.stacktrace", stackTrace)),
		trace.WithAttributes(ctxotel.AsAttributes(err)...),
	)
	span.SetStatus(codes.Error, err.Error())
}

func main() {
	flag.Parse()

	var exporter sdktrace.SpanExporter
	var err error
	if httpExport {
		exporter, err = otelHTTPTraceExporter(context.Background())
	} else {
		exporter, err = stdoutTraceExporter()
	}
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	tp, err := initTracer(exporter)
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	// ensure tracer provider is shut down so exporter flushes
	ctx := context.Background()
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	defer func() {
		_ = tp.Shutdown(shutdownCtx)
	}()

	tracer := otel.Tracer("go-bruh-tracer")
	if err := doWork(ctx, tracer); err != nil {
		log.Printf("doWork returned error: %v", err)
	}
}

func doWork(ctx context.Context, tracer trace.Tracer) error {
	ctx, span := tracer.Start(ctx, "doWork")
	defer span.End()

	err := subFunction(ctx, tracer)
	if err != nil {
		err = bruh.Wrap(err, "perform work")
		recordError(span, err)
		return err
	}
	return nil
}

func subFunction(ctx context.Context, tracer trace.Tracer) error {
	ctx, span := tracer.Start(ctx, "subFunction")
	defer span.End()

	// simulate work
	time.Sleep(50 * time.Millisecond)

	span.SetAttributes(attribute.String("subFunction.attribute", "value"))

	return ctxerror.New("root cause").
		SetContext("req", map[string]any{
			"path": "/v1",
		}).
		SetContext("user", map[string]any{
			"id":   123,
			"name": "Alice",
		}).
		SetTag("kind", "example")
}

func initTracer(exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	res := resource.NewSchemaless(attribute.String("service.name", "bruh"))
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func stdoutTraceExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

func otelHTTPTraceExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	client := otlptracehttp.NewClient(otlptracehttp.WithEndpointURL("http://localhost:4318/v1/traces"))
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

var httpExport bool

func init() {
	flag.BoolVar(&httpExport, "http-export", false, "Use the OTLP HTTP exporter instead of stdout")
}
