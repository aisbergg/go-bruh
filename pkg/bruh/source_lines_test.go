package bruh

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestGetSourceLinesWithFS(t *testing.T) {
	t.Parallel()

	assertSourceLines := func(name string, fsys fstest.MapFS, file string, lines []int, ctxLines, colCap int, unindent bool, exp []SourceLines) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			act, err := getSourceLines(fsys, file, lines, ctxLines, colCap, unindent)
			if err != nil {
				t.Fatalf("getSourceLines() error = %v", err)
			}
			if len(act) != len(exp) {
				t.Fatalf("len(getSourceLines()) = %d, want %d", len(act), len(exp))
			}
			for i := range exp {
				if len(act[i]) != len(exp[i]) {
					t.Fatalf("len(getSourceLines()[%d]) = %d, want %d", i, len(act[i]), len(exp[i]))
				}
				for j := range exp[i] {
					if act[i][j] != exp[i][j] {
						t.Fatalf("getSourceLines()[%d][%d] = %#v, want %#v", i, j, act[i][j], exp[i][j])
					}
				}
			}
		})
	}

	assertSourceLines(
		"ReadsFromInjectedFS",
		fstest.MapFS{
			"testdata/sample.go": {
				Data: []byte("package main\n\nfunc main() {\n\tprintln(\"hi\")\n}\n"),
			},
		},
		"testdata/sample.go",
		[]int{3},
		1,
		0,
		false,
		[]SourceLines{{
			{LineNum: 2, Source: ""},
			{LineNum: 3, Source: "func main() {"},
			{LineNum: 4, Source: "\tprintln(\"hi\")"},
		}},
	)

	assertSourceLines(
		"SkipsNegativeContextAndPadsMissingTrailingContext",
		fstest.MapFS{
			"testdata/sample.go": {
				Data: []byte("line 1\nline 2\n"),
			},
		},
		"testdata/sample.go",
		[]int{1, 2},
		2,
		0,
		false,
		[]SourceLines{
			{
				{},
				{},
				{LineNum: 1, Source: "line 1"},
				{LineNum: 2, Source: "line 2"},
				{LineNum: 3},
			},
			{
				{},
				{LineNum: 1, Source: "line 1"},
				{LineNum: 2, Source: "line 2"},
				{LineNum: 3},
				{LineNum: 4},
			},
		},
	)

	assertSourceLines(
		"UnindentsAndTrimsColumns",
		fstest.MapFS{
			"testdata/indented.go": {
				Data: []byte("\t\talpha\n\t\tbeta\n"),
			},
		},
		"testdata/indented.go",
		[]int{1, 2},
		-1,
		3,
		true,
		[]SourceLines{
			{{LineNum: 1, Source: "alp"}},
			{{LineNum: 2, Source: "bet"}},
		},
	)

	t.Run("HandlesEmptyFileWhenNoLinesRequested", func(t *testing.T) {
		act, err := getSourceLines(
			fstest.MapFS{"testdata/empty.go": {Data: nil}},
			"testdata/empty.go",
			nil,
			0,
			0,
			false,
		)
		if err != nil {
			t.Fatalf("getSourceLines() error = %v", err)
		}
		if len(act) != 0 {
			t.Fatalf("len(getSourceLines()) = %d, want 0", len(act))
		}
	})

	t.Run("ReadsAbsolutePathWithOSFS", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "sample.go")
		if err := os.WriteFile(file, []byte("first\nsecond\n"), 0o600); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		act, err := getSourceLines(osFS{}, file, []int{2}, 0, 0, false)
		if err != nil {
			t.Fatalf("getSourceLines() error = %v", err)
		}
		exp := []SourceLines{{{LineNum: 2, Source: "second"}}}
		if len(act) != len(exp) {
			t.Fatalf("len(getSourceLines()) = %d, want %d", len(act), len(exp))
		}
		if len(act[0]) != len(exp[0]) {
			t.Fatalf("len(getSourceLines()[0]) = %d, want %d", len(act[0]), len(exp[0]))
		}
		if act[0][0] != exp[0][0] {
			t.Fatalf("getSourceLines()[0][0] = %#v, want %#v", act[0][0], exp[0][0])
		}
	})
}

func TestGetSourceLinesErrors(t *testing.T) {
	t.Parallel()

	assertError := func(name string, fsys osFS, file string, lines []int, exp string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			_, err := getSourceLines(fsys, file, lines, 0, 0, false)
			if err == nil {
				t.Fatalf("getSourceLines() error = nil, want %q", exp)
			}
			if err.Error() != exp {
				t.Fatalf("getSourceLines() error = %q, want %q", err.Error(), exp)
			}
		})
	}

	assertError("RejectsNonGoFiles", osFS{}, "testdata/sample.txt", []int{1}, "source file must have a .go extension")

	t.Run("FailsWhenRequestedLineIsBeyondEOF", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "sample.go")
		if err := os.WriteFile(file, []byte("line 1\n"), 0o600); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}

		_, err := getSourceLines(osFS{}, file, []int{2}, 0, 0, false)
		if err == nil {
			t.Fatal("getSourceLines() error = nil, want source file too short")
		}
		if err.Error() != "source file too short" {
			t.Fatalf("getSourceLines() error = %q, want %q", err.Error(), "source file too short")
		}
	})
}
