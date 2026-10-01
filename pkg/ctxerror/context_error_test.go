package ctxerror

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aisbergg/go-bruh/internal/lib/test/assert"
	"github.com/aisbergg/go-bruh/internal/lib/test/require"
	"github.com/aisbergg/go-bruh/pkg/bruh"
)

// mustErr unpacks a ModifiableContextErr to *Err, failing the test if that is
// not possible.
func mustErr(t *testing.T, err ModifiableContextErr) ModifiableContextErr {
	t.Helper()
	if err == nil {
		t.Fatal("expected error to be non-nil")
	}
	e, ok := err.(*Err)
	if !ok {
		t.Fatalf("expected *Err, got %T", err)
	}
	return e
}

func isSameObject(x, y any) bool {
	return reflect.ValueOf(x).Pointer() == reflect.ValueOf(y).Pointer()
}

// -----------------------------------------------------------------------------
// Constructors and wrappers
// -----------------------------------------------------------------------------

func TestConstructors(t *testing.T) {
	t.Parallel()

	root := errors.New("root")

	assertConstructor := func(name string, build func() error, expMsg string) {
		t.Run(name, func(t *testing.T) {
			got := build()
			require.NotNil(t, got, "expected non-nil")
			require.Equal(t, got.Error(), expMsg, "unexpected error message")
		})
	}

	assertNilConstructor := func(name string, build func() error) {
		t.Run(name, func(t *testing.T) {
			got := build()
			require.Nil(t, got, "expected nil")
		})
	}

	assertConstructor("New", func() error { return New("x") }, "x")
	assertConstructor("Errorf", func() error { return Errorf("x=%d", 1) }, "x=1")
	assertConstructor("Wrap", func() error { return Wrap(root, "outer") }, "outer: root")
	assertConstructor("Wrapf", func() error { return Wrapf(root, "outer=%d", 7) }, "outer=7: root")
	assertNilConstructor("WrapNilReturnsNil", func() error { return Wrap(nil, "x") })
	assertNilConstructor("WrapfNilReturnsNil", func() error { return Wrapf(nil, "x=%d", 1) })
}

// -----------------------------------------------------------------------------
// Modifier methods – nil receiver safety
// -----------------------------------------------------------------------------

func TestNilReceiverModifiers(t *testing.T) {
	t.Parallel()

	var e *Err
	assert.Nil(t, e.SetContext("group", map[string]any{"k": "v"}))
	assert.Nil(t, e.SetContexts(Context{"group": {"k": "v"}}))
	assert.Nil(t, e.SetTag("k", "v"))
	assert.Nil(t, e.SetTags(Tags{"k": "v"}))
	assert.Nil(t, e.Unshare())
}

// -----------------------------------------------------------------------------
// SetContext / SetContexts
// -----------------------------------------------------------------------------

func TestContextModifiers(t *testing.T) {
	t.Parallel()

	t.Run("SetContextCreatesGroupWhenMissing", func(t *testing.T) {
		e := mustErr(t, New("x"))
		e.SetContext("req", map[string]any{"id": "1"})
		assert.Equal(t, GetContext(e), Context{"req": {"id": "1"}})
	})

	t.Run("SetContextMergesIntoExistingGroupAndOverwritesDuplicateKeys", func(t *testing.T) {
		e := mustErr(t, New("x"))
		e.SetContext("req", map[string]any{"id": "a", "retry": false})
		e.SetContext("req", map[string]any{"id": "b", "path": "/x"})
		assert.Equal(t, GetContext(e), Context{"req": {"id": "b", "retry": false, "path": "/x"}})
	})

	t.Run("SetContextsCreatesMissingGroupsAndMergesExistingOnes", func(t *testing.T) {
		e := mustErr(t, New("x"))
		e.SetContext("req", map[string]any{"id": "1"})
		e.SetContexts(Context{
			"req":  {"method": "GET"},
			"user": {"id": "u1"},
		})
		assert.Equal(t, GetContext(e), Context{
			"req":  {"id": "1", "method": "GET"},
			"user": {"id": "u1"},
		})
	})
}

// -----------------------------------------------------------------------------
// AddTag / AddTags
// -----------------------------------------------------------------------------

func TestTagModifiers(t *testing.T) {
	t.Parallel()

	t.Run("AddTagInsertsAndOverwrites", func(t *testing.T) {
		e := mustErr(t, New("x"))
		e.SetTag("k", "v1")
		e.SetTag("k", "v2")
		assert.Equal(t, GetTags(e), Tags{"k": "v2"})
	})

	t.Run("AddTagsMergesIntoExistingTags", func(t *testing.T) {
		e := mustErr(t, New("x"))
		e.SetTag("a", "1")
		e.SetTags(Tags{"b": "2", "a": "overwritten"})
		assert.Equal(t, GetTags(e), Tags{"a": "overwritten", "b": "2"})
	})
}

