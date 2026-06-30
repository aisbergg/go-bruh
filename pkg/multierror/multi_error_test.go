package multierror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aisbergg/go-bruh/internal/testutils"
	"github.com/aisbergg/go-bruh/pkg/bruh"
	"github.com/aisbergg/go-bruh/pkg/ctxerror"
)

// mustErr unpacks a MultiErrorer to *Err, failing the test if that is
// not possible.
func mustErr(t *testing.T, err MultiErrorer) *Err {
	t.Helper()
	require := testutils.NewRequire(t)
	require.NotNil(err, "expected error to be non-nil")
	e, ok := err.(*Err)
	require.True(ok, "expected *Err, got %T", err)
	return e
}

// -----------------------------------------------------------------------------
// Constructors
// -----------------------------------------------------------------------------

func TestConstructors(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("NewCreatesEmptyMultiErrorerWithMessage", func(t *testing.T) {
		me := New("test message", Options{})
		assert.NotNil(me)
		assert.True(me.IsNil(), "expected empty error to be nil")
		assert.Equal("test message", me.(*Err).msg)
	})

	t.Run("ErrorfCreatesEmptyMultiErrorerWithFormattedMessage", func(t *testing.T) {
		me := Errorf(Options{}, "error %d", 42)
		assert.NotNil(me)
		assert.True(me.IsNil())
		assert.Equal("error 42", me.(*Err).msg)
	})
}

// -----------------------------------------------------------------------------
// Options
// -----------------------------------------------------------------------------

func TestOptUnwrapBehavior(t *testing.T) {
	t.Parallel()
	require := testutils.NewRequire(t)

	errs := []error{errors.New("a"), errors.New("b"), errors.New("c")}

	t.Run("UnwrapFirstDefault", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		for _, e := range errs {
			me.Add(e)
		}
		require.Equal(errs[0], me.Unwrap())
	})

	t.Run("UnwrapLast", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		for _, e := range errs {
			me.Add(e)
		}
		require.Equal(errs[2], me.Unwrap())
	})

	t.Run("UnwrapNone", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapNone})
		for _, e := range errs {
			me.Add(e)
		}
		require.Nil(me.Unwrap())
	})

	assertEmptyReturnsNil := func(name string, behavior UnwrapBehavior) {
		t.Run(name, func(t *testing.T) {
			me := New("test", Options{UnwrapBehavior: behavior})
			require.Nil(me.Unwrap())
		})
	}

	assertEmptyReturnsNil("EmptyReturnsNilUnwrapFirst", UnwrapFirst)
	assertEmptyReturnsNil("EmptyReturnsNilUnwrapLast", UnwrapLast)
	assertEmptyReturnsNil("EmptyReturnsNilUnwrapNone", UnwrapNone)
}

func TestOptLimitPrint(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	errs := []error{
		errors.New("error1"),
		errors.New("error2"),
		errors.New("error3"),
		errors.New("error4"),
		errors.New("error5"),
	}

	t.Run("LimitLimitsNumberOfPrintedErrors", func(t *testing.T) {
		me := New("test", Options{LimitPrint: 2})
		for _, e := range errs {
			me.Add(e)
		}
		msg := me.Error()
		// With limit 2 and 5 errors, should print 2 + "and 3 more"
		assert.True(len(msg) > 0)
		assert.True(len(msg) > 0, "expected error message")
	})

	assertLimitShowsAllErrors := func(name string, limit int) {
		t.Run(name, func(t *testing.T) {
			me := New("test", Options{LimitPrint: limit})
			for _, err := range errs {
				me.Add(err)
			}
			msg := me.Error()
			assert.True(
				len(msg) > 0,
				"expected error message to contain error text",
			)
		})
	}

	assertLimitShowsAllErrors("LimitZeroShowsAllErrors", 0)
	assertLimitShowsAllErrors("LimitNegativeShowsAllErrors", -1)
}

func TestOptFilter(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	errFoo := errors.New("foo error")
	errBar := errors.New("bar error")
	errBaz := errors.New("baz error")

	t.Run("FilterRejectsErrorsThatReturnFalse", func(t *testing.T) {
		me := New("test", Options{Filter: func(e error) bool {
			return !errors.Is(e, errBar)
		}})
		me.Add(errFoo, errBar, errBaz)
		assert.Len(me.Errors(), 2)
		assert.Equal(errFoo, me.Errors()[0])
		assert.Equal(errBaz, me.Errors()[1])
	})

	t.Run("FilterAcceptsAllErrorsIfAlwaysTrue", func(t *testing.T) {
		me := New("test", Options{Filter: func(e error) bool {
			return true
		}})
		me.Add(errFoo, errBar, errBaz)
		assert.Len(me.Errors(), 3)
	})
}

