# Changelog

All notable changes to this project will be documented in this file.

- [1.3.0 (2026-06-26)](#130-2026-06-26)
- [1.2.0 (2025-08-15)](#120-2025-08-15)
- [1.1.0 (2024-07-14)](#110-2024-07-14)
- [1.0.2 (2023-05-28)](#102-2023-05-28)
- [1.0.1 (2023-05-27)](#101-2023-05-27)
- [1.0.0 (2023-05-27)](#100-2023-05-27)

---

<a name="1.3.0"></a>

## [1.3.0](https://github.com/aisbergg/go-bruh/compare/v1.2.0...v1.3.0) (2026-06-26)

### Features

- Configurable stack depth of individual errors (`MaxErrorStackDepth`) and error chains (`MaxChainStackDepth`).
- Added message helper `bruh.MessageLastN` to return the last N messages of a chain.

### Performance

- Reduce allocations by introducing object pools
- `Error()` function call now lazily creates the error message and caches the result for later invocations.

### Documentation

- Extended the documentation.
- Examples: new example showing custom JSON formatter and a `TimestampedError` custom error; OTEL example demonstrates `ctxerror/ctxotel` integration and exporting `bruh` traces as span attributes.
- Examples updated: custom formatter (JSON), custom error with timestamps, OTEL integration sample, and usage of the new fancy/sourced formatters.
    - `ctxerror/ctxotel` example shows how to convert error context/tags to OpenTelemetry attributes.

### ⚠ BREAKING CHANGES

- `ctxerror` overhauled and improved metadata model to be directly compatible with Sentry
    - convenience setters: `SetContext`, `SetContexts`, `SetTag`, `SetTags`.
    - chain-shared metadata by default, plus `Unshare()` to make private copies.
- `multierror` overhauled and improved model.

<a name="1.2.0"></a>

## [1.2.0](https://github.com/aisbergg/go-bruh/compare/v1.1.0...v1.2.0) (2025-08-15)

### Code Refactoring

- rewrite entire package

### Documentation

- clarifies error message storytelling

### ⚠ BREAKING CHANGES

- Renamed `bruh.TraceableError` to `bruh.Err`
- Renamed `bruh.ToString` to `bruh.String` and gave it a new signature: `func String(err error) string`
- Renamed `bruh.ToCustomString` to `bruh.StringFormat` and gave it a new signature: `func StringFormat(err error, f Formatter, unpackAll ...bool) string`
- New signature for `bruh.Formatter`: `type Formatter func(b []byte, unpacker *Unpacker) []byte`
- Removed `*bruh.TraceableError.FullStack()`, now it is just `*bruh.Err.Stack()`
- Moved package `github.com/aisbergg/go-bruh/pkg/bruh/ctxerror` to `github.com/aisbergg/go-bruh/pkg/ctxerror`
- Renamed `ctxerror.ContextableError` to `ctxerror.Err`

<a name="1.1.0"></a>

## [1.1.0](https://github.com/aisbergg/go-bruh/compare/v1.0.2...v1.1.0) (2024-07-14)

### Bug Fixes

- properly identify and exclude globally defined errors again

### Features

- add `RangeContext` to ContextableError and improve general performance by using a slice as a backing for the context data
- add ProgramCounter2 that indicates the actual instruction being executed and not the one beforehand

<a name="1.0.2"></a>

## [1.0.2](https://github.com/aisbergg/go-bruh/compare/v1.0.1...v1.0.2) (2023-05-28)

### Performance Improvements

- improve performance of FormatWithCombinedTrace, FormatPythonTraceback and Stack.String()

<a name="1.0.1"></a>

## [1.0.1](https://github.com/aisbergg/go-bruh/compare/v1.0.0...v1.0.1) (2023-05-27)

### Documentation

- improve examples

<a name="1.0.0"></a>

## [1.0.0]() (2023-05-27)

Initial Release