// -----------------------------------------------------------------------------
// Owned vs shared metadata across the chain
// -----------------------------------------------------------------------------

func TestChainOwnsContextAndTagsByDefault(t *testing.T) {
	t.Parallel()

	inner := mustErr(t, New("inner")).
		SetContext("req", map[string]any{"id": "1"}).
		SetTag("id", "1")

	outer := mustErr(t, Wrap(inner, "outer")).
		SetContext("req", map[string]any{"path": "/v1"}).
		SetTag("path", "/v1")

	innerCtx := GetContext(inner)
	outerCtx := GetContext(outer)
	// Default is shared: outer mutates same context object visible to inner.
	assert.True(t, isSameObject(innerCtx, outerCtx), "expected same context object by default")
	assert.Equal(t, innerCtx, Context{"req": {"id": "1", "path": "/v1"}})
	assert.Equal(t, outerCtx, Context{"req": {"id": "1", "path": "/v1"}})

	innerTags := GetTags(inner)
	outerTags := GetTags(outer)
	assert.True(t, isSameObject(innerTags, outerTags), "expected same tags object by default")
	assert.Equal(t, innerTags, Tags{"id": "1", "path": "/v1"})
	assert.Equal(t, outerTags, Tags{"id": "1", "path": "/v1"})
}

func TestUnshare(t *testing.T) {
	t.Parallel()

	base := mustErr(t, New("base")).
		SetContext("req", map[string]any{"id": "1"}).
		SetTag("id", "1")

	// mark wrappedUnshared to become unshared; copy happens lazily on first access
	wrappedUnshared := mustErr(t, Wrap(base, "inner")).
		SetContext("req", map[string]any{"path": "/v1"}).
		SetTag("kind", "foo").
		Unshare()
	wrappedOuter := mustErr(t, Wrap(wrappedUnshared, "outer")).
		SetContext("req", map[string]any{"path": "/v2"}).
		SetTag("kind", "foo")

	wrappedInnterCtx := GetContext(wrappedUnshared)
	wrappedInnerTags := GetTags(wrappedUnshared)
	wrappedOuterCtx := GetContext(wrappedOuter)
	wrappedOuterTags := GetTags(wrappedOuter)

	// wrappedUnshared should have its own context map, while wrappedOuter should still share with base
	require.True(
		t,
		isSameObject(wrappedUnshared.(*Err).context, base.(*Err).context),
		"expected wrappedUnshared to have same context object as base before unshare takes effect",
	)
	require.False(
		t,
		isSameObject(wrappedOuter.(*Err).context, base.(*Err).context),
		"expected wrappedOuter to have different context object than base due to unshare",
	)
	require.True(
		t,
		isSameObject(wrappedUnshared.(*Err).tags, base.(*Err).tags),
		"expected wrappedUnshared to have same tags object as base before unshare takes effect",
	)
	require.False(
		t,
		isSameObject(wrappedOuter.(*Err).tags, base.(*Err).tags),
		"expected wrappedOuter to have different tags object than base due to unshare",
	)

	require.Equal(t, wrappedInnterCtx, Context{"req": {"id": "1", "path": "/v1"}})
	require.Equal(t, wrappedInnerTags, Tags{"id": "1", "kind": "foo"})

	// wrappedOuter should reflect latest mutations to base since it shares metadata
	require.Equal(t, wrappedOuterCtx, Context{"req": {"id": "1", "path": "/v2"}})
	require.Equal(t, wrappedOuterTags, Tags{"id": "1", "kind": "foo"})

	// mutation of unshared should not affect outer
	wrappedUnshared.SetContext("req", map[string]any{"path": "/v3"})
	wrappedUnshared.SetTag("kind", "bar")
	wrappedInnterCtx = GetContext(wrappedUnshared)
	wrappedInnerTags = GetTags(wrappedUnshared)
	wrappedOuterCtx = GetContext(wrappedOuter)
	wrappedOuterTags = GetTags(wrappedOuter)
	require.Equal(t, wrappedInnterCtx, Context{"req": {"id": "1", "path": "/v3"}})
	require.Equal(t, wrappedInnerTags, Tags{"id": "1", "kind": "bar"})
	require.Equal(t, wrappedOuterCtx, Context{"req": {"id": "1", "path": "/v2"}})
	require.Equal(t, wrappedOuterTags, Tags{"id": "1", "kind": "foo"})
}

// -----------------------------------------------------------------------------
// GetContext
// -----------------------------------------------------------------------------