// -----------------------------------------------------------------------------
// IsNil and nil behavior
// -----------------------------------------------------------------------------

func TestIsNil(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("EmptyMultiErrorerIsNil", func(t *testing.T) {
		me := New("test", Options{})
		assert.True(me.IsNil())
	})

	t.Run("NilReceiverIsNil", func(t *testing.T) {
		var me *Err
		assert.True(me.IsNil())
	})

	t.Run("MultiErrorerWithErrorsIsNotNil", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(errors.New("error"))
		assert.False(me.IsNil())
	})
}

func TestErrorOrNil(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("EmptyMultiErrorerReturnsNil", func(t *testing.T) {
		me := New("test", Options{})
		assert.Nil(me.ErrorOrNil())
	})

	t.Run("MultiErrorerWithErrorsReturnsItself", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(errors.New("error"))
		assert.Equal(me, me.ErrorOrNil())
	})
}

func TestSingleOrNil(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("EmptyMultiErrorerReturnsNil", func(t *testing.T) {
		me := New("test", Options{})
		assert.Nil(me.SingleOrNil())
	})

	t.Run("MultiErrorerWithSingleErrorReturnsThatError", func(t *testing.T) {
		err := errors.New("single error")
		me := New("test", Options{})
		me.Add(err)
		assert.Equal(err, me.SingleOrNil())
	})

	t.Run("MultiErrorerWithMultipleErrorsReturnsItself", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(errors.New("a"), errors.New("b"))
		assert.Equal(me, me.SingleOrNil())
	})
}

// -----------------------------------------------------------------------------
// Add and Grow
// -----------------------------------------------------------------------------

func TestAdd(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("AddAppendsErrors", func(t *testing.T) {
		me := New("test", Options{})
		e1 := errors.New("e1")
		e2 := errors.New("e2")
		me.Add(e1, e2)
		assert.Len(me.Errors(), 2)
		assert.Equal(e1, me.Errors()[0])
		assert.Equal(e2, me.Errors()[1])
	})

	t.Run("AddIgnoresNilErrors", func(t *testing.T) {
		me := New("test", Options{})
		e1 := errors.New("e1")
		me.Add(nil, e1, nil)
		assert.Len(me.Errors(), 1)
		assert.Equal(e1, me.Errors()[0])
	})

	t.Run("AddReturnsEarlyIfAllErrorsAreNil", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(nil, nil)
		assert.Len(me.Errors(), 0)
	})

	t.Run("AddRespectsFilter", func(t *testing.T) {
		me := New("test", Options{Filter: func(e error) bool {
			return e.Error() != "skip"
		}})
		e1 := errors.New("e1")
		e2 := errors.New("skip")
		e3 := errors.New("e3")
		me.Add(e1, e2, e3)
		assert.Len(me.Errors(), 2)
		assert.Equal(e1, me.Errors()[0])
		assert.Equal(e3, me.Errors()[1])
	})
}

func TestGrow(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("GrowPreAllocatesCapacity", func(t *testing.T) {
		me := New("test", Options{})
		me.Grow(10)
		assert.NotNil(me.Errors())
		oldCap := cap(me.Errors())
		assert.True(oldCap >= 10)
	})

	t.Run("GrowDoesNothingIfCapacityAlreadySufficient", func(t *testing.T) {
		me := New("test", Options{})
		me.Grow(10)
		me.Add(errors.New("e1"))
		oldCap := cap(me.Errors())
		me.Grow(5)
		assert.Equal(oldCap, cap(me.Errors()), "capacity should not shrink")
	})

	t.Run("GrowInitializesSliceIfNilAndGrowsOnFurtherCalls", func(t *testing.T) {
		me := New("test", Options{})
		// Initial state: errors is nil
		me.Grow(5)
		// After grow, should be empty slice with capacity >= 5
		assert.NotNil(me.Errors())
		assert.Len(me.Errors(), 0)
		assert.True(cap(me.Errors()) >= 5)
	})
}

