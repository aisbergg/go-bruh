package require

import (
	"testing"
)

// fakeT implements the subset of testing.TB used by this package.
type fakeT struct {
	testing.TB
	errored bool
	failNow int
}

func (f *fakeT) Helper() {}

func (f *fakeT) Errorf(format string, args ...any) { f.errored = true }

func (f *fakeT) FailNow() { f.failNow++ }

func TestEqualPass(t *testing.T) {
	f := &fakeT{}
	Equal(f, 1, 1)
	if f.errored || f.failNow != 0 {
		t.Fatalf("expected pass, got errored=%v failNow=%d", f.errored, f.failNow)
	}
}

func TestEqualFail(t *testing.T) {
	f := &fakeT{}
	Equal(f, 1, 2)
	if !f.errored || f.failNow != 1 {
		t.Fatalf("expected Errorf and FailNow, got errored=%v failNow=%d", f.errored, f.failNow)
	}
}

func TestNoErrorPass(t *testing.T) {
	f := &fakeT{}
	NoError(f, nil)
	if f.errored || f.failNow != 0 {
		t.Fatalf("expected pass, got errored=%v failNow=%d", f.errored, f.failNow)
	}
}
