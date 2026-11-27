package require

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
)

// fakeT implements the subset of testing.TB used by this package.
type fakeT struct {
	testing.TB
	errored bool
	failNow int
	msgs    []string
}

func (f *fakeT) Helper() {}

func (f *fakeT) Errorf(format string, args ...any) {
	f.errored = true
	f.msgs = append(f.msgs, fmt.Sprintf(format, args...))
}

func (f *fakeT) FailNow() { f.failNow++ }

// passOK asserts the assertion did not fail (no Errorf, no FailNow).
func passOK(t *testing.T, label string, f *fakeT) {
	t.Helper()
	if f.errored || f.failNow != 0 {
		t.Fatalf("%s: expected pass, got errored=%v failNow=%d msgs=%v", label, f.errored, f.failNow, f.msgs)
	}
}

// failStop asserts the assertion failed and stopped (Errorf + FailNow).
func failStop(t *testing.T, label string, f *fakeT, substr string) {
	t.Helper()
	if !f.errored || f.failNow == 0 {
		t.Fatalf(
			"%s: expected Errorf and FailNow, got errored=%v failNow=%d msgs=%v",
			label,
			f.errored,
			f.failNow,
			f.msgs,
		)
	}
	if substr != "" && (len(f.msgs) == 0 || !strings.Contains(f.msgs[0], substr)) {
		t.Fatalf("%s: message %q missing %q", label, f.msgs, substr)
	}
}

func TestEqual(t *testing.T) {
	assertOK := func(name string, got, want any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			Equal(f, got, want)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, got, want any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			Equal(f, got, want)
			failStop(t, name, f, substr)
		})
	}
	assertOK("Ints", 1, 1)
	assertOK("NilSlices", []int(nil), []int(nil))
	assertOK("ByteSlices", []byte("ab"), []byte("ab"))
	assertFail("Mismatch", 1, 2, "got:  1")
}

func TestNotEqual(t *testing.T) {
	t.Run("Different", func(t *testing.T) {
		f := &fakeT{}
		NotEqual(f, 1, 2)
		passOK(t, "Different", f)
	})
	t.Run("Same", func(t *testing.T) {
		f := &fakeT{}
		NotEqual(f, 1, 1)
		failStop(t, "Same", f, "got equal to want:")
	})
}

func TestEqualUnordered(t *testing.T) {
	assertOK := func(name string, got, want []int) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			EqualUnordered(f, got, want)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, got, want []int, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			EqualUnordered(f, got, want)
			failStop(t, name, f, substr)
		})
	}
	assertOK("OrderDiffers", []int{1, 2, 3}, []int{3, 1, 2})
	assertOK("Duplicates", []int{1, 1, 2}, []int{1, 2, 1})
	assertFail("Extra", []int{1, 2, 3}, []int{1, 2}, "extra:")
	assertFail("Missing", []int{1, 2}, []int{1, 2, 3}, "missing:")
}

func TestError(t *testing.T) {
	t.Run("NonNil", func(t *testing.T) {
		f := &fakeT{}
		Error(f, errors.New("boom"))
		passOK(t, "NonNil", f)
	})
	t.Run("Nil", func(t *testing.T) {
		f := &fakeT{}
		Error(f, nil)
		failStop(t, "Nil", f, "expected an error")
	})
}

func TestNoError(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		f := &fakeT{}
		NoError(f, nil)
		passOK(t, "Nil", f)
	})
	t.Run("NonNil", func(t *testing.T) {
		f := &fakeT{}
		NoError(f, errors.New("boom"))
		failStop(t, "NonNil", f, "expected no error")
	})
}

func TestEqualError(t *testing.T) {
	sentinel := errors.New("base")

	assertOK := func(name string, gotErr error, wantErr any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			EqualError(f, gotErr, wantErr)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, gotErr error, wantErr any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			EqualError(f, gotErr, wantErr)
			failStop(t, name, f, substr)
		})
	}
	assertOK("NilWantNil", nil, nil)
	assertOK("MatchSame", sentinel, sentinel)
	assertOK("StringMatch", errors.New("hello"), "hello")
	assertFail("Mismatch", errors.New("other"), sentinel, "got error:")
	assertFail("Unsupported", errors.New("x"), 42, "unsupported wantErr type")
}

func TestTrue(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		True(f, true)
		passOK(t, "Pass", f)
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		True(f, false)
		failStop(t, "Fail", f, "expected true")
	})
}

func TestFalse(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		False(f, false)
		passOK(t, "Pass", f)
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		False(f, true)
		failStop(t, "Fail", f, "expected false")
	})
}