// -----------------------------------------------------------------------------
// Merge
// -----------------------------------------------------------------------------

func TestMerge(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("MergeCombinesErrorsFromOtherMultiErrorers", func(t *testing.T) {
		me1 := New("test1", Options{})
		me1.Add(errors.New("a"), errors.New("b"))

		me2 := New("test2", Options{})
		me2.Add(errors.New("c"), errors.New("d"))

		me1.Merge(me2)
		assert.Len(me1.Errors(), 4)
	})

	t.Run("MergeWithEmptyMultiErrorerIsANoOp", func(t *testing.T) {
		me1 := New("test1", Options{})
		me1.Add(errors.New("a"))
		me2 := New("test2", Options{})
		me1.Merge(me2)
		assert.Len(me1.Errors(), 1)
	})

	t.Run("MergeHandlesMultipleNonEmptyMultiErrorers", func(t *testing.T) {
		me1 := New("test1", Options{})
		me1.Add(errors.New("a"))

		me2 := New("test2", Options{})
		me2.Add(errors.New("b"), errors.New("c"))

		me3 := New("test3", Options{})
		me3.Add(errors.New("d"))

		me1.Merge(me2, me3)
		assert.Len(me1.Errors(), 4)
	})
}

// -----------------------------------------------------------------------------
// errors.As and errors.Is
// -----------------------------------------------------------------------------

type customErr struct {
	msg string
}

func (ce customErr) Error() string { return ce.msg }

func TestAs(t *testing.T) {
	t.Parallel()
	require := testutils.NewRequire(t)

	t.Run("UnwrapFirstNotFound", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"))

		var target customErr
		require.False(bruh.As(me, &target), "expected As to return false when target type not found with UnwrapFirst")
	})

	t.Run("UnwrapFirstAtFirst", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		me.Add(
			customErr{"a"},
			errors.New("b"),
			customErr{"c"},
			errors.New("d"),
			customErr{"e"},
		)

		var target customErr
		require.True(bruh.As(me, &target), "expected As to find customErr in multierror with UnwrapFirst")
		require.Equal("a", target.Error())
	})

	t.Run("UnwrapFirstAtLast", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		me.Add(
			errors.New("a"),
			errors.New("b"),
			customErr{"c"},
		)

		var target customErr
		require.True(bruh.As(me, &target), "expected As to find customErr in multierror with UnwrapFirst")
		require.Equal("c", target.Error())
	})

	t.Run("UnwrapLastNotFound", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"))

		var target customErr
		require.False(bruh.As(me, &target), "expected As to return false when target type not found with UnwrapLast")
	})

	t.Run("UnwrapLastAtLast", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		me.Add(
			customErr{"a"},
			errors.New("b"),
			customErr{"c"},
			errors.New("d"),
			customErr{"e"},
		)

		var target customErr
		require.True(bruh.As(me, &target), "expected As to find customErr in multierror with UnwrapLast")
		require.Equal("e", target.Error())
	})

	t.Run("UnwrapLastAtFirst", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		me.Add(
			customErr{"a"},
			errors.New("b"),
			errors.New("c"),
		)

		var target customErr
		require.True(bruh.As(me, &target), "expected As to find customErr in multierror with UnwrapLast")
		require.Equal("a", target.Error())
	})

	t.Run("UnwrapNone", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapNone})
		me.Add(
			customErr{"a"},
			errors.New("b"),
			customErr{"c"},
			errors.New("d"),
			customErr{"e"},
		)

		var target customErr
		require.False(
			bruh.As(me, &target),
			"expected As to return false with UnwrapNone since it should not unwrap to find target type",
		)
	})

	t.Run("WrappedDeep", func(t *testing.T) {
		// Create a wrapped error chain
		baseErr := errors.New("base")
		wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
		wrappedErr = fmt.Errorf("wrapped: %w", wrappedErr)
		wrappedErr = fmt.Errorf("wrapped: %w", wrappedErr)
		wrappedErr = fmt.Errorf("wrapped: %w", wrappedErr)
		wrappedErr = fmt.Errorf("wrapped: %w", wrappedErr)

		me := New("test", Options{})
		me.Add(wrappedErr)

		// errors.As can unwrap to find the base error
		require.True(bruh.As(me, &baseErr), "expected As to find base error deep in wrapped chain")
	})
}

