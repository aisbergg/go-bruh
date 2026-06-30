package ctxslog_test

import (
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/aisbergg/go-bruh/internal/testutils"
	"github.com/aisbergg/go-bruh/pkg/ctxerror"
	"github.com/aisbergg/go-bruh/pkg/ctxerror/ctxslog"
)

// logValueError is a test helper that wraps an error and returns a fixed
// slog.Value from LogValue to exercise AsAttributes's LogValuer branch.
type logValueError struct {
	err   error
	value slog.Value
}

func (e logValueError) Error() string        { return e.err.Error() }
func (e logValueError) Unwrap() error        { return e.err }
func (e logValueError) LogValue() slog.Value { return e.value }

// attrsByKey indexes a slice of slog.Attr by key for order-independent checks.
func attrsByKey(attrs []slog.Attr) map[string]slog.Value {
	out := make(map[string]slog.Value, len(attrs))
	for _, a := range attrs {
		out[a.Key] = a.Value
	}
	return out
}

func TestAsAttributes(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("NilErrorReturnsZeroSlogValue", func(t *testing.T) {
		assert.Equal([]slog.Attr{}, ctxslog.AsAttributes(nil))
	})

	t.Run("ErrorMessageIsPresentUnderErrorKey", func(t *testing.T) {
		err := ctxerror.New("boom")
		attrs := attrsByKey(ctxslog.AsAttributes(err))
		assert.Equal("boom", attrs["error"].String())
	})

	t.Run("ContextGroupsAreFlattenedAsGroupKeyAttributes", func(t *testing.T) {
		err := ctxerror.New("x").SetContext("req", map[string]any{"id": "r1", "path": "/v1"})
		attrs := attrsByKey(ctxslog.AsAttributes(err))
		assert.Equal("r1", attrs["req.id"].String())
		assert.Equal("/v1", attrs["req.path"].String())
	})

	t.Run("TagsAppearAsTopLevelStringAttributes", func(t *testing.T) {
		err := ctxerror.New("x").
			SetTag("env", "prod").
			SetTag("op", "write")
		attrs := attrsByKey(ctxslog.AsAttributes(err))
		assert.Equal("prod", attrs["env"].String())
		assert.Equal("write", attrs["op"].String())
	})

	t.Run("AllSlogTypedContextValuesAreConvertedCorrectly", func(t *testing.T) {
		now := time.Unix(1700000000, 0).UTC()
		err := ctxerror.New("x").SetContext(
			"ctx", map[string]any{
				"str":    "s",
				"bool":   true,
				"int":    7,
				"int64":  int64(8),
				"float":  1.25,
				"time":   now,
				"dur":    3 * time.Second,
				"nested": map[string]any{"leaf": "x"},
				"labels": map[string]string{"a": "b"},
				"any":    struct{ N int }{N: 9},
			},
		)
		attrs := attrsByKey(ctxslog.AsAttributes(err))
		assert.Equal("s", attrs["ctx.str"].String())
		assert.True(attrs["ctx.bool"].Bool())
		assert.Equal(int64(7), attrs["ctx.int"].Int64())
		assert.Equal(int64(8), attrs["ctx.int64"].Int64())
		assert.Equal(1.25, attrs["ctx.float"].Float64())
		assert.Equal(now, attrs["ctx.time"].Time())
		assert.Equal(3*time.Second, attrs["ctx.dur"].Duration())
		assert.Equal("x", attrs["ctx.nested.leaf"].String())
		assert.Equal("b", attrs["ctx.labels.a"].String())
		assert.True(reflect.DeepEqual(attrs["ctx.any"].Any(), struct{ N int }{N: 9}), "unexpected Any value")
	})

	t.Run("GroupLogValuerAttrsAreAppended", func(t *testing.T) {
		e := ctxerror.New("x")
		wrapped := logValueError{
			err:   e,
			value: slog.GroupValue(slog.String("extra", "yes")),
		}
		attrs := attrsByKey(ctxslog.AsAttributes(wrapped))
		assert.Equal("yes", attrs["extra"].String())
	})

	t.Run("NonGroupLogValuerValueIsIgnored", func(t *testing.T) {
		e := ctxerror.New("x")
		wrapped := logValueError{
			err:   e,
			value: slog.StringValue("ignored"),
		}
		attrs := attrsByKey(ctxslog.AsAttributes(wrapped))
		_, hasIgnored := attrs["ignored"]
		assert.False(hasIgnored, "non-group LogValuer value must not be appended")
	})

	t.Run("AsAttributes", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		baseErr := ctxerror.New("base error").
			SetContext("user", map[string]any{"id": "123"}).
			SetTag("region", "us-west")
		wrappedErr := ctxerror.Wrap(baseErr, "wrapped error").SetContext("request", map[string]any{"id": "req-abc"})
		slogValue := ctxslog.AsAttributes(wrappedErr)
		attrs := slogValue

		// Expected attributes
		expectedAttrs := map[string]string{
			"error":      "wrapped error: base error",
			"user.id":    "123",
			"region":     "us-west",
			"request.id": "req-abc",
		}

		// Convert actual attributes to a map for easy lookup
		actualAttrs := map[string]string{}
		for _, attr := range attrs {
			actualAttrs[attr.Key] = attr.Value.String()
		}

		// Check for expected attributes
		for key, val := range expectedAttrs {
			got, ok := actualAttrs[key]
			assert.True(ok, "expected attribute %s not found", key)
			assert.Equal(val, got)
		}
	})
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

	attrs := attrsByKey(ctxslog.ContextToAttributes(ctx))

	assert.Equal("s", attrs["grp.str"].String())
	assert.True(attrs["grp.bool"].Bool())
	assert.Equal(int64(7), attrs["grp.int"].Int64())
	assert.Equal(int64(8), attrs["grp.int64"].Int64())
	assert.Equal(1.25, attrs["grp.float"].Float64())
	assert.Equal(now, attrs["grp.time"].Time())
	assert.Equal(3*time.Second, attrs["grp.dur"].Duration())
	assert.Equal("x", attrs["grp.nested.leaf"].String())
	assert.Equal("b", attrs["grp.labels.a"].String())
}

func TestTagsToAttributes(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	tags := ctxerror.Tags{"env": "prod", "op": "write"}
	attrs := attrsByKey(ctxslog.TagsToAttributes(tags))

	assert.Equal("prod", attrs["env"].String())
	assert.Equal("write", attrs["op"].String())
}
