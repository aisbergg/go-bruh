package assert

import (
	"errors"
	"fmt"
	"testing"
)

type point struct{ X, Y int }

func TestCheckEqual(t *testing.T) {
	assertEqual := func(name string, got, want any, ok bool) {
		t.Run(name, func(t *testing.T) {
			if msg := checkEqual(got, want); (msg == "") != ok {
				t.Fatalf("checkEqual(%v, %v) = %q, want ok=%v", got, want, msg, ok)
			}
		})
	}
	assertEqual("Ints", 1, 1, true)
	assertEqual("Strings", "a", "a", true)
	assertEqual("Mismatch", 1, "1", false)
	assertEqual("SlicesOrdered", []int{1, 2}, []int{2, 1}, false)
	assertEqual("MapsOrderIrrelevant", map[string]int{"a": 1, "b": 2}, map[string]int{"b": 2, "a": 1}, true)
	assertEqual("NilVsEmptySlice", []int(nil), []int{}, false)
	assertEqual("ByteSlices", []byte("ab"), []byte("ab"), true)
	assertEqual("Structs", point{1, 2}, point{1, 2}, true)
}

func TestCheckNotEqual(t *testing.T) {
	assertNotEqual := func(name string, got, want any, ok bool) {
		t.Run(name, func(t *testing.T) {
			if msg := checkNotEqual(got, want); (msg == "") != ok {
				t.Fatalf("checkNotEqual(%v, %v) = %q, want ok=%v", got, want, msg, ok)
			}
		})
	}
	assertNotEqual("Different", 1, 2, true)
	assertNotEqual("Same", 1, 1, false)
}

func TestCheckEqualUnordered(t *testing.T) {
	assertUnordered := func(name string, got, want []int, ok bool) {
		t.Run(name, func(t *testing.T) {
			if msg := checkEqualUnordered(got, want); (msg == "") != ok {
				t.Fatalf("checkEqualUnordered(%v, %v) = %q, want ok=%v", got, want, msg, ok)
			}
		})
	}
	assertUnordered("Reordered", []int{1, 2, 3}, []int{3, 1, 2}, true)
	assertUnordered("DuplicatesCount", []int{1, 1, 2}, []int{1, 2, 2}, false)
	assertUnordered("ExtraAndMissing", []int{1, 2}, []int{1, 3}, false)
	assertUnordered("NilEqualsEmpty", nil, []int{}, true)
	assertUnordered("DifferentLengths", []int{1, 2}, []int{1, 2, 3}, false)

	t.Run("StructElements", func(t *testing.T) {
		got := []point{{1, 2}, {3, 4}}
		want := []point{{3, 4}, {1, 2}}
		if msg := checkEqualUnordered(got, want); msg != "" {
			t.Fatalf("checkEqualUnordered(%v, %v) = %q, want ok", got, want, msg)
		}
	})
}

func TestCheckEqualError(t *testing.T) {
	base := errors.New("base")
	wrapped := fmt.Errorf("wrapped: %w", base)
	other := errors.New("other")

	assertEqualError := func(name string, gotErr, wantErr error, ok bool) {
		t.Run(name, func(t *testing.T) {
			if msg := checkEqualError(gotErr, wantErr); (msg == "") != ok {
				t.Fatalf("checkEqualError(%v, %v) = %q, want ok=%v", gotErr, wantErr, msg, ok)
			}
		})
	}
	assertEqualError("MatchingWrapped", wrapped, base, true)
	assertEqualError("MatchingExact", base, base, true)
	assertEqualError("Mismatch", base, other, false)
	assertEqualError("BothNil", nil, nil, true)
	assertEqualError("WantNilGotErr", base, nil, false)
	assertEqualError("GotNilWantErr", nil, base, false)
}

func TestCheckErrorNoError(t *testing.T) {
	assertError := func(name string, err error, ok bool) {
		t.Run(name, func(t *testing.T) {
			if msg := checkError(err); (msg == "") != ok {
				t.Fatalf("checkError(%v) = %q, want ok=%v", err, msg, ok)
			}
			if msg := checkNoError(err); (msg == "") == ok {
				t.Fatalf("checkNoError(%v) = %q, want ok=%v", err, msg, !ok)
			}
		})
	}
	assertError("Error", errors.New("x"), true)
	assertError("NoError", nil, false)
}