func TestNil(t *testing.T) {
	assertOK := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			Nil(f, obj)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			Nil(f, obj)
			failStop(t, name, f, "expected nil")
		})
	}
	assertOK("UntypedNil", nil)
	assertOK("TypedSliceNil", []int(nil))
	assertOK("TypedMapNil", map[string]int(nil))
	assertFail("Slice", []int{1})
}

func TestNotNil(t *testing.T) {
	assertOK := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			NotNil(f, obj)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			NotNil(f, obj)
			failStop(t, name, f, "expected not nil")
		})
	}
	assertOK("Slice", []int{1})
	assertOK("Number", 0)
	assertFail("UntypedNil", nil)
	assertFail("TypedPointerNil", (*int)(nil))
}

func TestLen(t *testing.T) {
	assertOK := func(name string, obj any, length int) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			Len(f, obj, length)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, obj any, length int, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			Len(f, obj, length)
			failStop(t, name, f, substr)
		})
	}
	assertOK("Slice", []int{1, 2, 3}, 3)
	assertOK("String", "abc", 3)
	assertFail("Mismatch", []int{1, 2}, 3, "got length: 2")
	assertFail("UnsupportedKind", 42, 0, "expected length of array")
}

func TestPanic(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		Panic(f, func() { panic("boom") })
		passOK(t, "Pass", f)
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		Panic(f, func() {})
		failStop(t, "Fail", f, "expected a panic")
	})
}

func TestNotPanic(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		NotPanic(f, func() {})
		passOK(t, "Pass", f)
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		NotPanic(f, func() { panic("boom") })
		failStop(t, "Fail", f, "")
	})
}

func TestContains(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		f := &fakeT{}
		Contains(f, []int{1, 2, 3}, 2)
		passOK(t, "Found", f)
	})
	t.Run("Missing", func(t *testing.T) {
		f := &fakeT{}
		Contains(f, []int{1, 2, 3}, 4)
		failStop(t, "Missing", f, "contain 4")
	})
}

func TestNotContains(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		f := &fakeT{}
		NotContains(f, []int{1, 2, 3}, 4)
		passOK(t, "Missing", f)
	})
	t.Run("Found", func(t *testing.T) {
		f := &fakeT{}
		NotContains(f, []int{1, 2, 3}, 2)
		failStop(t, "Found", f, "not to contain 2")
	})
}

func TestEmpty(t *testing.T) {
	assertOK := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			Empty(f, obj)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, obj any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			Empty(f, obj)
			failStop(t, name, f, substr)
		})
	}
	assertOK("EmptySlice", []int{})
	assertOK("EmptyString", "")
	assertOK("NilSlice", []int(nil))
	assertFail("NonEmpty", []int{1}, "expected empty")
	assertFail("Unsupported", 1, "expected empty")
}

func TestNotEmpty(t *testing.T) {
	assertOK := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			NotEmpty(f, obj)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, obj any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			NotEmpty(f, obj)
			failStop(t, name, f, substr)
		})
	}
	assertOK("NonEmptySlice", []int{1})
	assertOK("NonEmptyString", "x")
	assertFail("EmptySlice", []int{}, "expected non-empty")
}

func TestErrorContains(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		f := &fakeT{}
		ErrorContains(f, errors.New("hello world"), "world")
		passOK(t, "Found", f)
	})
	t.Run("Missing", func(t *testing.T) {
		f := &fakeT{}
		ErrorContains(f, errors.New("hello"), "missing")
		failStop(t, "Missing", f, "does not contain")
	})
	t.Run("NilErr", func(t *testing.T) {
		f := &fakeT{}
		ErrorContains(f, nil, "anything")
		failStop(t, "NilErr", f, "expected an error")
	})
}

func TestGreaterOrEqual(t *testing.T) {
	assertOK := func(name string, got, want any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			GreaterOrEqual(f, got, want)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, got, want any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			GreaterOrEqual(f, got, want)
			failStop(t, name, f, substr)
		})
	}
	assertOK("IntGE", 5, 3)
	assertOK("IntEqual", 3, 3)
	assertOK("StringGE", "b", "a")
	assertFail("IntLess", 2, 3, "expected 2 >= 3")
	assertFail("UnsupportedFloat", float64(1), float64(2), "unsupported type")
}

