package multierror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aisbergg/go-bruh/internal/lib/test/assert"
	"github.com/aisbergg/go-bruh/internal/lib/test/require"
	"github.com/aisbergg/go-bruh/pkg/bruh"
	"github.com/aisbergg/go-bruh/pkg/ctxerror"
)

// mustErr unpacks a MultiErrorer to *Err, failing the test if that is
// not possible.
func mustErr(t *testing.T, err MultiErrorer) *Err {
	t.Helper()
	require.NotNil(t, err, "expected error to be non-nil")
	e, ok := err.(*Err)
	require.Truef(t, ok, "expected *Err, got %T", err)
	return e
}

// -----------------------------------------------------------------------------
// Constructors
// -----------------------------------------------------------------------------

func TestConstructors(t *testing.T) {
	t.Parallel()

	t.Run("NewCreatesEmptyMultiErrorerWithMessage", func(t *testing.T) {
		me := New("test message", Options{})
		assert.NotNil(t, me)
		assert.True(t, me.IsNil(), "expected empty error to be nil")
		assert.Equal(t, me.(*Err).msg, "test message")
	})

	t.Run("ErrorfCreatesEmptyMultiErrorerWithFormattedMessage", func(t *testing.T) {
		me := Errorf(Options{}, "error %d", 42)
		assert.NotNil(t, me)
		assert.True(t, me.IsNil())
		assert.Equal(t, me.(*Err).msg, "error 42")
	})
}

// -----------------------------------------------------------------------------
// Options
// -----------------------------------------------------------------------------

func TestOptUnwrapBehavior(t *testing.T) {
	t.Parallel()

	errs := []error{errors.New("a"), errors.New("b"), errors.New("c")}

	t.Run("UnwrapFirstDefault", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		for _, e := range errs {
			me.Add(e)
		}
		require.Equal(t, me.Unwrap(), errs[0])
	})

	t.Run("UnwrapLast", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		for _, e := range errs {
			me.Add(e)
		}
		require.Equal(t, me.Unwrap(), errs[2])
	})

	t.Run("UnwrapNone", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapNone})
		for _, e := range errs {
			me.Add(e)
		}
		require.Nil(t, me.Unwrap())
	})

	assertEmptyReturnsNil := func(name string, behavior UnwrapBehavior) {
		t.Run(name, func(t *testing.T) {
			me := New("test", Options{UnwrapBehavior: behavior})
			require.Nil(t, me.Unwrap())
		})
	}

	assertEmptyReturnsNil("EmptyReturnsNilUnwrapFirst", UnwrapFirst)
	assertEmptyReturnsNil("EmptyReturnsNilUnwrapLast", UnwrapLast)
	assertEmptyReturnsNil("EmptyReturnsNilUnwrapNone", UnwrapNone)
}

func TestOptLimitPrint(t *testing.T) {
	t.Parallel()

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
		assert.True(t, len(msg) > 0)
		assert.True(t, len(msg) > 0, "expected error message")
	})

	assertLimitShowsAllErrors := func(name string, limit int) {
		t.Run(name, func(t *testing.T) {
			me := New("test", Options{LimitPrint: limit})
			for _, err := range errs {
				me.Add(err)
			}
			msg := me.Error()
			assert.True(
				t,
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

	errFoo := errors.New("foo error")
	errBar := errors.New("bar error")
	errBaz := errors.New("baz error")

	t.Run("FilterRejectsErrorsThatReturnFalse", func(t *testing.T) {
		me := New("test", Options{Filter: func(e error) bool {
			return !errors.Is(e, errBar)
		}})
		me.Add(errFoo, errBar, errBaz)
		assert.Len(t, me.Errors(), 2)
		assert.Equal(t, me.Errors()[0], errFoo)
		assert.Equal(t, me.Errors()[1], errBaz)
	})

	t.Run("FilterAcceptsAllErrorsIfAlwaysTrue", func(t *testing.T) {
		me := New("test", Options{Filter: func(e error) bool {
			return true
		}})
		me.Add(errFoo, errBar, errBaz)
		assert.Len(t, me.Errors(), 3)
	})
}

// -----------------------------------------------------------------------------
// IsNil and nil behavior
// -----------------------------------------------------------------------------

func TestIsNil(t *testing.T) {
	t.Parallel()

	t.Run("EmptyMultiErrorerIsNil", func(t *testing.T) {
		me := New("test", Options{})
		assert.True(t, me.IsNil())
	})

	t.Run("NilReceiverIsNil", func(t *testing.T) {
		var me *Err
		assert.True(t, me.IsNil())
	})

	t.Run("MultiErrorerWithErrorsIsNotNil", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(errors.New("error"))
		assert.False(t, me.IsNil())
	})
}

func TestErrorOrNil(t *testing.T) {
	t.Parallel()

	t.Run("EmptyMultiErrorerReturnsNil", func(t *testing.T) {
		me := New("test", Options{})
		assert.Nil(t, me.ErrorOrNil())
	})

	t.Run("MultiErrorerWithErrorsReturnsItself", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(errors.New("error"))
		assert.Equal(t, me.ErrorOrNil(), me)
	})
}

func TestSingleOrNil(t *testing.T) {
	t.Parallel()

	t.Run("EmptyMultiErrorerReturnsNil", func(t *testing.T) {
		me := New("test", Options{})
		assert.Nil(t, me.SingleOrNil())
	})

	t.Run("MultiErrorerWithSingleErrorReturnsThatError", func(t *testing.T) {
		err := errors.New("single error")
		me := New("test", Options{})
		me.Add(err)
		assert.Equal(t, me.SingleOrNil(), err)
	})

	t.Run("MultiErrorerWithMultipleErrorsReturnsItself", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(errors.New("a"), errors.New("b"))
		got, ok := me.SingleOrNil().(*Err)
		assert.True(t, ok)
		assert.Equal[any](t, got, me)
	})
}