func TestGetContext(t *testing.T) {
	t.Parallel()

	t.Run("NilErrorReturnsEmptyMap", func(t *testing.T) {
		assert.Len(t, GetContext(nil), 0)
	})

	t.Run("ExternalErrorReturnsEmptyMap", func(t *testing.T) {
		assert.Len(t, GetContext(errors.New("x")), 0)
	})

	t.Run("NoContextOnSingleErrorReturnsEmptyMap", func(t *testing.T) {
		e := mustErr(t, New("x"))
		assert.Len(t, GetContext(e), 0)
	})

	t.Run("SingleContextMapIsReturnedWithoutAllocatingAMergedCopy", func(t *testing.T) {
		baseCtx := Context{"req": {"id": "1"}}
		e := mustErr(t, New("x")).SetContexts(baseCtx)
		got := GetContext(e)
		assert.Equal(t, got, baseCtx)
		assert.True(t, isSameObject(baseCtx, got), "expected same map object, not a copy")
	})

	t.Run("BruhWrappedCtxerrorExposesInnerContext", func(t *testing.T) {
		inner := mustErr(t, New("inner")).
			SetContexts(Context{"req": {"id": "1"}})
		outer := bruh.Wrap(inner, "outer")
		assert.Equal(t, GetContext(outer), Context{"req": {"id": "1"}})
	})
}

// -----------------------------------------------------------------------------
// GetTags
// -----------------------------------------------------------------------------

// testTagsDumper implements error and tagsAppender for testing purposes.
type testTagsDumper struct{ err error }

func (d testTagsDumper) Error() string       { return d.err.Error() }
func (d testTagsDumper) Unwrap() error       { return d.err }
func (d testTagsDumper) AppendTags(out Tags) { out["dumped"] = "yes" }

func TestGetTags(t *testing.T) {
	t.Parallel()

	t.Run("NilErrorReturnsEmptyMap", func(t *testing.T) {
		assert.Len(t, GetTags(nil), 0)
	})

	t.Run("ExternalErrorReturnsEmptyMap", func(t *testing.T) {
		assert.Len(t, GetTags(errors.New("x")), 0)
	})

	t.Run("NoTagsOnSingleErrorReturnsEmptyMap", func(t *testing.T) {
		e := mustErr(t, New("x"))
		assert.Len(t, GetTags(e), 0)
	})

	t.Run("SingleTagsMapIsReturnedWithoutAllocatingAMergedCopy", func(t *testing.T) {
		baseTags := Tags{"a": "1"}
		e := mustErr(t, New("x")).SetTags(baseTags)
		got := GetTags(e)
		assert.Equal(t, got, baseTags)
		assert.True(t, isSameObject(baseTags, got), "expected same map object, not a copy")
	})

	t.Run("DistinctMapsInChainAreMergedWithOuterPrecedence", func(t *testing.T) {
		inner := mustErr(t, New("inner")).
			SetTags(Tags{"env": "prod", "zone": "eu"})
		outer := mustErr(t, Wrap(inner, "outer")).
			SetTags(Tags{"env": "staging", "op": "write"})
		got := GetTags(outer)
		outerTags := outer.(*Err).tags
		innerTags := inner.(*Err).tags
		assert.Equal(t, got, Tags{"env": "staging", "op": "write", "zone": "eu"})
		assert.True(
			t,
			isSameObject(got, outerTags),
			"expected the merged map to be the same object as outer.tags for efficiency",
		)
		assert.True(
			t,
			isSameObject(got, innerTags),
			"expected the merged map to be the same object as inner.tags for efficiency",
		)
	})

	// three-level merge: inner <- mid <- outer; outer should have precedence
	t.Run("ThreeLevelMergeOuterPrecedence", func(t *testing.T) {
		inner := mustErr(t, New("inner"))
		inner.SetTags(Tags{"a": "1", "b": "inner"})
		mid := mustErr(t, Wrap(inner, "mid"))
		mid.SetTags(Tags{"b": "mid", "c": "mid"})
		outer := mustErr(t, Wrap(mid, "outer"))
		outer.SetTags(Tags{"c": "outer", "d": "out"})

		got := GetTags(outer)
		exp := Tags{"a": "1", "b": "mid", "c": "outer", "d": "out"}
		assert.Equal(t, got, exp)
	})

	// TagsAppender: a wrapped error implementing AppendTags should not unexpectedly
	// mutate the returned map (covers tagsAppender branch in GetTags).
	t.Run("TagsAppenderBehaviour", func(t *testing.T) {
		inner := testTagsDumper{err: errors.New("inner")}
		outer := mustErr(t, Wrap(inner, "outer"))
		// allocationRequired should be true because inner implements tagsAppender
		got := GetTags(outer)
		val, has := got["dumped"]
		assert.True(t, has, "AppendTags must have contributed an entry into merged tags")
		assert.Equal(t, val, "yes")
	})
}

