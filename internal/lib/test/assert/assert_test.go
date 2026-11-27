package assert

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

// passed asserts the assertion returned true and did not call Errorf.
func passed(t *testing.T, label string, f *fakeT, ok bool) {
	t.Helper()
	if !ok || f.errored {
		t.Fatalf("%s: expected pass, got ok=%v errored=%v msgs=%v", label, ok, f.errored, f.msgs)
	}
}

// failed asserts the assertion returned false and called Errorf with msg matching substr.
func failed(t *testing.T, label string, f *fakeT, substr string) string {
	t.Helper()
	if !f.errored || f.failNow != 0 {
		t.Fatalf(
			"%s: expected Errorf without FailNow, got errored=%v failNow=%d msgs=%v",
			label,
			f.errored,
			f.failNow,
			f.msgs,
		)
	}
	if len(f.msgs) == 0 {
		t.Fatalf("%s: expected a failure message, got none", label)
	}
	if substr != "" && !strings.Contains(f.msgs[0], substr) {
		t.Fatalf("%s: message %q missing %q", label, f.msgs[0], substr)
	}
	return f.msgs[0]
}

func TestEqual(t *testing.T) {
	assertOK := func(name string, got, want any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !Equal(f, got, want) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, got, want any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if Equal(f, got, want) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, substr)
		})
	}
	assertOK("Ints", 1, 1)
	assertOK("Strings", "a", "a")
	assertOK("Structs", struct{ X, Y int }{1, 2}, struct{ X, Y int }{1, 2})
	assertOK("SlicesNil", []int(nil), []int(nil))
	assertOK("MapsUnordered", map[string]int{"a": 1, "b": 2}, map[string]int{"b": 2, "a": 1})
	assertOK("ByteSlices", []byte("ab"), []byte("ab"))
	assertFail("Mismatch", 1, 2, "got:  1")
	assertFail("MismatchedKind", 1, "1", "got:  1")
	assertFail("NilVsEmptySlice", []int(nil), []int{}, "got:")
	assertFail("SlicesOrdered", []int{1, 2}, []int{2, 1}, "got:")

	// custom message is prepended on failure
	t.Run("CustomMsg", func(t *testing.T) {
		f := &fakeT{}
		Equal(f, 1, 2, "ctx")
		want := "ctx\ngot:  1\nwant: 2"
		if f.msgs[0] != want {
			t.Fatalf("got %q, want %q", f.msgs[0], want)
		}
	})
}

func TestNotEqual(t *testing.T) {
	assertOK := func(name string, got, want any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !NotEqual(f, got, want) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, got, want any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if NotEqual(f, got, want) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, "got equal to want:")
		})
	}
	assertOK("DifferentInts", 1, 2)
	assertOK("DifferentSlices", []int{1}, []int{2})
	assertOK("OneNil", []int(nil), []int{})
	assertFail("SameInt", 1, 1)
	assertFail("SameSlice", []int{1, 2}, []int{1, 2})
}

func TestEqualUnordered(t *testing.T) {
	assertOK := func(name string, got, want []int) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !EqualUnordered(f, got, want) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, got, want []int, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if EqualUnordered(f, got, want) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, substr)
		})
	}
	assertOK("BothNil", nil, nil)
	assertOK("BothEmpty", []int{}, []int{})
	assertOK("NilVsEmpty", nil, []int{})
	assertOK("OrderDiffers", []int{1, 2, 3}, []int{3, 1, 2})
	assertOK("Duplicates", []int{1, 1, 2}, []int{1, 2, 1})
	assertFail("Extra", []int{1, 2, 3}, []int{1, 2}, "extra:")
	assertFail("Missing", []int{1, 2}, []int{1, 2, 3}, "missing:")
	assertFail("DupMismatch", []int{1, 1}, []int{1, 2}, "missing:")
}

func TestError(t *testing.T) {
	t.Run("NonNil", func(t *testing.T) {
		f := &fakeT{}
		if !Error(f, errors.New("boom")) || f.errored {
			t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
		}
	})
	t.Run("Nil", func(t *testing.T) {
		f := &fakeT{}
		if Error(f, nil) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Nil", f, "expected an error")
	})
}