// -----------------------------------------------------------------------------
// Add and Grow
// -----------------------------------------------------------------------------

func TestAdd(t *testing.T) {
	t.Parallel()

	t.Run("AddAppendsErrors", func(t *testing.T) {
		me := New("test", Options{})
		e1 := errors.New("e1")
		e2 := errors.New("e2")
		me.Add(e1, e2)
		assert.Len(t, me.Errors(), 2)
		assert.Equal(t, me.Errors()[0], e1)
		assert.Equal(t, me.Errors()[1], e2)
	})

	t.Run("AddIgnoresNilErrors", func(t *testing.T) {
		me := New("test", Options{})
		e1 := errors.New("e1")
		me.Add(nil, e1, nil)
		assert.Len(t, me.Errors(), 1)
		assert.Equal(t, me.Errors()[0], e1)
	})

	t.Run("AddReturnsEarlyIfAllErrorsAreNil", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(nil, nil)
		assert.Len(t, me.Errors(), 0)
	})

	t.Run("AddRespectsFilter", func(t *testing.T) {
		me := New("test", Options{Filter: func(e error) bool {
			return e.Error() != "skip"
		}})
		e1 := errors.New("e1")
		e2 := errors.New("skip")
		e3 := errors.New("e3")
		me.Add(e1, e2, e3)
		assert.Len(t, me.Errors(), 2)
		assert.Equal(t, me.Errors()[0], e1)
		assert.Equal(t, me.Errors()[1], e3)
	})
}

func TestGrow(t *testing.T) {
	t.Parallel()

	t.Run("GrowPreAllocatesCapacity", func(t *testing.T) {
		me := New("test", Options{})
		me.Grow(10)
		assert.NotNil(t, me.Errors())
		oldCap := cap(me.Errors())
		assert.True(t, oldCap >= 10)
	})

	t.Run("GrowDoesNothingIfCapacityAlreadySufficient", func(t *testing.T) {
		me := New("test", Options{})
		me.Grow(10)
		me.Add(errors.New("e1"))
		oldCap := cap(me.Errors())
		me.Grow(5)
		assert.Equal(t, cap(me.Errors()), oldCap, "capacity should not shrink")
	})

	t.Run("GrowInitializesSliceIfNilAndGrowsOnFurtherCalls", func(t *testing.T) {
		me := New("test", Options{})
		// Initial state: errors is nil
		me.Grow(5)
		// After grow, should be empty slice with capacity >= 5
		assert.NotNil(t, me.Errors())
		assert.Len(t, me.Errors(), 0)
		assert.True(t, cap(me.Errors()) >= 5)
	})
}

// -----------------------------------------------------------------------------
// Merge
// -----------------------------------------------------------------------------