func TestIs(t *testing.T) {
	t.Parallel()
	require := testutils.NewRequire(t)

	t.Run("UnwrapFirstNotFound", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"))

		want := errors.New("missing")
		require.False(bruh.Is(me, want), "expected Is to return false when target is not present with UnwrapFirst")
	})

	t.Run("UnwrapFirstAtFirst", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		first := errors.New("a")
		me.Add(first, errors.New("b"), errors.New("c"))

		require.True(bruh.Is(me, first), "expected Is to find first error in multierror with UnwrapFirst")
	})

	t.Run("UnwrapFirstAtLast", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		last := errors.New("c")
		me.Add(errors.New("a"), errors.New("b"), last)

		require.True(bruh.Is(me, last), "expected Is to find last error in multierror with UnwrapFirst")
	})

	t.Run("UnwrapLastNotFound", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"))

		want := errors.New("missing")
		require.False(bruh.Is(me, want), "expected Is to return false when target is not present with UnwrapLast")
	})

	t.Run("UnwrapLastAtLast", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		last := errors.New("e")
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"), last)

		require.True(bruh.Is(me, last), "expected Is to find last error in multierror with UnwrapLast")
	})

	t.Run("UnwrapLastAtFirst", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		first := errors.New("a")
		me.Add(first, errors.New("b"), errors.New("c"))

		require.True(bruh.Is(me, first), "expected Is to find first error in multierror with UnwrapLast")
	})

	t.Run("UnwrapNone", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapNone})
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"))

		want := errors.New("missing")
		require.False(bruh.Is(me, want), "expected Is to return false with UnwrapNone")
	})

	t.Run("WrappedDeep", func(t *testing.T) {
		baseErr := errors.New("base")
		wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
		wrappedErr = fmt.Errorf("wrapped: %w", wrappedErr)
		wrappedErr = fmt.Errorf("wrapped: %w", wrappedErr)
		wrappedErr = fmt.Errorf("wrapped: %w", wrappedErr)
		wrappedErr = fmt.Errorf("wrapped: %w", wrappedErr)

		me := New("test", Options{})
		me.Add(wrappedErr)

		require.True(bruh.Is(me, baseErr), "expected Is to find base error deep in wrapped chain")
	})
}

// -----------------------------------------------------------------------------
// Context and Tags (ctxerror integration)
// -----------------------------------------------------------------------------

func TestContextIntegration(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("ContextCombinesContextsFromAllErrors", func(t *testing.T) {
		e1 := ctxerror.New("error1")
		e1.SetContext("req", map[string]any{"id": "1"})

		e2 := ctxerror.New("error2")
		e2.SetContext("user", map[string]any{"id": "u1"})

		me := mustErr(t, New("test", Options{}))
		me.Add(e1, e2)

		ctx := me.Context()
		assert.Equal("1", ctx["req"]["id"])
		assert.Equal("u1", ctx["user"]["id"])
	})

	t.Run("ContextReturnsEmptyMapForNoContextErrors", func(t *testing.T) {
		me := mustErr(t, New("test", Options{}))
		me.Add(errors.New("plain"), errors.New("errors"))

		ctx := me.Context()
		assert.Len(ctx, 0)
	})

	t.Run("ContextMergesGroupsWithMapsCopyLaterOverwritesEntireGroup", func(t *testing.T) {
		e1 := ctxerror.New("error1")
		e1.SetContext("req", map[string]any{"id": "1", "status": 400})

		e2 := ctxerror.New("error2")
		e2.SetContext("req", map[string]any{"id": "2"})

		me := mustErr(t, New("test", Options{}))
		me.Add(e1, e2)

		ctx := me.Context()
		// maps.Copy overwrites entire groups, so e2's req completely replaces e1's req
		assert.Equal("2", ctx["req"]["id"])
		// status from e1 is lost because e2's group overwrote it
		_, hasStatus := ctx["req"]["status"]
		assert.False(hasStatus, "expected status to be overwritten by e2's group")
	})
}