func TestLessOrEqual(t *testing.T) {
	assertOK := func(name string, got, want any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			LessOrEqual(f, got, want)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, got, want any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			LessOrEqual(f, got, want)
			failStop(t, name, f, substr)
		})
	}
	assertOK("IntLE", 3, 5)
	assertOK("IntEqual", 3, 3)
	assertOK("StringLE", "a", "b")
	assertFail("IntGreater", 5, 2, "expected 5 <= 2")
	assertFail("UnsupportedFloat", float64(1), float64(2), "unsupported type")
}

func TestZero(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		Zero(f, 0)
		passOK(t, "Pass", f)
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		Zero(f, 1)
		failStop(t, "Fail", f, "got:  1")
	})
}

func TestNotZero(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		NotZero(f, 1)
		passOK(t, "Pass", f)
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		NotZero(f, 0)
		failStop(t, "Fail", f, "equal to want")
	})
}

func TestErrorIs(t *testing.T) {
	sentinel := errors.New("base")

	assertOK := func(name string, err, target error) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			ErrorIs(f, err, target)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, err, target error, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			ErrorIs(f, err, target)
			failStop(t, name, f, substr)
		})
	}
	assertOK("Match", sentinel, sentinel)
	assertOK("BothNil", nil, nil)
	assertFail("Mismatch", errors.New("other"), sentinel, "does not match")
}

func TestContainsString(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		f := &fakeT{}
		ContainsString(f, "hello world", "world")
		passOK(t, "Found", f)
	})
	t.Run("Missing", func(t *testing.T) {
		f := &fakeT{}
		ContainsString(f, "hello", "xyz")
		failStop(t, "Missing", f, "does not contain")
	})
}

func TestNotContainsString(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		f := &fakeT{}
		NotContainsString(f, "hello", "xyz")
		passOK(t, "Missing", f)
	})
	t.Run("Found", func(t *testing.T) {
		f := &fakeT{}
		NotContainsString(f, "hello world", "world")
		failStop(t, "Found", f, "should not contain")
	})
}

func TestNotPanics(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		NotPanics(f, func() {})
		passOK(t, "Pass", f)
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		NotPanics(f, func() { panic("boom") })
		failStop(t, "Fail", f, "unexpected panic")
	})
}

type customErr struct{ Code int }

func (e *customErr) Error() string { return "code " + strconv.Itoa(e.Code) }

func TestErrorAs(t *testing.T) {
	wrapped := fmt.Errorf("ctx: %w", &customErr{Code: 7})

	t.Run("Match", func(t *testing.T) {
		f := &fakeT{}
		var target *customErr
		ErrorAs(f, wrapped, &target)
		passOK(t, "Match", f)
	})
	t.Run("BothNil", func(t *testing.T) {
		f := &fakeT{}
		ErrorAs(f, nil, nil)
		passOK(t, "BothNil", f)
	})
	t.Run("Mismatch", func(t *testing.T) {
		f := &fakeT{}
		var target *customErr
		ErrorAs(f, errors.New("other"), &target)
		failStop(t, "Mismatch", f, "does not match type")
	})
}

func TestJSONEq(t *testing.T) {
	assertOK := func(name, expected, actual string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			JSONEq(f, expected, actual)
			passOK(t, name, f)
		})
	}
	assertFail := func(name, expected, actual, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			JSONEq(f, expected, actual)
			failStop(t, name, f, substr)
		})
	}
	assertOK("ObjectOrder", `{"a":1,"b":2}`, `{"b":2,"a":1}`)
	assertOK("ArrayOrder", `[1,2,3]`, `[1,2,3]`)
	assertFail("DifferentValue", `{"a":1}`, `{"a":2}`, "got:")
	assertFail("InvalidExpected", `not-json`, `{}`, "failed to unmarshal")
}

type reader interface{ Read(p []byte) (int, error) }

func TestImplements(t *testing.T) {
	assertOK := func(name string, iface, object any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			Implements(f, iface, object)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, iface, object any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			Implements(f, iface, object)
			failStop(t, name, f, "does not implement")
		})
	}
	assertOK("Pass", (*reader)(nil), strings.NewReader(""))
	assertOK("PassIOReader", (*io.Reader)(nil), strings.NewReader(""))
	assertFail("Fail", (*reader)(nil), 42)
}

func TestIsType(t *testing.T) {
	assertOK := func(name string, expectedType, object any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			IsType(f, expectedType, object)
			passOK(t, name, f)
		})
	}
	assertFail := func(name string, expectedType, object any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			IsType(f, expectedType, object)
			failStop(t, name, f, "expected type")
		})
	}
	assertOK("Same", 1, 2)
	assertFail("Different", 1, "1")
}