func TestNoError(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		f := &fakeT{}
		if !NoError(f, nil) || f.errored {
			t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
		}
	})
	t.Run("NonNil", func(t *testing.T) {
		f := &fakeT{}
		if NoError(f, errors.New("boom")) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "NonNil", f, "expected no error")
	})
}

func TestEqualError(t *testing.T) {
	sentinel := errors.New("base")
	wrapped := fmt.Errorf("ctx: %w", sentinel)

	assertOK := func(name string, gotErr error, wantErr any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !EqualError(f, gotErr, wantErr) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, gotErr error, wantErr any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if EqualError(f, gotErr, wantErr) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, substr)
		})
	}
	assertOK("NilWantNil", nil, nil)
	assertOK("ErrorWantMatch", sentinel, sentinel)
	assertOK("ErrorWantWrapped", wrapped, sentinel)
	assertOK("StringWantMatch", errors.New("hello"), "hello")
	assertFail("NilWantError", nil, sentinel, "got error:")
	assertFail("ErrorWantMismatch", errors.New("other"), sentinel, "got error:")
	assertFail("StringWantMismatch", errors.New("hello"), "world", "got error:")
	assertFail("UnsupportedType", errors.New("x"), 42, "unsupported wantErr type")
}

func TestTrue(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "Pass", f, True(f, true))
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		if True(f, false) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Fail", f, "expected true")
	})
}

func TestFalse(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "Pass", f, False(f, false))
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		if False(f, true) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Fail", f, "expected false")
	})
}

func TestNil(t *testing.T) {
	var p *int
	var s []int
	var m map[string]int
	var c chan int
	var fn func()

	assertOK := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !Nil(f, obj) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if Nil(f, obj) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, "expected nil")
		})
	}
	assertOK("UntypedNil", nil)
	assertOK("TypedPtrNil", p)
	assertOK("TypedSliceNil", s)
	assertOK("TypedMapNil", m)
	assertOK("TypedChanNil", c)
	assertOK("TypedFuncNil", fn)
	assertOK("InterfaceWrappedNil", any(p))
	assertFail("Int", 1)
	assertFail("Slice", []int{1})
	assertFail("PtrToValue", new(strings.Builder))
}

func TestNotNil(t *testing.T) {
	assertOK := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !NotNil(f, obj) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if NotNil(f, obj) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, "expected not nil")
		})
	}
	assertOK("Slice", []int{1})
	assertOK("Map", map[string]int{"a": 1})
	assertOK("Ptr", new(strings.Builder))
	assertOK("Number", 0)
	assertFail("UntypedNil", nil)
	assertFail("TypedSliceNil", []int(nil))
	assertFail("TypedMapNil", map[string]int(nil))
}

func TestLen(t *testing.T) {
	assertOK := func(name string, obj any, length int) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !Len(f, obj, length) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, obj any, length int, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if Len(f, obj, length) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, substr)
		})
	}
	assertOK("Slice", []int{1, 2, 3}, 3)
	assertOK("Map", map[string]int{"a": 1}, 1)
	assertOK("String", "abc", 3)
	assertOK("Array", [3]int{1, 2, 3}, 3)
	assertOK("ChanWithCap", make(chan int, 4), 0) // length is 0, only buffered items count
	assertFail("SliceMismatch", []int{1, 2}, 3, "got length: 2")
	assertFail("StringMismatch", "ab", 3, "got length: 2")
	assertFail("UnsupportedKind", 42, 0, "expected length of array")
}

func TestPanic(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		if !Panic(f, func() { panic("boom") }) || f.errored {
			t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		if Panic(f, func() {}) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Fail", f, "expected a panic")
	})
}

func TestNotPanic(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		if !NotPanic(f, func() {}) || f.errored {
			t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		if NotPanic(f, func() { panic("boom") }) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Fail", f, "boom")
	})
}

func TestContains(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "Found", f, Contains(f, []int{1, 2, 3}, 2))
	})
	t.Run("Missing", func(t *testing.T) {
		f := &fakeT{}
		if Contains(f, []int{1, 2, 3}, 4) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Missing", f, "contain 4")
	})
}

func TestNotContains(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "Missing", f, NotContains(f, []int{1, 2, 3}, 4))
	})
	t.Run("Found", func(t *testing.T) {
		f := &fakeT{}
		if NotContains(f, []int{1, 2, 3}, 2) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Found", f, "not to contain 2")
	})
}