// -----------------------------------------------------------------------------
// GetContext and GetTags with foreign errors
// -----------------------------------------------------------------------------

// foreignContexter is a test helper that implements error and contexter.
type foreignContexter struct{ err error }

func (d foreignContexter) Error() string    { return d.err.Error() }
func (d foreignContexter) Context() Context { return Context{"foreign": {"contexter": "yes"}} }
func (d foreignContexter) Tags() Tags       { return Tags{"foreign": "tagser"} }

// foreignContexter is a test helper that implements error and contexter.
type foreignContexter2 struct{ err error }

func (d foreignContexter2) Error() string { return d.err.Error() }
func (d foreignContexter2) AppendContext(out Context) {
	out["foreign"] = map[string]any{"dumper": "yes"}
}
func (d foreignContexter2) AppendTags(out Tags) { out["foreign"] = "dumper" }

// nilPrivateWrapper is a test helper that implements privateContexter but
// returns nil maps to exercise initContext/initTags merge branches.
type nilPrivateWrapper struct{ err error }

func (w nilPrivateWrapper) Error() string           { return w.err.Error() }
func (w nilPrivateWrapper) Unwrap() error           { return w.err }
func (w nilPrivateWrapper) privateContext() Context { return nil }
func (w nilPrivateWrapper) privateTags() Tags       { return nil }

func TestGetContextAndTagsWithForeignError(t *testing.T) {
	t.Parallel()

	t.Run("ForeignContexter", func(t *testing.T) {
		inner := foreignContexter{err: errors.New("inner")}
		outer := Wrap(inner, "outer")
		expCtx := Context{"foreign": {"contexter": "yes"}}
		assert.Equal(t, GetContext(outer), expCtx)
		expTags := Tags{"foreign": "tagser"}
		assert.Equal(t, GetTags(outer), expTags)
	})

	t.Run("ForeignDumper", func(t *testing.T) {
		inner := foreignContexter2{err: errors.New("inner")}
		outer := Wrap(inner, "outer")
		expCtx := Context{"foreign": {"dumper": "yes"}}
		assert.Equal(t, GetContext(outer), expCtx)
		expTags := Tags{"foreign": "dumper"}
		assert.Equal(t, GetTags(outer), expTags)
	})
}

func TestGetContextAndTagsWithForeignErrorThroughBruhWrap(t *testing.T) {
	t.Parallel()

	t.Run("ContexterAndTagserAreCollectedWithoutCtxerrorInChain", func(t *testing.T) {
		err := bruh.Wrap(foreignContexter{err: errors.New("inner")}, "outer")
		assert.Equal(t, GetContext(err), Context{"foreign": {"contexter": "yes"}})
		assert.Equal(t, GetTags(err), Tags{"foreign": "tagser"})
	})

	t.Run("ContextAppenderAndTagsAppenderAreCollectedWithoutCtxerrorInChain", func(t *testing.T) {
		err := bruh.Wrap(foreignContexter2{err: errors.New("inner")}, "outer")
		assert.Equal(t, GetContext(err), Context{"foreign": {"dumper": "yes"}})
		assert.Equal(t, GetTags(err), Tags{"foreign": "dumper"})
	})

	t.Run("StandaloneTagsAppenderAllocatesAndAppends", func(t *testing.T) {
		err := bruh.Wrap(testTagsDumper{err: errors.New("inner")}, "outer")
		got := GetTags(err)
		assert.Equal(t, got["dumped"], "yes")
	})

	t.Run("InitContextMergesContexterAfterNilPrivateWrapper", func(t *testing.T) {
		inner := foreignContexter{err: errors.New("inner")}
		outer := mustErr(t, Wrap(nilPrivateWrapper{err: inner}, "outer")).
			SetContext("req", map[string]any{"id": "1"})

		got := GetContext(outer)
		exp := Context{
			"req":     {"id": "1"},
			"foreign": {"contexter": "yes"},
		}
		assert.Equal(t, got, exp)
	})

	t.Run("InitTagsMergesTagserAfterNilPrivateWrapper", func(t *testing.T) {
		inner := foreignContexter{err: errors.New("inner")}
		outer := mustErr(t, Wrap(nilPrivateWrapper{err: inner}, "outer")).
			SetTag("k", "v")

		got := GetTags(outer)
		assert.Equal(t, got, Tags{"k": "v", "foreign": "tagser"})
	})

	t.Run("InitTagsMergesTagsAppenderAfterNilPrivateWrapper", func(t *testing.T) {
		inner := foreignContexter2{err: errors.New("inner")}
		outer := mustErr(t, Wrap(nilPrivateWrapper{err: inner}, "outer")).
			SetTag("k", "v")

		got := GetTags(outer)
		assert.Equal(t, got, Tags{"k": "v", "foreign": "dumper"})
	})
}
