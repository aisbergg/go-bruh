package bruh

import (
	"bufio"
	"cmp"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// -----------------------------------------------------------------------------

// SourceLines represents a collection of SourceLine elements used for
// generating snippets of source code in error messages.
type SourceLines []SourceLine

// SourceLine represents a single line of source code along with its line number.
// LineNum specifies the line number in the source file.
// Source contains the actual content of the line.
type SourceLine struct {
	LineNum int
	Source  string
}

type lineToIndex struct {
	lineNum int
	index1  int
	index2  int
	isLine  bool
}

type osFS struct{}

func (osFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

// getSourceLines reads the given lines (index starting at 1) of source
// code from a file. ctxLines are the number of lines before and after the
// requested lines that should be included. colCap is the maximum number of
// characters per line. If unindent is true, the source lines are unindented.
func getSourceLines(
	fsys fs.FS,
	file string,
	lines []int,
	ctxLines, colCap int,
	unindent bool,
) ([]SourceLines, error) {
	ctxLines = max(0, ctxLines)
	colCap = max(0, colCap)

	// initialize the source lines data structure
	numLines := len(lines) * (2*ctxLines + 1)
	sourceLines := make([]SourceLines, len(lines))
	for i := range lines {
		sourceLines[i] = make([]SourceLine, 2*ctxLines+1)
	}

	// create a mapping from line numbers to indices in the sourceLines data
	// structure for efficient access to the file
	linesToIndex := make([]lineToIndex, 0, numLines)
	for i, l := range lines {
		index2 := 0
		for j := ctxLines - 1; j >= 0; j-- {
			linesToIndex = append(
				linesToIndex,
				lineToIndex{lineNum: l - j - 1, index1: i, index2: index2, isLine: false},
			)
			index2++
		}
		linesToIndex = append(
			linesToIndex,
			lineToIndex{lineNum: l, index1: i, index2: index2, isLine: true},
		)
		index2++
		for j := range ctxLines {
			linesToIndex = append(
				linesToIndex,
				lineToIndex{lineNum: l + j + 1, index1: i, index2: index2, isLine: false},
			)
			index2++
		}
	}
	slices.SortFunc(linesToIndex, func(a, b lineToIndex) int {
		return cmp.Compare(a.lineNum, b.lineNum)
	})

	// ensure we only read files in the current working directory
	if filepath.IsAbs(file) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, Wrap(err, "getting current working directory")
		}
		file, err = filepath.Rel(wd, file)
		if err != nil {
			return nil, Wrap(err, "getting relative path to source file")
		}
	}
	// ensure '.go' file extension, so we reduce the risk of reading anything
	// that is not supposed to be read
	if !strings.HasSuffix(file, ".go") {
		return nil, New("source file must have a .go extension")
	}
	f, err := fsys.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// try to find the source lines in the file and store them in the sourceLines data structure
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	moreLinesInFile := true
	currentLine := 1
outer:
	for _, lti := range linesToIndex {
		if lti.lineNum < 1 {
			continue
		}
		if !moreLinesInFile {
			sourceLines[lti.index1][lti.index2] = SourceLine{LineNum: lti.lineNum}
			continue
		}
		for currentLine < lti.lineNum {
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return nil, err
				}
				if lti.isLine {
					return nil, New("source file too short")
				}
				moreLinesInFile = false
				sourceLines[lti.index1][lti.index2] = SourceLine{LineNum: lti.lineNum}
				continue outer
			}
			currentLine++
		}
		sourceLines[lti.index1][lti.index2] = SourceLine{
			LineNum: lti.lineNum,
			Source:  scanner.Text(),
		}
	}

	// unindent the source lines
	if unindent {
		for i := range sourceLines {
			// count the number of leading tabs
			minTabIndents := int(^uint(0) >> 1)
			for _, sl := range sourceLines[i] {
				// skip lines that are not in the file or are empty
				if sl.LineNum <= 0 || sl.Source == "" {
					continue
				}
				lineTabIndents := 0
				for _, c := range sl.Source {
					if c != '\t' {
						break
					}
					lineTabIndents++
				}
				if lineTabIndents < minTabIndents {
					minTabIndents = lineTabIndents
				}
			}
			// strip leading tabs
			for j, sl := range sourceLines[i] {
				// skip lines that are not in the file or are empty
				if sl.LineNum <= 0 || sl.Source == "" {
					continue
				}
				sourceLines[i][j].Source = sourceLines[i][j].Source[minTabIndents:]
			}
		}
	}

	// trim the source lines to the given column capacity
	if colCap > 0 {
		for i := range sourceLines {
			for j := range sourceLines[i] {
				if len(sourceLines[i][j].Source) > colCap {
					sourceLines[i][j].Source = sourceLines[i][j].Source[:colCap]
				}
			}
		}
	}

	return sourceLines, nil
}