func TestEmpty(t *testing.T) {
	assertOK := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !Empty(f, obj) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, obj any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if Empty(f, obj) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, substr)
		})
	}
	assertOK("NilSlice", []int(nil))
	assertOK("EmptySlice", []int{})
	assertOK("EmptyMap", map[string]int{})
	assertOK("EmptyString", "")
	assertOK("NilChan", (chan int)(nil))
	assertOK("UnbufferedChan", make(chan int))
	assertFail("NonEmptySlice", []int{1}, "expected empty")
	assertFail("NonEmptyMap", map[string]int{"a": 1}, "expected empty")
	assertFail("NonEmptyString", "a", "expected empty")
	assertFail("NonEmptyNumber", 1, "expected empty")
}

func TestNotEmpty(t *testing.T) {
	assertOK := func(name string, obj any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !NotEmpty(f, obj) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, obj any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if NotEmpty(f, obj) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, substr)
		})
	}
	assertOK("NonEmptySlice", []int{1})
	assertOK("NonEmptyString", "x")
	assertOK("NonEmptyNumber", 1)
	assertFail("EmptySlice", []int{}, "expected non-empty")
	assertFail("NilMap", map[string]int(nil), "expected non-empty")
}

func TestErrorContains(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "Found", f, ErrorContains(f, errors.New("hello world"), "world"))
	})
	t.Run("Missing", func(t *testing.T) {
		f := &fakeT{}
		if ErrorContains(f, errors.New("hello"), "missing") {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Missing", f, "does not contain")
	})
	t.Run("NilErr", func(t *testing.T) {
		f := &fakeT{}
		if ErrorContains(f, nil, "anything") {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "NilErr", f, "expected an error")
	})
}

func TestGreaterOrEqual(t *testing.T) {
	assertOK := func(name string, got, want any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !GreaterOrEqual(f, got, want) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, got, want any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if GreaterOrEqual(f, got, want) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, substr)
		})
	}
	assertOK("IntGE", 5, 3)
	assertOK("IntEqual", 3, 3)
	assertOK("Int8GE", int8(2), int8(1))
	assertOK("Int64GE", int64(10), int64(10))
	assertOK("UintGE", uint(7), uint(5))
	assertOK("Uint64GE", uint64(1), uint64(1))
	assertOK("StringGE", "b", "a")
	assertFail("IntLess", 2, 3, "expected 2 >= 3")
	assertFail("StringLess", "a", "b", "expected a >= b")
	assertFail("UnsupportedFloat", float64(1), float64(2), "unsupported type")
}

func TestLessOrEqual(t *testing.T) {
	assertOK := func(name string, got, want any) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !LessOrEqual(f, got, want) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, got, want any, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if LessOrEqual(f, got, want) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, substr)
		})
	}
	assertOK("IntLE", 3, 5)
	assertOK("IntEqual", 3, 3)
	assertOK("Int32LE", int32(1), int32(1))
	assertOK("UintLE", uint(3), uint(5))
	assertOK("Uint8LE", uint8(0), uint8(0))
	assertOK("StringLE", "a", "b")
	assertFail("IntGreater", 5, 2, "expected 5 <= 2")
	assertFail("StringGreater", "z", "a", "expected z <= a")
	assertFail("UnsupportedFloat", float64(1), float64(2), "unsupported type")
}

func TestZero(t *testing.T) {
	t.Run("IntZero", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "IntZero", f, Zero(f, 0))
	})
	t.Run("StringZero", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "StringZero", f, Zero(f, ""))
	})
	t.Run("SliceZero", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "SliceZero", f, Zero(f, []int(nil)))
	})
	t.Run("NonZero", func(t *testing.T) {
		f := &fakeT{}
		if Zero(f, 1) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "NonZero", f, "got:  1")
	})
}

func TestNotZero(t *testing.T) {
	t.Run("NonZeroInt", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "NonZeroInt", f, NotZero(f, 1))
	})
	t.Run("NonZeroString", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "NonZeroString", f, NotZero(f, "x"))
	})
	t.Run("Zero", func(t *testing.T) {
		f := &fakeT{}
		if NotZero(f, 0) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Zero", f, "equal to want")
	})
}

