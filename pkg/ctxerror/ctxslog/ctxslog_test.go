package ctxslog_test

import (
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/aisbergg/go-bruh/internal/lib/test/assert"
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

	t.Run("NilErrorReturnsZeroSlogValue", func(t *testing.T) {
		assert.Equal(t, ctxslog.AsAttributes(nil), []slog.Attr{})
	})

	t.Run("ErrorMessageIsPresentUnderErrorKey", func(t *testing.T) {
		err := ctxerror.New("boom")
		attrs := attrsByKey(ctxslog.AsAttributes(err))
		assert.Equal(t, attrs["error"].String(), "boom")
	})

	t.Run("ContextGroupsAreFlattenedAsGroupKeyAttributes", func(t *testing.T) {
		err := ctxerror.New("x").SetContext("req", map[string]any{"id": "r1", "path": "/v1"})
		attrs := attrsByKey(ctxslog.AsAttributes(err))
		assert.Equal(t, attrs["req.id"].String(), "r1")
		assert.Equal(t, attrs["req.path"].String(), "/v1")
	})

	t.Run("TagsAppearAsTopLevelStringAttributes", func(t *testing.T) {
		err := ctxerror.New("x").
			SetTag("env", "prod").
			SetTag("op", "write")
		attrs := attrsByKey(ctxslog.AsAttributes(err))
		assert.Equal(t, attrs["env"].String(), "prod")
		assert.Equal(t, attrs["op"].String(), "write")
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
		assert.Equal(t, attrs["ctx.str"].String(), "s")
		assert.True(t, attrs["ctx.bool"].Bool())
		assert.Equal(t, attrs["ctx.int"].Int64(), int64(7))
		assert.Equal(t, attrs["ctx.int64"].Int64(), int64(8))
		assert.Equal(t, attrs["ctx.float"].Float64(), 1.25)
		assert.Equal(t, attrs["ctx.time"].Time(), now)
		assert.Equal(t, attrs["ctx.dur"].Duration(), 3*time.Second)
		assert.Equal(t, attrs["ctx.nested.leaf"].String(), "x")
		assert.Equal(t, attrs["ctx.labels.a"].String(), "b")
		assert.True(t, reflect.DeepEqual(attrs["ctx.any"].Any(), struct{ N int }{N: 9}), "unexpected Any value")
	})

	t.Run("GroupLogValuerAttrsAreAppended", func(t *testing.T) {
		e := ctxerror.New("x")
		wrapped := logValueError{
			err:   e,
			value: slog.GroupValue(slog.String("extra", "yes")),
		}
		attrs := attrsByKey(ctxslog.AsAttributes(wrapped))
		assert.Equal(t, attrs["extra"].String(), "yes")
	})

	t.Run("NonGroupLogValuerValueIsIgnored", func(t *testing.T) {
		e := ctxerror.New("x")
		wrapped := logValueError{
			err:   e,
			value: slog.StringValue("ignored"),
		}
		attrs := attrsByKey(ctxslog.AsAttributes(wrapped))
		_, hasIgnored := attrs["ignored"]
		assert.False(t, hasIgnored, "non-group LogValuer value must not be appended")
	})

	t.Run("AsAttributes", func(t *testing.T) {
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
			assert.Truef(t, ok, "expected attribute %s not found", key)
			assert.Equal(t, got, val)
		}
	})
}

func TestContextToAttributes(t *testing.T) {
	t.Parallel()

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

	assert.Equal(t, attrs["grp.str"].String(), "s")
	assert.True(t, attrs["grp.bool"].Bool())
	assert.Equal(t, attrs["grp.int"].Int64(), int64(7))
	assert.Equal(t, attrs["grp.int64"].Int64(), int64(8))
	assert.Equal(t, attrs["grp.float"].Float64(), 1.25)
	assert.Equal(t, attrs["grp.time"].Time(), now)
	assert.Equal(t, attrs["grp.dur"].Duration(), 3*time.Second)
	assert.Equal(t, attrs["grp.nested.leaf"].String(), "x")
	assert.Equal(t, attrs["grp.labels.a"].String(), "b")
}

func TestTagsToAttributes(t *testing.T) {
	t.Parallel()

	tags := ctxerror.Tags{"env": "prod", "op": "write"}
	attrs := attrsByKey(ctxslog.TagsToAttributes(tags))

	assert.Equal(t, attrs["env"].String(), "prod")
	assert.Equal(t, attrs["op"].String(), "write")
}
