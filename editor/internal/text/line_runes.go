package text

type Line struct {
	s []rune

	// Adv can be used by the client to store pixel advances when
	// drawing the line.
	Adv []int16
}

func newLineFromString(s string) *Line {
	return &Line{s: []rune(s)}
}

func (l *Line) String() string {
	return string(l.s)
}

func (l *Line) bytes() []byte {
	return []byte(string(l.s))
}

func (l *Line) runes() []rune {
	return l.s
}

func (l *Line) RuneCount() int {
	return len(l.s)
}

func (l *Line) elemFromCol(col int) (elem int) {
	return col
}

// insert inserts s into l at column col, and returns the new
// column (i.e. col + the number of columns inserted)
func (l *Line) insertString(col int, s string) int {
	runes := []rune(s)
	l.s = append(l.s, runes...) // grow l by len(s)
	copy(l.s[col+len(runes):], l.s[col:])

	// insert s
	for i := 0; i < len(runes); i++ {
		l.s[col+i] = runes[i]
	}

	return col + len(runes)
}
