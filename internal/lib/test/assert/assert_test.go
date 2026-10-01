package assert

import (
	"fmt"
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

func TestEqualPass(t *testing.T) {
	f := &fakeT{}
	if !Equal(f, 1, 1) || f.errored {
		t.Fatal("expected pass without error")
	}
}

func TestEqualFail(t *testing.T) {
	f := &fakeT{}
	if Equal(f, 1, 2) {
		t.Fatal("expected failure")
	}
	if !f.errored || f.failNow != 0 {
		t.Fatalf("expected Errorf without FailNow, got errored=%v failNow=%d", f.errored, f.failNow)
	}
	if !strings.Contains(f.msgs[0], "got:  1") || !strings.Contains(f.msgs[0], "want: 2") {
		t.Fatalf("unexpected message: %q", f.msgs[0])
	}
}

func TestCustomMessage(t *testing.T) {
	f := &fakeT{}
	Equalf(f, 1, 2, "ctx %d", 42)
	want := "ctx 42\ngot:  1\nwant: 2"
	if f.msgs[0] != want {
		t.Fatalf("got %q, want %q", f.msgs[0], want)
	}
}

func TestBareMessage(t *testing.T) {
	f := &fakeT{}
	Equal(f, 1, 2, "just context")
	want := "just context\ngot:  1\nwant: 2"
	if f.msgs[0] != want {
		t.Fatalf("got %q, want %q", f.msgs[0], want)
	}
}

func TestEqualUnordered(t *testing.T) {
	f := &fakeT{}
	if !EqualUnordered(f, []int{1, 2, 3}, []int{3, 1, 2}) || f.errored {
		t.Fatal("expected pass without error")
	}
	if EqualUnordered(f, []int{1, 1, 2}, []int{1, 2, 2}) {
		t.Fatal("expected failure")
	}
}

func TestPanic(t *testing.T) {
	f := &fakeT{}
	if !Panic(f, func() { panic("boom") }) || f.errored {
		t.Fatal("expected pass without error")
	}
	if NotPanic(f, func() { panic("boom") }) || !f.errored {
		t.Fatal("expected failure")
	}
}
