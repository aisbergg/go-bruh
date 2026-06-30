// Package ctxotel provides OpenTelemetry integration and helps converting errors and metadata into OTEL attributes.
package ctxotel

import (
	"fmt"
	"time"

	"github.com/aisbergg/go-bruh/pkg/ctxerror"
	"go.opentelemetry.io/otel/attribute"
)

// AsAttributes converts an error (with ctxerror context/tags) into a slice of
// OTEL attribute key/value pairs. The error message is stored under the
// "error" key. Context groups are flattened using dot notation (group.key).
func AsAttributes(err error) []attribute.KeyValue {
	if err == nil {
		return []attribute.KeyValue{}
	}

	// estimate capacity
	attrsSizeGuess := 1
	ctx := ctxerror.GetContext(err)
	for _, m := range ctx {
		attrsSizeGuess += len(m)
	}
	tags := ctxerror.GetTags(err)
	if tags != nil {
		attrsSizeGuess += len(tags)
	}

	attrs := make([]attribute.KeyValue, 0, attrsSizeGuess)
	attrs = append(attrs, attribute.String("error", err.Error()))

	for group, m := range ctx {
		for k, v := range m {
			attrs = toOTELAttributesRec(v, group+"."+k, attrs)
		}
	}
	for k, v := range tags {
		attrs = append(attrs, attribute.String(k, v))
	}

	return attrs
}

// ContextToAttributes converts a ctxerror.Context into OTEL attributes.
func ContextToAttributes(ctx ctxerror.Context) []attribute.KeyValue {
	if len(ctx) == 0 {
		return []attribute.KeyValue{}
	}
	attrsSizeGuess := 0
	for _, m := range ctx {
		attrsSizeGuess += len(m)
	}
	attrsSizeGuess += 2

	attrs := make([]attribute.KeyValue, 0, attrsSizeGuess)
	for group, m := range ctx {
		for k, v := range m {
			attrs = toOTELAttributesRec(v, group+"."+k, attrs)
		}
	}
	return attrs
}

// TagsToAttributes converts ctxerror.Tags into OTEL attributes.
func TagsToAttributes(tags ctxerror.Tags) []attribute.KeyValue {
	if len(tags) == 0 {
		return []attribute.KeyValue{}
	}
	attrs := make([]attribute.KeyValue, 0, len(tags))
	for k, v := range tags {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}

// toOTELAttributesRec converts arbitrary data into OTEL attribute key/value
// pairs. Nested maps are flattened using dot notation in the key.
func toOTELAttributesRec(data any, key string, attrs []attribute.KeyValue) []attribute.KeyValue {
	switch t := data.(type) {
	case string:
		attrs = append(attrs, attribute.String(key, t))
	case bool:
		attrs = append(attrs, attribute.Bool(key, t))
	case int:
		attrs = append(attrs, attribute.Int(key, t))
	case int64:
		attrs = append(attrs, attribute.Int64(key, t))
	case float64:
		attrs = append(attrs, attribute.Float64(key, t))
	case time.Time:
		attrs = append(attrs, attribute.String(key, t.Format(time.RFC3339Nano)))
	case time.Duration:
		attrs = append(attrs, attribute.String(key, t.String()))
	case map[string]any:
		for k, v := range t {
			attrs = toOTELAttributesRec(v, key+"."+k, attrs)
		}
	case map[string]string:
		for k, v := range t {
			attrs = append(attrs, attribute.String(key+"."+k, v))
		}
	case []string:
		attrs = append(attrs, attribute.StringSlice(key, t))
	case []bool:
		attrs = append(attrs, attribute.BoolSlice(key, t))
	case []int:
		attrs = append(attrs, attribute.IntSlice(key, t))
	case []int64:
		attrs = append(attrs, attribute.Int64Slice(key, t))
	case []float64:
		attrs = append(attrs, attribute.Float64Slice(key, t))
	default:
		attrs = append(attrs, attribute.String(key, fmt.Sprint(data)))
	}
	return attrs
}
