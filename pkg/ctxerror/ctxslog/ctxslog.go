// Package ctxslog provides slog integration and helps converting errors and metadata into slog attributes.
package ctxslog

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unsafe"

	"github.com/aisbergg/go-bruh/pkg/bruh"
	"github.com/aisbergg/go-bruh/pkg/ctxerror"
)

// AsAttributes turns the given error into a slice of slog.Attr, which then can
// be used with slog's LogAttrs method.
//
// The error message is included under the "error" key, and context and tags are
// included as additional attributes. If the error implements slog.LogValuer,
// its LogValue is also included.
func AsAttributes(err error) []slog.Attr {
	if err == nil {
		return []slog.Attr{}
	}

	attrsSizeGuess := 5 // error message and maybe some from LogValuers
	ctx := ctxerror.GetContext(err)
	for _, v := range ctx {
		attrsSizeGuess += len(v)
	}
	tags := ctxerror.GetTags(err)
	if tags != nil {
		attrsSizeGuess += len(tags)
	}

	attrs := make([]slog.Attr, 0, attrsSizeGuess)
	keyBuilder := &keyBuilder{}
	keyBuilder.InitialSize(ctx)
	attrs = append(attrs, slog.String("error", err.Error()))
	for k, v := range ctx {
		attrs = convertContextMapToAttributes(keyBuilder, k, v, attrs)
	}
	for k, v := range tags {
		attrs = append(attrs, slog.String(k, v))
	}
	for ; err != nil; err = bruh.Unwrap(err) {
		if e, ok := err.(slog.LogValuer); ok {
			value := e.LogValue()
			if value.Kind() == slog.KindGroup {
				attrs = append(attrs, value.Group()...)
			}
		}
	}

	return attrs
}

// ContextToAttributes converts a Sentry-style context
// (map[string]map[string]any) into a slice of slog.Attr. Nested maps are
// flattened with dot notation (e.g. "group.key"). Unsupported types are
// converted to strings using fmt.Sprint.
func ContextToAttributes(ctx ctxerror.Context) []slog.Attr {
	if len(ctx) == 0 {
		return []slog.Attr{}
	}
	// count total number of attributes to preallocate the slice with a good capacity
	attrsSizeGuess := 0
	for _, m := range ctx {
		attrsSizeGuess += len(m)
	}
	attrsSizeGuess += 4 // extra space

	keyBuilder := &keyBuilder{}
	keyBuilder.InitialSize(ctx)
	attrs := make([]slog.Attr, 0, attrsSizeGuess)
	for group, m := range ctx {
		attrs = convertContextMapToAttributes(keyBuilder, group, m, attrs)
	}
	return attrs
}

// TagsToAttributes converts a Sentry-style tags (map[string]string) into a slice
// of slog.Attr. Unsupported types are converted to strings using fmt.Sprint.
func TagsToAttributes(tags ctxerror.Tags) []slog.Attr {
	if len(tags) == 0 {
		return []slog.Attr{}
	}
	// count total number of attributes to preallocate the slice with a good capacity
	attrs := make([]slog.Attr, 0, len(tags))
	for k, v := range tags {
		attrs = append(attrs, slog.String(k, v))
	}
	return attrs
}

func convertContextMapToAttributes(
	keyBuilder *keyBuilder,
	group string,
	value map[string]any,
	attrs []slog.Attr,
) []slog.Attr {
	for k, v := range value {
		k = keyBuilder.Build(group, k)
		switch t := v.(type) {
		case string:
			attrs = append(attrs, slog.String(k, t))
		case bool:
			attrs = append(attrs, slog.Bool(k, t))
		case int:
			attrs = append(attrs, slog.Int(k, t))
		case int64:
			attrs = append(attrs, slog.Int64(k, t))
		case float64:
			attrs = append(attrs, slog.Float64(k, t))
		case time.Time:
			attrs = append(attrs, slog.Time(k, t))
		case time.Duration:
			attrs = append(attrs, slog.Duration(k, t))
		case map[string]any:
			attrs = convertContextMapToAttributes(keyBuilder, k, t, attrs)
		case map[string]string:
			for mk, mv := range t {
				subKey := keyBuilder.Build(k, mk)
				attrs = append(attrs, slog.String(subKey, mv))
			}
		case []string:
			attrs = append(attrs, slog.String(k, strings.Join(t, ",")))
		case []bool:
			attrs = append(attrs, slog.String(k, sliceAsCommaSeparatedString(t)))
		case []int:
			attrs = append(attrs, slog.String(k, sliceAsCommaSeparatedString(t)))
		case []int64:
			attrs = append(attrs, slog.String(k, sliceAsCommaSeparatedString(t)))
		case []float64:
			attrs = append(attrs, slog.String(k, sliceAsCommaSeparatedString(t)))
		default:
			attrs = append(attrs, slog.Any(k, t))
		}
	}
	return attrs
}

func sliceAsCommaSeparatedString[T any](slice []T) string {
	strs := make([]string, len(slice))
	for i, v := range slice {
		strs[i] = fmt.Sprint(v)
	}
	return strings.Join(strs, ",")
}

type keyBuilder struct {
	buf []byte
}

func (b *keyBuilder) Build(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}
	requiredLen := len(parts[0])
	for _, part := range parts[1:] {
		requiredLen += len(part) + 1
	}
	curBuf := b.buf[len(b.buf):len(b.buf)]
	if cap(b.buf)-len(b.buf) < requiredLen {
		if cap(b.buf) < requiredLen {
			curBuf = make([]byte, 0, requiredLen)
			b.buf = curBuf
		} else {
			// allocate a new slice
			curBuf = make([]byte, 0, cap(b.buf))
			b.buf = curBuf
		}
	}

	curBuf = append(curBuf, parts[0]...)
	for _, part := range parts[1:] {
		curBuf = append(curBuf, '.')
		curBuf = append(curBuf, part...)
	}
	b.buf = b.buf[:len(b.buf)+len(curBuf)]

	return unsafe.String(unsafe.SliceData(curBuf), len(curBuf))
}

func (b *keyBuilder) InitialSize(ctx ctxerror.Context) {
	guess := 0
	for group, m := range ctx {
		for k := range m {
			guess += len(group) + len(k) + 1 // key name + dot or end
		}
	}
	// 15% extra space for tags and LogValuer attributes
	guess += guess * 15 / 100
	b.buf = make([]byte, 0, guess)
}
