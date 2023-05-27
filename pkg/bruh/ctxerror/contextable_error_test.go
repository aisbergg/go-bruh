package ctxerror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aisbergg/go-bruh/internal/testutils"
	"github.com/aisbergg/go-bruh/pkg/bruh"
)

func TestContextableError(t *testing.T) {
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
	assert.Equal(err.Error(), msg, "New returned error with message %q, expected %q", err.Error(), msg)
	ctx := GetContext(err)
	assert.Equal(len(ctx), 2, "New returned error with context %+v, expected %+v", ctx, map[string]interface{}{key1: value1, key2: value2})

	// Test NewSkip function
	err = NewSkip(skip, msg, key1, value1, key2, value2)
	require.NotNil(err, "NewSkip returned nil error")
	assert.Equal(err.Error(), msg, "NewSkip returned error with message %q, expected %q", err.Error(), msg)
	ctx = GetContext(err)
	assert.Equal(len(ctx), 2, "NewSkip returned error with context %+v, expected %+v", ctx, map[string]interface{}{key1: value1, key2: value2})

	// Test Wrap function
	wrappedErr := New("root error")
	err = Wrap(wrappedErr, wrapErrMsg, key1, value1, key2, value2)
	require.NotNil(err, "Wrap returned nil error")
	assert.Equal(err.Error(), fmt.Sprintf("%s: %s", wrapErrMsg, wrappedErr.Error()), "Wrap returned error with message %q, expected %q", err.Error(), wrapErrMsg)
	ctx = GetContext(err)
	assert.Equal(len(ctx), 2, "Wrap returned error with context %+v, expected %+v", ctx, map[string]interface{}{key1: value1, key2: value2})

	// Test WrapSkip function
	err = WrapSkip(wrappedErr, skip, wrapErrMsg, key1, value1, key2, value2)
	require.NotNil(err, "WrapSkip returned nil error")
	assert.Equal(err.Error(), fmt.Sprintf("%s: %s", wrapErrMsg, wrappedErr.Error()), "WrapSkip returned error with message %q, expected %q", err.Error(), wrapErrMsg)
	ctx = GetContext(err)
	assert.Equal(len(ctx), 2, "WrapSkip returned error with context %+v, expected %+v", ctx, map[string]interface{}{key1: value1, key2: value2})
}

func TestContextableError_Modifiers(t *testing.T) {
	const (
		msg    = "test error"
		key1   = "key1"
		value1 = "value1"
		key2   = "key2"
		value2 = "value2"
	)

	assert := testutils.NewAssert(t)
	require := testutils.NewRequire(t)

	// Test Add
	err := New(msg).(ContextableErrorer).Add(key1, value1)
	require.NotNil(err, "New returned nil error")
	assert.Equal(err.Context(), map[string]any{key1: value1}, "Add did not add the expected context")

	// Test AddAll
	err = New(msg).(ContextableErrorer).AddAll(map[string]any{key1: value1, key2: value2})
	require.NotNil(err, "New returned nil error")
	assert.Equal(err.Context(), map[string]any{key1: value1, key2: value2}, "AddAll did not add the expected context")

	// Test Remove
	err = New(msg).(ContextableErrorer).AddAll(map[string]any{key1: value1, key2: value2}).Remove(key1)
	require.NotNil(err, "New returned nil error")
	assert.Equal(err.Context(), map[string]any{key2: value2}, "Remove did not remove the expected context")
}

func TestGetContext(t *testing.T) {
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
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			assert := testutils.NewAssert(t)
			assert.Equal(tc.exp, GetContext(tc.err))
		})
	}
}

func TestGetFullContext(t *testing.T) {
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
			assert.Equal(tc.exp, GetFullContext(tc.err))
		})
	}
}
