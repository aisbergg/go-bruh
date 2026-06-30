package ctxerror

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aisbergg/go-bruh/internal/testutils"
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
	require := testutils.NewRequire(t)

	root := errors.New("root")

	assertConstructor := func(name string, build func() error, expMsg string) {
		t.Run(name, func(t *testing.T) {
			got := build()
			require.NotNil(got, "expected non-nil")
			require.Equal(expMsg, got.Error(), "unexpected error message")
		})
	}

	assertNilConstructor := func(name string, build func() error) {
		t.Run(name, func(t *testing.T) {
			got := build()
			require.Nil(got, "expected nil")
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
	assert := testutils.NewAssert(t)

	var e *Err
	assert.Nil(e.SetContext("group", map[string]any{"k": "v"}))
	assert.Nil(e.SetContexts(Context{"group": {"k": "v"}}))
	assert.Nil(e.SetTag("k", "v"))
	assert.Nil(e.SetTags(Tags{"k": "v"}))
	assert.Nil(e.Unshare())
}

// -----------------------------------------------------------------------------
// SetContext / SetContexts
// -----------------------------------------------------------------------------

func TestContextModifiers(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("SetContextCreatesGroupWhenMissing", func(t *testing.T) {
		e := mustErr(t, New("x"))
		e.SetContext("req", map[string]any{"id": "1"})
		assert.Equal(Context{"req": {"id": "1"}}, GetContext(e))
	})

	t.Run("SetContextMergesIntoExistingGroupAndOverwritesDuplicateKeys", func(t *testing.T) {
		e := mustErr(t, New("x"))
		e.SetContext("req", map[string]any{"id": "a", "retry": false})
		e.SetContext("req", map[string]any{"id": "b", "path": "/x"})
		assert.Equal(Context{"req": {"id": "b", "retry": false, "path": "/x"}}, GetContext(e))
	})

	t.Run("SetContextsCreatesMissingGroupsAndMergesExistingOnes", func(t *testing.T) {
		e := mustErr(t, New("x"))
		e.SetContext("req", map[string]any{"id": "1"})
		e.SetContexts(Context{
			"req":  {"method": "GET"},
			"user": {"id": "u1"},
		})
		assert.Equal(Context{
			"req":  {"id": "1", "method": "GET"},
			"user": {"id": "u1"},
		}, GetContext(e))
	})
}

// -----------------------------------------------------------------------------
// AddTag / AddTags
// -----------------------------------------------------------------------------

func TestTagModifiers(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("AddTagInsertsAndOverwrites", func(t *testing.T) {
		e := mustErr(t, New("x"))
		e.SetTag("k", "v1")
		e.SetTag("k", "v2")
		assert.Equal(Tags{"k": "v2"}, GetTags(e))
	})

	t.Run("AddTagsMergesIntoExistingTags", func(t *testing.T) {
		e := mustErr(t, New("x"))
		e.SetTag("a", "1")
		e.SetTags(Tags{"b": "2", "a": "overwritten"})
		assert.Equal(Tags{"a": "overwritten", "b": "2"}, GetTags(e))
	})
}

// -----------------------------------------------------------------------------
// Owned vs shared metadata across the chain
// -----------------------------------------------------------------------------

func TestChainOwnsContextAndTagsByDefault(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	inner := mustErr(t, New("inner")).
		SetContext("req", map[string]any{"id": "1"}).
		SetTag("id", "1")

	outer := mustErr(t, Wrap(inner, "outer")).
		SetContext("req", map[string]any{"path": "/v1"}).
		SetTag("path", "/v1")

	innerCtx := GetContext(inner)
	outerCtx := GetContext(outer)
	// Default is shared: outer mutates same context object visible to inner.
	assert.True(isSameObject(innerCtx, outerCtx), "expected same context object by default")
	assert.Equal(Context{"req": {"id": "1", "path": "/v1"}}, innerCtx)
	assert.Equal(Context{"req": {"id": "1", "path": "/v1"}}, outerCtx)

	innerTags := GetTags(inner)
	outerTags := GetTags(outer)
	assert.True(isSameObject(innerTags, outerTags), "expected same tags object by default")
	assert.Equal(Tags{"id": "1", "path": "/v1"}, innerTags)
	assert.Equal(Tags{"id": "1", "path": "/v1"}, outerTags)
}

func TestUnshare(t *testing.T) {
	t.Parallel()
	require := testutils.NewRequire(t)

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
		isSameObject(wrappedUnshared.(*Err).context, base.(*Err).context),
		"expected wrappedUnshared to have same context object as base before unshare takes effect",
	)
	require.False(
		isSameObject(wrappedOuter.(*Err).context, base.(*Err).context),
		"expected wrappedOuter to have different context object than base due to unshare",
	)
	require.True(
		isSameObject(wrappedUnshared.(*Err).tags, base.(*Err).tags),
		"expected wrappedUnshared to have same tags object as base before unshare takes effect",
	)
	require.False(
		isSameObject(wrappedOuter.(*Err).tags, base.(*Err).tags),
		"expected wrappedOuter to have different tags object than base due to unshare",
	)

	require.Equal(Context{"req": {"id": "1", "path": "/v1"}}, wrappedInnterCtx)
	require.Equal(Tags{"id": "1", "kind": "foo"}, wrappedInnerTags)

	// wrappedOuter should reflect latest mutations to base since it shares metadata
	require.Equal(Context{"req": {"id": "1", "path": "/v2"}}, wrappedOuterCtx)
	require.Equal(Tags{"id": "1", "kind": "foo"}, wrappedOuterTags)

	// mutation of unshared should not affect outer
	wrappedUnshared.SetContext("req", map[string]any{"path": "/v3"})
	wrappedUnshared.SetTag("kind", "bar")
	wrappedInnterCtx = GetContext(wrappedUnshared)
	wrappedInnerTags = GetTags(wrappedUnshared)
	wrappedOuterCtx = GetContext(wrappedOuter)
	wrappedOuterTags = GetTags(wrappedOuter)
	require.Equal(Context{"req": {"id": "1", "path": "/v3"}}, wrappedInnterCtx)
	require.Equal(Tags{"id": "1", "kind": "bar"}, wrappedInnerTags)
	require.Equal(Context{"req": {"id": "1", "path": "/v2"}}, wrappedOuterCtx)
	require.Equal(Tags{"id": "1", "kind": "foo"}, wrappedOuterTags)
}

// -----------------------------------------------------------------------------
// GetContext
// -----------------------------------------------------------------------------

func TestGetContext(t *testing.T) {
	t.Parallel()

	t.Run("NilErrorReturnsEmptyMap", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		assert.Len(GetContext(nil), 0)
	})

	t.Run("ExternalErrorReturnsEmptyMap", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		assert.Len(GetContext(errors.New("x")), 0)
	})

	t.Run("NoContextOnSingleErrorReturnsEmptyMap", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		e := mustErr(t, New("x"))
		assert.Len(GetContext(e), 0)
	})

	t.Run("SingleContextMapIsReturnedWithoutAllocatingAMergedCopy", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		baseCtx := Context{"req": {"id": "1"}}
		e := mustErr(t, New("x")).SetContexts(baseCtx)
		got := GetContext(e)
		assert.Equal(baseCtx, got)
		assert.True(isSameObject(baseCtx, got), "expected same map object, not a copy")
	})

	t.Run("BruhWrappedCtxerrorExposesInnerContext", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		inner := mustErr(t, New("inner")).
			SetContexts(Context{"req": {"id": "1"}})
		outer := bruh.Wrap(inner, "outer")
		assert.Equal(Context{"req": {"id": "1"}}, GetContext(outer))
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
		assert := testutils.NewAssert(t)
		assert.Len(GetTags(nil), 0)
	})

	t.Run("ExternalErrorReturnsEmptyMap", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		assert.Len(GetTags(errors.New("x")), 0)
	})

	t.Run("NoTagsOnSingleErrorReturnsEmptyMap", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		e := mustErr(t, New("x"))
		assert.Len(GetTags(e), 0)
	})

	t.Run("SingleTagsMapIsReturnedWithoutAllocatingAMergedCopy", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		baseTags := Tags{"a": "1"}
		e := mustErr(t, New("x")).SetTags(baseTags)
		got := GetTags(e)
		assert.Equal(baseTags, got)
		assert.True(isSameObject(baseTags, got), "expected same map object, not a copy")
	})

	t.Run("DistinctMapsInChainAreMergedWithOuterPrecedence", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		inner := mustErr(t, New("inner")).
			SetTags(Tags{"env": "prod", "zone": "eu"})
		outer := mustErr(t, Wrap(inner, "outer")).
			SetTags(Tags{"env": "staging", "op": "write"})
		got := GetTags(outer)
		outerTags := outer.(*Err).tags
		innerTags := inner.(*Err).tags
		assert.Equal(Tags{"env": "staging", "op": "write", "zone": "eu"}, got)
		assert.True(
			isSameObject(got, outerTags),
			"expected the merged map to be the same object as outer.tags for efficiency",
		)
		assert.True(
			isSameObject(got, innerTags),
			"expected the merged map to be the same object as inner.tags for efficiency",
		)
	})

	// three-level merge: inner <- mid <- outer; outer should have precedence
	t.Run("ThreeLevelMergeOuterPrecedence", func(t *testing.T) {
		assert := testutils.NewAssert(t)
		inner := mustErr(t, New("inner"))
		inner.SetTags(Tags{"a": "1", "b": "inner"})
		mid := mustErr(t, Wrap(inner, "mid"))
		mid.SetTags(Tags{"b": "mid", "c": "mid"})
		outer := mustErr(t, Wrap(mid, "outer"))
		outer.SetTags(Tags{"c": "outer", "d": "out"})

		got := GetTags(outer)
		exp := Tags{"a": "1", "b": "mid", "c": "outer", "d": "out"}
		assert.Equal(exp, got)
	})

	// TagsAppender: a wrapped error implementing AppendTags should not unexpectedly
	// mutate the returned map (covers tagsAppender branch in GetTags).
	t.Run("TagsAppenderBehaviour", func(t *testing.T) {
		assert := testutils.NewAssert(t)

		inner := testTagsDumper{err: errors.New("inner")}
		outer := mustErr(t, Wrap(inner, "outer"))
		// allocationRequired should be true because inner implements tagsAppender
		got := GetTags(outer)
		val, has := got["dumped"]
		assert.True(has, "AppendTags must have contributed an entry into merged tags")
		assert.Equal("yes", val)
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
	assert := testutils.NewAssert(t)

	t.Run("ForeignContexter", func(t *testing.T) {
		inner := foreignContexter{err: errors.New("inner")}
		outer := Wrap(inner, "outer")
		expCtx := Context{"foreign": {"contexter": "yes"}}
		assert.Equal(expCtx, GetContext(outer))
		expTags := Tags{"foreign": "tagser"}
		assert.Equal(expTags, GetTags(outer))
	})

	t.Run("ForeignDumper", func(t *testing.T) {
		inner := foreignContexter2{err: errors.New("inner")}
		outer := Wrap(inner, "outer")
		expCtx := Context{"foreign": {"dumper": "yes"}}
		assert.Equal(expCtx, GetContext(outer))
		expTags := Tags{"foreign": "dumper"}
		assert.Equal(expTags, GetTags(outer))
	})
}

func TestGetContextAndTagsWithForeignErrorThroughBruhWrap(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("ContexterAndTagserAreCollectedWithoutCtxerrorInChain", func(t *testing.T) {
		err := bruh.Wrap(foreignContexter{err: errors.New("inner")}, "outer")
		assert.Equal(Context{"foreign": {"contexter": "yes"}}, GetContext(err))
		assert.Equal(Tags{"foreign": "tagser"}, GetTags(err))
	})

	t.Run("ContextAppenderAndTagsAppenderAreCollectedWithoutCtxerrorInChain", func(t *testing.T) {
		err := bruh.Wrap(foreignContexter2{err: errors.New("inner")}, "outer")
		assert.Equal(Context{"foreign": {"dumper": "yes"}}, GetContext(err))
		assert.Equal(Tags{"foreign": "dumper"}, GetTags(err))
	})

	t.Run("StandaloneTagsAppenderAllocatesAndAppends", func(t *testing.T) {
		err := bruh.Wrap(testTagsDumper{err: errors.New("inner")}, "outer")
		got := GetTags(err)
		assert.Equal("yes", got["dumped"])
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
		assert.Equal(exp, got)
	})

	t.Run("InitTagsMergesTagserAfterNilPrivateWrapper", func(t *testing.T) {
		inner := foreignContexter{err: errors.New("inner")}
		outer := mustErr(t, Wrap(nilPrivateWrapper{err: inner}, "outer")).
			SetTag("k", "v")

		got := GetTags(outer)
		assert.Equal(Tags{"k": "v", "foreign": "tagser"}, got)
	})

	t.Run("InitTagsMergesTagsAppenderAfterNilPrivateWrapper", func(t *testing.T) {
		inner := foreignContexter2{err: errors.New("inner")}
		outer := mustErr(t, Wrap(nilPrivateWrapper{err: inner}, "outer")).
			SetTag("k", "v")

		got := GetTags(outer)
		assert.Equal(Tags{"k": "v", "foreign": "dumper"}, got)
	})
}
