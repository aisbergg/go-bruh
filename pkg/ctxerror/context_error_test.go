package ctxerror

import (
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/aisbergg/go-bruh/internal/testutils"
	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func TestContextableError(t *testing.T) {
	t.Parallel()
	const (
		msg        = "test error"
		key1       = "key1"
		value1     = "value1"
		key2       = "key2"
		value2     = "value2"
		skip       = 1
		wrapErrMsg = "wrapped error"
	)
	assert := testutils.NewAssert(t)
	require := testutils.NewRequire(t)

	// Test New function
	err := New(msg, key1, value1, key2, value2)
	require.NotNil(err, "New returned nil error")
	assert.Equal(
		err.Error(),
		msg,
		"New returned error with message %q, expected %q",
		err.Error(),
		msg,
	)
	ctx := GetContext(err)
	assert.Equal(
		len(ctx),
		2,
		"New returned error with context %+v, expected %+v",
		ctx,
		map[string]any{key1: value1, key2: value2},
	)

	// Test NewSkip function
	err = NewSkip(skip, msg, key1, value1, key2, value2)
	require.NotNil(err, "NewSkip returned nil error")
	assert.Equal(
		err.Error(),
		msg,
		"NewSkip returned error with message %q, expected %q",
		err.Error(),
		msg,
	)
	ctx = GetContext(err)
	assert.Equal(
		len(ctx),
		2,
		"NewSkip returned error with context %+v, expected %+v",
		ctx,
		map[string]any{key1: value1, key2: value2},
	)

	// Test Wrap function
	wrappedErr := New("root error")
	err = Wrap(wrappedErr, wrapErrMsg, key1, value1, key2, value2)
	require.NotNil(err, "Wrap returned nil error")
	assert.Equal(
		err.Error(),
		fmt.Sprintf("%s: %s", wrapErrMsg, wrappedErr.Error()),
		"Wrap returned error with message %q, expected %q",
		err.Error(),
		wrapErrMsg,
	)
	ctx = GetContext(err)
	assert.Equal(
		len(ctx),
		2,
		"Wrap returned error with context %+v, expected %+v",
		ctx,
		map[string]any{key1: value1, key2: value2},
	)

	// Test WrapSkip function
	err = WrapSkip(wrappedErr, skip, wrapErrMsg, key1, value1, key2, value2)
	require.NotNil(err, "WrapSkip returned nil error")
	assert.Equal(
		err.Error(),
		fmt.Sprintf("%s: %s", wrapErrMsg, wrappedErr.Error()),
		"WrapSkip returned error with message %q, expected %q",
		err.Error(),
		wrapErrMsg,
	)
	ctx = GetContext(err)
	assert.Equal(
		len(ctx),
		2,
		"WrapSkip returned error with context %+v, expected %+v",
		ctx,
		map[string]any{key1: value1, key2: value2},
	)
}

func TestModifiers(t *testing.T) {
	t.Parallel()
	const (
		msg    = "test error"
		key1   = "key1"
		value1 = "value1"
		key2   = "key2"
		value2 = "value2"
		key3   = 123
		value3 = 123
	)

	assert := testutils.NewAssert(t)
	require := testutils.NewRequire(t)

	// Test Add
	err := New(msg).Add(key1, value1)
	require.NotNil(err, "New returned nil error")
	assert.Equal(
		err.Context(),
		map[string]any{key1: value1},
		"Add did not add the expected context",
	)
	err = err.Add(key2, value2, key1, value2)
	require.NotNil(err, "New returned nil error")
	assert.Equal(
		err.Context(),
		map[string]any{key1: value2, key2: value2},
		"Add did not overwrite the expected context",
	)
	err = err.Add(key3, value3)
	require.NotNil(err, "New returned nil error")
	assert.Equal(
		err.Context(),
		map[string]any{key1: value2, key2: value2, "123": value3},
		"Add did not add the expected context",
	)

	// Test AddAll
	err = New(msg).AddAll(map[string]any{key1: value1, key2: value2})
	require.NotNil(err, "New returned nil error")
	assert.Equal(
		err.Context(),
		map[string]any{key1: value1, key2: value2},
		"AddAll did not add the expected context",
	)

	// Test Remove
	err = New(msg).AddAll(map[string]any{key1: value1, key2: value2}).Remove(key1)
	require.NotNil(err, "New returned nil error")
	assert.Equal(
		err.Context(),
		map[string]any{key2: value2},
		"Remove did not remove the expected context",
	)
}

func TestGetContext(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		err error
		exp map[string]any
	}{
		"nil error": {
			err: nil,
			exp: map[string]any{},
		},
		"external error": {
			err: errors.New("test error"),
			exp: map[string]any{},
		},
		"empty context": {
			err: New("test error"),
			exp: map[string]any{},
		},
		"simple key-value entry": {
			err: New("test error", "key", "value"),
			exp: map[string]any{"key": "value"},
		},
		"multiple key-value entries": {
			err: New("test error", "key1", "value1", "key2", "value2"),
			exp: map[string]any{"key1": "value1", "key2": "value2"},
		},
		"overwriting key-value entry": {
			err: New("test error", "key", "value1", "key", "value2"),
			exp: map[string]any{"key": "value2"},
		},
		"wrapped error": {
			err: Wrap(
				New("root error", "key1", "value1"),
				"wrapped error", "key2", "value2",
			),
			exp: map[string]any{"key1": "value1", "key2": "value2"},
		},
		"different wrapped error": {
			err: bruh.Wrap(
				New("root error", "key1", "value1", "key2", "value2"),
				"wrapped error",
			),
			exp: map[string]any{"key1": "value1", "key2": "value2"},
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			assert.Equal(tc.exp, GetContext(tc.err))
		})
	}
}

