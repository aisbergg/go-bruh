package ctxotel_test

import (
	"fmt"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/aisbergg/go-bruh/internal/testutils"
	"github.com/aisbergg/go-bruh/pkg/ctxerror"
	"github.com/aisbergg/go-bruh/pkg/ctxerror/ctxotel"
)

func attrsByKey(attrs []attribute.KeyValue) map[string]attribute.Value {
	out := make(map[string]attribute.Value, len(attrs))
	for _, a := range attrs {
		out[string(a.Key)] = a.Value
	}
	return out
}

func TestContextToAttributes(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	now := time.Unix(1700000000, 0).UTC()
	ctx := ctxerror.Context{
		"grp": {
			"str":    "s",
			"bool":   true,
			"int":    7,
			"int64":  int64(8),
			"float":  1.25,
			"time":   now,
			"dur":    3 * time.Second,
			"nested": map[string]any{"leaf": "x"},
			"labels": map[string]string{"a": "b"},
		},
	}

	attrs := attrsByKey(ctxotel.ContextToAttributes(ctx))

	assert.Equal("s", fmt.Sprint(attrs["grp.str"].AsInterface()))
	assert.True(attrs["grp.bool"].AsInterface().(bool))
	// numeric attrs: compare stringified values to avoid type differences
	assert.Equal("7", fmt.Sprint(attrs["grp.int"].AsInterface()))
	assert.Equal("8", fmt.Sprint(attrs["grp.int64"].AsInterface()))
	assert.Equal("1.25", fmt.Sprint(attrs["grp.float"].AsInterface()))

	// time stored as RFC3339Nano string
	gotTimeStr := fmt.Sprint(attrs["grp.time"].AsInterface())
	parsed, err := time.Parse(time.RFC3339Nano, gotTimeStr)
	assert.NoError(err)
	assert.Equal(now, parsed)

	// duration stored as string
	assert.Equal((3 * time.Second).String(), fmt.Sprint(attrs["grp.dur"].AsInterface()))

	assert.Equal("x", fmt.Sprint(attrs["grp.nested.leaf"].AsInterface()))
	assert.Equal("b", fmt.Sprint(attrs["grp.labels.a"].AsInterface()))
}

func TestTagsToAttributes(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	tags := ctxerror.Tags{"env": "prod", "op": "write"}
	attrs := attrsByKey(ctxotel.TagsToAttributes(tags))

	assert.Equal("prod", fmt.Sprint(attrs["env"].AsInterface()))
	assert.Equal("write", fmt.Sprint(attrs["op"].AsInterface()))
}

func TestAsAttributes(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	baseErr := ctxerror.New("base error").SetContext("user", map[string]any{"id": "123"}).SetTag("region", "us-west")
	wrappedErr := ctxerror.Wrap(baseErr, "wrapped error").SetContext("request", map[string]any{"id": "req-abc"})

	attrs := attrsByKey(ctxotel.AsAttributes(wrappedErr))

	expected := map[string]string{
		"error":      "wrapped error: base error",
		"user.id":    "123",
		"region":     "us-west",
		"request.id": "req-abc",
	}

	for k, v := range expected {
		got := fmt.Sprint(attrs[k].AsInterface())
		assert.Equal(v, got)
	}
}
