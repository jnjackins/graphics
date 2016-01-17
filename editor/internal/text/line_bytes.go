// +build ignore

package text

import (
	"bytes"
	"unicode/utf8"
)

type Line struct {
	s []byte

	// Adv can be used by the client to store pixel advances when
	// drawing the line.
	Adv []int16
}

func newLineFromString(s string) *Line {
	return &Line{s: []byte(s)}
}

func (l *Line) String() string {
	return string(l.s)
}

func (l *Line) bytes() []byte {
	return l.s
}

func (l *Line) runes() []rune {
	return bytes.Runes(l.s)
}

func (l *Line) RuneCount() int {
	return utf8.RuneCount(l.s)
}

func (l *Line) elemFromCol(col int) (elem int) {
	for i := 0; i < col; i++ {
		_, n := utf8.DecodeRune(l.s[elem:])
		elem += n
	}
	return
}

// insert inserts s into l at column col, and returns the new
// column (i.e. col + the number of columns inserted)
func (l *Line) insertString(col int, s string) int {
	l.Dirty = true
	elem := l.elemFromCol(col)
	l.s = append(l.s, s...) // grow l by len(s)
	copy(l.s[elem+len(s):], l.s[elem:])

	// insert s
	for i := 0; i < len(s); i++ {
		l.s[elem+i] = s[i]
	}

	return col + utf8.RuneCountInString(s)
}