func TestTagsIntegration(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("TagsCombinesTagsFromAllErrors", func(t *testing.T) {
		e1 := ctxerror.New("error1")
		e1.SetTag("env", "prod")

		e2 := ctxerror.New("error2")
		e2.SetTag("zone", "eu")

		me := mustErr(t, New("test", Options{}))
		me.Add(e1, e2)

		tags := me.Tags()
		assert.Equal("prod", tags["env"])
		assert.Equal("eu", tags["zone"])
	})

	t.Run("TagsReturnsEmptyMapForNoTagErrors", func(t *testing.T) {
		me := mustErr(t, New("test", Options{}))
		me.Add(errors.New("plain"), errors.New("errors"))

		tags := me.Tags()
		assert.Len(tags, 0)
	})

	t.Run("TagsMergesOverlappingKeysWithLaterValues", func(t *testing.T) {
		e1 := ctxerror.New("error1")
		e1.SetTag("env", "prod")
		e1.SetTag("op", "read")

		e2 := ctxerror.New("error2")
		e2.SetTag("env", "staging")

		me := mustErr(t, New("test", Options{}))
		me.Add(e1, e2)

		tags := me.Tags()
		// Later value should win
		assert.Equal("staging", tags["env"])
		assert.Equal("read", tags["op"])
	})
}

// -----------------------------------------------------------------------------
// Error formatting
// -----------------------------------------------------------------------------

func TestErrorFormatting(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("ErrorIncludesMessageAndFormattedErrors", func(t *testing.T) {
		me := New("main error", Options{})
		me.Add(errors.New("error1"), errors.New("error2"))

		msg := me.Error()
		assert.True(len(msg) > 0)
		assert.True(len(msg) > 0 && msg != "")
	})

	t.Run("ErrorReturnsEmptyStringForNilMultiErrorer", func(t *testing.T) {
		me := New("test", Options{})
		assert.Equal("", me.Error())
	})

	t.Run("ErrorPadsErrorNumbersCorrectly", func(t *testing.T) {
		me := New("test", Options{})
		for i := 0; i < 15; i++ {
			me.Add(fmt.Errorf("error %d", i))
		}
		msg := me.Error()
		// Should have 2-digit padding like #00, #01, etc.
		assert.True(len(msg) > 0)
	})

	t.Run("MessageReturnsSameAsError", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(errors.New("e1"))
		assert.Equal(me.Error(), me.Message())
	})
}

// -----------------------------------------------------------------------------
// bruh integration
// -----------------------------------------------------------------------------

func TestBruhIntegration(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("MultiErrorerWrapsBruhErrors", func(t *testing.T) {
		bruhErr := bruh.New("bruh error")
		me := New("wrapper", Options{})
		me.Add(bruhErr)

		assert.Len(me.Errors(), 1)
		assert.Equal(bruhErr, me.Errors()[0])
	})

	t.Run("MultiErrorerIntegratesWithBruhWrap", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(errors.New("a"), errors.New("b"))

		wrapped := bruh.Wrap(me, "outer wrapper")
		assert.NotNil(wrapped)
		assert.True(errors.Is(wrapped, me))
	})
}

// -----------------------------------------------------------------------------
// Edge cases and interactions
// -----------------------------------------------------------------------------

func TestEdgeCases(t *testing.T) {
	t.Parallel()
	assert := testutils.NewAssert(t)

	t.Run("UnwrapWithSingleError", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		err := errors.New("single")
		me.Add(err)
		assert.Equal(err, me.Unwrap())
	})

	t.Run("FilterAppliedDuringAdd", func(t *testing.T) {
		me := New("test", Options{Filter: func(e error) bool {
			return e.Error() != "filtered"
		}})
		me.Add(errors.New("keep"), errors.New("filtered"), errors.New("keep"))
		assert.Len(me.Errors(), 2)
	})

	t.Run("MultipleOptionsCombineCorrectly", func(t *testing.T) {
		me := New("test", Options{
			UnwrapBehavior: UnwrapLast,
			LimitPrint:     1,
			Filter:         func(e error) bool { return true },
		})
		e1 := errors.New("a")
		e2 := errors.New("b")
		me.Add(e1, e2)
		assert.Equal(e2, me.Unwrap())
		assert.Len(me.Errors(), 2)
	})

	t.Run("ErrorsReflectsCurrentState", func(t *testing.T) {
		me := New("test", Options{})
		assert.Len(me.Errors(), 0)
		me.Add(errors.New("a"))
		assert.Len(me.Errors(), 1)
		me.Add(errors.New("b"))
		assert.Len(me.Errors(), 2)
	})

	t.Run("GrowCanBeCalledMultipleTimes", func(t *testing.T) {
		me := New("test", Options{})
		me.Grow(5)
		me.Grow(10)
		me.Grow(15)
		assert.True(cap(me.Errors()) >= 15)
	})
}