func TestRangeContext(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		err error
		exp map[string]any
	}{
		"nil error": {
			err: nil,
			exp: map[string]any{},
		},
		"external error": {
			err: errors.New("test error"),
			exp: map[string]any{},
		},
		"empty context": {
			err: New("test error"),
			exp: map[string]any{},
		},
		"simple key-value entry": {
			err: New("test error", "key", "value"),
			exp: map[string]any{"key": "value"},
		},
		"multiple key-value entries": {
			err: New("test error", "key1", "value1", "key2", "value2"),
			exp: map[string]any{"key1": "value1", "key2": "value2"},
		},
		"overwriting key-value entry": {
			err: New("test error", "key", "value1", "key", "value2"),
			exp: map[string]any{"key": "value2"},
		},
		"wrapped error": {
			err: Wrap(
				New("root error", "key1", "value1"),
				"wrapped error", "key2", "value2",
			),
			exp: map[string]any{"key1": "value1", "key2": "value2"},
		},
		"different wrapped error": {
			err: bruh.Wrap(
				New("root error", "key1", "value1", "key2", "value2"),
				"wrapped error",
			),
			exp: map[string]any{"key1": "value1", "key2": "value2"},
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			ctx := make(map[string]any)
			RangeContext(tc.err, func(key string, value any) bool {
				ctx[key] = value
				return true
			})
			assert.Equal(tc.exp, ctx)
		})
	}
}

func TestAsSLogValue(t *testing.T) {
	t.Parallel()
	const (
		msg    = "test error"
		key1   = "key1"
		value1 = "value1"
		key2   = "key2"
		value2 = "value2"
		key3   = 123
		value3 = 123
	)
	assert := testutils.NewAssert(t)

	type testCase struct {
		name     string
		err      error
		expected slog.Value
	}

	tests := []testCase{
		{
			name:     "nil error",
			err:      nil,
			expected: slog.Value{},
		},
		{
			name:     "empty context",
			err:      New(msg),
			expected: slog.GroupValue(slog.String("message", msg)),
		},
		{
			name: "single key-value",
			err:  New(msg, key1, value1),
			expected: slog.GroupValue(
				slog.String("message", msg),
				slog.Any(key1, value1),
			),
		},
		{
			name: "multiple key-values",
			err:  New(msg, key1, value1, key2, value2),
			expected: slog.GroupValue(
				slog.String("message", msg),
				slog.Any(key1, value1),
				slog.Any(key2, value2),
			),
		},
		{
			name: "non-string key",
			err:  New(msg, key3, value3),
			expected: slog.GroupValue(
				slog.String("message", msg),
				slog.Any(fmt.Sprint(key3), value3),
			),
		},
		{
			name: "wrapped error with context",
			err:  Wrap(New("root", key1, value1), msg, key2, value2),
			expected: slog.GroupValue(
				slog.String("message", fmt.Sprintf("%s: %s", msg, "root")),
				slog.Any(key2, value2),
				slog.Any(key1, value1),
			),
		},
		{
			name: "external error",
			err:  fmt.Errorf("plain error"),
			expected: slog.GroupValue(
				slog.String("message", "plain error"),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value := AsSLogValue(tc.err)
			// slog.Value does not implement equality, so compare group attrs
			if tc.expected.Kind() == slog.KindGroup {
				expAttrs := tc.expected.Group()
				gotAttrs := value.Group()
				assert.Equal(len(expAttrs), len(gotAttrs), "number of slog attrs mismatch")
				expMap := make(map[string]any, len(expAttrs))
				gotMap := make(map[string]any, len(gotAttrs))
				for _, attr := range expAttrs {
					expMap[attr.Key] = attr.Value.Any()
				}
				for _, attr := range gotAttrs {
					gotMap[attr.Key] = attr.Value.Any()
				}
				assert.Equal(expMap, gotMap, "slog group values mismatch")
			} else {
				assert.Equal(value, tc.expected, "slog values mismatch")
			}
		})
	}
}