func TestErrorIs(t *testing.T) {
	sentinel := errors.New("base")
	wrapped := fmt.Errorf("ctx: %w", sentinel)

	assertOK := func(name string, err, target error) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !ErrorIs(f, err, target) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name string, err, target error) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if ErrorIs(f, err, target) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, "does not match")
		})
	}
	assertOK("Direct", sentinel, sentinel)
	assertOK("Wrapped", wrapped, sentinel)
	assertOK("BothNil", nil, nil)
	assertFail("Different", errors.New("other"), sentinel)
	assertFail("NilErr", nil, sentinel)
}

func TestContainsString(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "Found", f, ContainsString(f, "hello world", "world"))
	})
	t.Run("Missing", func(t *testing.T) {
		f := &fakeT{}
		if ContainsString(f, "hello", "xyz") {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Missing", f, "does not contain")
	})
}

func TestNotContainsString(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "Missing", f, NotContainsString(f, "hello", "xyz"))
	})
	t.Run("Found", func(t *testing.T) {
		f := &fakeT{}
		if NotContainsString(f, "hello world", "world") {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Found", f, "should not contain")
	})
}

func TestNotPanics(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		if !NotPanics(f, func() {}) || f.errored {
			t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		if NotPanics(f, func() { panic("boom") }) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Fail", f, "unexpected panic")
	})
}

type customErr struct{ Code int }

func (e *customErr) Error() string { return "code " + strconv.Itoa(e.Code) }

func TestErrorAs(t *testing.T) {
	wrapped := fmt.Errorf("ctx: %w", &customErr{Code: 7})

	t.Run("Match", func(t *testing.T) {
		f := &fakeT{}
		var target *customErr
		if !ErrorAs(f, wrapped, &target) || f.errored {
			t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
		}
		if target == nil || target.Code != 7 {
			t.Fatalf("target not populated: %+v", target)
		}
	})
	t.Run("BothNil", func(t *testing.T) {
		f := &fakeT{}
		passed(t, "BothNil", f, ErrorAs(f, nil, nil))
	})
	t.Run("Mismatch", func(t *testing.T) {
		f := &fakeT{}
		var target *customErr
		if ErrorAs(f, errors.New("other"), &target) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Mismatch", f, "does not match type")
	})
}

func TestJSONEq(t *testing.T) {
	assertOK := func(name, expected, actual string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if !JSONEq(f, expected, actual) || f.errored {
				t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
			}
		})
	}
	assertFail := func(name, expected, actual, substr string) {
		t.Run(name, func(t *testing.T) {
			f := &fakeT{}
			if JSONEq(f, expected, actual) {
				t.Fatalf("expected failure, got pass")
			}
			failed(t, name, f, substr)
		})
	}
	assertOK("ObjectOrder", `{"a":1,"b":2}`, `{"b":2,"a":1}`)
	assertOK("NestedEqual", `{"a":{"x":1}}`, `{"a":{"x":1}}`)
	assertOK("ArrayOrder", `[1,2,3]`, `[1,2,3]`)
	assertFail("DifferentValue", `{"a":1}`, `{"a":2}`, "got:")
	assertFail("DifferentKey", `{"a":1}`, `{"b":1}`, "got:")
	assertFail("InvalidExpected", `not-json`, `{}`, "failed to unmarshal")
	assertFail("InvalidActual", `{}`, `not-json`, "failed to unmarshal")
}

type reader interface{ Read(p []byte) (int, error) }

func TestImplements(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		var r io.Reader = strings.NewReader("")
		if !Implements(f, (*reader)(nil), r) || f.errored {
			t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		if Implements(f, (*reader)(nil), 42) {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Fail", f, "does not implement")
	})
}

func TestIsType(t *testing.T) {
	t.Run("Pass", func(t *testing.T) {
		f := &fakeT{}
		if !IsType(f, 1, 2) || f.errored {
			t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
		}
	})
	t.Run("PassDeepEqual", func(t *testing.T) {
		f := &fakeT{}
		if !IsType(f, []int{1}, []int{2}) || f.errored {
			t.Fatalf("expected pass, got errored=%v msgs=%v", f.errored, f.msgs)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		f := &fakeT{}
		if IsType(f, 1, "1") {
			t.Fatalf("expected failure, got pass")
		}
		failed(t, "Fail", f, "expected type")
	})
}
