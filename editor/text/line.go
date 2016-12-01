package text

import (
	"unicode/utf8"

	"golang.org/x/image/math/fixed"
)

type Line struct {
	// Adv can be used by the client to store pixel advances when
	// drawing the line.
	Adv []fixed.Int26_6

	s []byte
}

func newLineFromString(s string) *Line {
	return &Line{s: []byte(s)}
}

//String returns the contents of l as a string.
func (l *Line) String() string { return string(l.s) }

// Bytes returns the contents of l as a byte slice. The caller should
// take care not to modify the slice.
func (l *Line) Bytes() []byte { return l.s }

func (l *Line) RuneCount() int { return utf8.RuneCount(l.s) }

func (l *Line) elemFromCol(col int) (elem int) {
	for i := 0; i < col; i++ {
		_, n := utf8.DecodeRune(l.s[elem:])
		elem += n
	}
	return
}

// insertString inserts s into l at column col, and returns the new
// column (i.e. col + the number of columns inserted)
func (l *Line) insertString(col int, s string) int {
	elem := l.elemFromCol(col)
	l.s = append(l.s, s...) // grow l by len(s)
	copy(l.s[elem+len(s):], l.s[elem:])

	// insert s
	for i := 0; i < len(s); i++ {
		l.s[elem+i] = s[i]
	}

	return col + utf8.RuneCountInString(s)
}