func TestCheckTrueFalse(t *testing.T) {
	assertBool := func(name string, value, ok bool) {
		t.Run(name, func(t *testing.T) {
			if msg := checkTrue(value); (msg == "") != ok {
				t.Fatalf("checkTrue(%v) = %q, want ok=%v", value, msg, ok)
			}
			if msg := checkFalse(value); (msg == "") == ok {
				t.Fatalf("checkFalse(%v) = %q, want ok=%v", value, msg, !ok)
			}
		})
	}
	assertBool("True", true, true)
	assertBool("FalseFails", false, false)
}

func TestCheckNilNotNil(t *testing.T) {
	var nilMap map[string]int
	var nilSlice []int
	var nilPtr *point
	var nilErr error

	assertNil := func(name string, obj any, ok bool) {
		t.Run(name, func(t *testing.T) {
			if msg := checkNil(obj); (msg == "") != ok {
				t.Fatalf("checkNil(%v) = %q, want ok=%v", obj, msg, ok)
			}
			if msg := checkNotNil(obj); (msg == "") == ok {
				t.Fatalf("checkNotNil(%v) = %q, want ok=%v", obj, msg, !ok)
			}
		})
	}
	assertNil("UntypedNil", nil, true)
	assertNil("TypedNilPointer", nilPtr, true)
	assertNil("NilMap", nilMap, true)
	assertNil("NilSlice", nilSlice, true)
	assertNil("NilError", nilErr, true)
	assertNil("Pointer", &point{}, false)
	assertNil("EmptyStruct", point{}, false)
}

func TestCheckLen(t *testing.T) {
	assertLen := func(name string, obj any, length int, ok bool) {
		t.Run(name, func(t *testing.T) {
			if msg := checkLen(obj, length); (msg == "") != ok {
				t.Fatalf("checkLen(%v, %d) = %q, want ok=%v", obj, length, msg, ok)
			}
		})
	}
	assertLen("Slice", []int{1, 2}, 2, true)
	assertLen("String", "ab", 2, true)
	assertLen("Map", map[string]int{"a": 1}, 1, true)
	assertLen("WrongLength", []int{1}, 2, false)
	assertLen("UnsupportedType", 42, 1, false)
	assertLen("Nil", nil, 0, false)
}

func TestCheckPanic(t *testing.T) {
	assertPanic := func(name string, f func(), ok bool) {
		t.Run(name, func(t *testing.T) {
			if msg := checkPanic(f); (msg == "") != ok {
				t.Fatalf("checkPanic() = %q, want ok=%v", msg, ok)
			}
			if msg := checkNotPanic(f); (msg == "") == ok {
				t.Fatalf("checkNotPanic() = %q, want ok=%v", msg, !ok)
			}
		})
	}
	assertPanic("Panics", func() { panic("boom") }, true)
	assertPanic("NoPanic", func() {}, false)
}

func TestFormatMsg(t *testing.T) {
	assertFormat := func(name, defaultMsg, want, msg string, args ...any) {
		t.Run(name, func(t *testing.T) {
			if got := formatMsg(defaultMsg, msg, args...); got != want {
				t.Fatalf("formatMsg(%q, %q, %v) = %q, want %q", defaultMsg, msg, args, got, want)
			}
		})
	}
	assertFormat("NoMsg", "default", "default", "")
	assertFormat("FormatArgs", "default", "ctx 42\ndefault", "ctx %d", 42)
	assertFormat("BareMsg", "default", "done\ndefault", "done")
	assertFormat("EmptyDefault", "", "", "ctx %d", 42)

	t.Run("BareMsgVerbatimPercent", func(t *testing.T) {
		msg := "100% done"
		if got := formatMsgArgs("default", msg); got != "100% done\ndefault" {
			t.Fatalf("got %q", got)
		}
	})
}