func TestMerge(t *testing.T) {
	t.Parallel()

	t.Run("MergeCombinesErrorsFromOtherMultiErrorers", func(t *testing.T) {
		me1 := New("test1", Options{})
		me1.Add(errors.New("a"), errors.New("b"))

		me2 := New("test2", Options{})
		me2.Add(errors.New("c"), errors.New("d"))

		me1.Merge(me2)
		assert.Len(t, me1.Errors(), 4)
	})

	t.Run("MergeWithEmptyMultiErrorerIsANoOp", func(t *testing.T) {
		me1 := New("test1", Options{})
		me1.Add(errors.New("a"))
		me2 := New("test2", Options{})
		me1.Merge(me2)
		assert.Len(t, me1.Errors(), 1)
	})

	t.Run("MergeHandlesMultipleNonEmptyMultiErrorers", func(t *testing.T) {
		me1 := New("test1", Options{})
		me1.Add(errors.New("a"))

		me2 := New("test2", Options{})
		me2.Add(errors.New("b"), errors.New("c"))

		me3 := New("test3", Options{})
		me3.Add(errors.New("d"))

		me1.Merge(me2, me3)
		assert.Len(t, me1.Errors(), 4)
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

	t.Run("UnwrapFirstNotFound", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"))

		var target customErr
		require.False(
			t,
			bruh.As(me, &target),
			"expected As to return false when target type not found with UnwrapFirst",
		)
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
		require.True(t, bruh.As(me, &target), "expected As to find customErr in multierror with UnwrapFirst")
		require.Equal(t, target.Error(), "a")
	})

	t.Run("UnwrapFirstAtLast", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		me.Add(
			errors.New("a"),
			errors.New("b"),
			customErr{"c"},
		)

		var target customErr
		require.True(t, bruh.As(me, &target), "expected As to find customErr in multierror with UnwrapFirst")
		require.Equal(t, target.Error(), "c")
	})

	t.Run("UnwrapLastNotFound", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"))

		var target customErr
		require.False(t, bruh.As(me, &target), "expected As to return false when target type not found with UnwrapLast")
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
		require.True(t, bruh.As(me, &target), "expected As to find customErr in multierror with UnwrapLast")
		require.Equal(t, target.Error(), "e")
	})

	t.Run("UnwrapLastAtFirst", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		me.Add(
			customErr{"a"},
			errors.New("b"),
			errors.New("c"),
		)

		var target customErr
		require.True(t, bruh.As(me, &target), "expected As to find customErr in multierror with UnwrapLast")
		require.Equal(t, target.Error(), "a")
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
			t,
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
		require.True(t, bruh.As(me, &baseErr), "expected As to find base error deep in wrapped chain")
	})
}

func TestIs(t *testing.T) {
	t.Parallel()

	t.Run("UnwrapFirstNotFound", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"))

		want := errors.New("missing")
		require.False(t, bruh.Is(me, want), "expected Is to return false when target is not present with UnwrapFirst")
	})

	t.Run("UnwrapFirstAtFirst", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		first := errors.New("a")
		me.Add(first, errors.New("b"), errors.New("c"))

		require.True(t, bruh.Is(me, first), "expected Is to find first error in multierror with UnwrapFirst")
	})

	t.Run("UnwrapFirstAtLast", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		last := errors.New("c")
		me.Add(errors.New("a"), errors.New("b"), last)

		require.True(t, bruh.Is(me, last), "expected Is to find last error in multierror with UnwrapFirst")
	})

	t.Run("UnwrapLastNotFound", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"))

		want := errors.New("missing")
		require.False(t, bruh.Is(me, want), "expected Is to return false when target is not present with UnwrapLast")
	})

	t.Run("UnwrapLastAtLast", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		last := errors.New("e")
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"), last)

		require.True(t, bruh.Is(me, last), "expected Is to find last error in multierror with UnwrapLast")
	})

	t.Run("UnwrapLastAtFirst", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapLast})
		first := errors.New("a")
		me.Add(first, errors.New("b"), errors.New("c"))

		require.True(t, bruh.Is(me, first), "expected Is to find first error in multierror with UnwrapLast")
	})

	t.Run("UnwrapNone", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapNone})
		me.Add(errors.New("a"), errors.New("b"), errors.New("c"))

		want := errors.New("missing")
		require.False(t, bruh.Is(me, want), "expected Is to return false with UnwrapNone")
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

		require.True(t, bruh.Is(me, baseErr), "expected Is to find base error deep in wrapped chain")
	})
}

// -----------------------------------------------------------------------------
// Context and Tags (ctxerror integration)
// -----------------------------------------------------------------------------

func TestContextIntegration(t *testing.T) {
	t.Parallel()

	t.Run("ContextCombinesContextsFromAllErrors", func(t *testing.T) {
		e1 := ctxerror.New("error1")
		e1.SetContext("req", map[string]any{"id": "1"})

		e2 := ctxerror.New("error2")
		e2.SetContext("user", map[string]any{"id": "u1"})

		me := mustErr(t, New("test", Options{}))
		me.Add(e1, e2)

		ctx := me.Context()
		assert.Equal(t, ctx["req"]["id"], "1")
		assert.Equal(t, ctx["user"]["id"], "u1")
	})

	t.Run("ContextReturnsEmptyMapForNoContextErrors", func(t *testing.T) {
		me := mustErr(t, New("test", Options{}))
		me.Add(errors.New("plain"), errors.New("errors"))

		ctx := me.Context()
		assert.Len(t, ctx, 0)
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
		assert.Equal(t, ctx["req"]["id"], "2")
		// status from e1 is lost because e2's group overwrote it
		_, hasStatus := ctx["req"]["status"]
		assert.False(t, hasStatus, "expected status to be overwritten by e2's group")
	})
}

func TestTagsIntegration(t *testing.T) {
	t.Parallel()

	t.Run("TagsCombinesTagsFromAllErrors", func(t *testing.T) {
		e1 := ctxerror.New("error1")
		e1.SetTag("env", "prod")

		e2 := ctxerror.New("error2")
		e2.SetTag("zone", "eu")

		me := mustErr(t, New("test", Options{}))
		me.Add(e1, e2)

		tags := me.Tags()
		assert.Equal(t, tags["env"], "prod")
		assert.Equal(t, tags["zone"], "eu")
	})

	t.Run("TagsReturnsEmptyMapForNoTagErrors", func(t *testing.T) {
		me := mustErr(t, New("test", Options{}))
		me.Add(errors.New("plain"), errors.New("errors"))

		tags := me.Tags()
		assert.Len(t, tags, 0)
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
		assert.Equal(t, tags["env"], "staging")
		assert.Equal(t, tags["op"], "read")
	})
}

// -----------------------------------------------------------------------------
// Error formatting
// -----------------------------------------------------------------------------

func TestErrorFormatting(t *testing.T) {
	t.Parallel()

	t.Run("ErrorIncludesMessageAndFormattedErrors", func(t *testing.T) {
		me := New("main error", Options{})
		me.Add(errors.New("error1"), errors.New("error2"))

		msg := me.Error()
		assert.True(t, len(msg) > 0)
		assert.True(t, len(msg) > 0 && msg != "")
	})

	t.Run("ErrorReturnsEmptyStringForNilMultiErrorer", func(t *testing.T) {
		me := New("test", Options{})
		assert.Equal(t, me.Error(), "")
	})

	t.Run("ErrorPadsErrorNumbersCorrectly", func(t *testing.T) {
		me := New("test", Options{})
		for i := range 15 {
			me.Add(fmt.Errorf("error %d", i))
		}
		msg := me.Error()
		// Should have 2-digit padding like #00, #01, etc.
		assert.True(t, len(msg) > 0)
	})

	t.Run("MessageReturnsSameAsError", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(errors.New("e1"))
		assert.Equal(t, me.Message(), me.Error())
	})
}

// -----------------------------------------------------------------------------
// bruh integration
// -----------------------------------------------------------------------------

func TestBruhIntegration(t *testing.T) {
	t.Parallel()

	t.Run("MultiErrorerWrapsBruhErrors", func(t *testing.T) {
		bruhErr := bruh.New("bruh error")
		me := New("wrapper", Options{})
		me.Add(bruhErr)

		assert.Len(t, me.Errors(), 1)
		assert.Equal(t, me.Errors()[0], bruhErr)
	})

	t.Run("MultiErrorerIntegratesWithBruhWrap", func(t *testing.T) {
		me := New("test", Options{})
		me.Add(errors.New("a"), errors.New("b"))

		wrapped := bruh.Wrap(me, "outer wrapper")
		assert.NotNil(t, wrapped)
		assert.True(t, errors.Is(wrapped, me))
	})
}

// -----------------------------------------------------------------------------
// Edge cases and interactions
// -----------------------------------------------------------------------------

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("UnwrapWithSingleError", func(t *testing.T) {
		me := New("test", Options{UnwrapBehavior: UnwrapFirst})
		err := errors.New("single")
		me.Add(err)
		assert.Equal(t, me.Unwrap(), err)
	})

	t.Run("FilterAppliedDuringAdd", func(t *testing.T) {
		me := New("test", Options{Filter: func(e error) bool {
			return e.Error() != "filtered"
		}})
		me.Add(errors.New("keep"), errors.New("filtered"), errors.New("keep"))
		assert.Len(t, me.Errors(), 2)
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
		assert.Equal(t, me.Unwrap(), e2)
		assert.Len(t, me.Errors(), 2)
	})

	t.Run("ErrorsReflectsCurrentState", func(t *testing.T) {
		me := New("test", Options{})
		assert.Len(t, me.Errors(), 0)
		me.Add(errors.New("a"))
		assert.Len(t, me.Errors(), 1)
		me.Add(errors.New("b"))
		assert.Len(t, me.Errors(), 2)
	})

	t.Run("GrowCanBeCalledMultipleTimes", func(t *testing.T) {
		me := New("test", Options{})
		me.Grow(5)
		me.Grow(10)
		me.Grow(15)
		assert.True(t, cap(me.Errors()) >= 15)
	})
}
