package text

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Buffer struct {
	Lines []*Line
}

type Line struct {
	s []byte
}

func (l *Line) String() string {
	return string(l.s)
}

func (l *Line) RuneCount() int {
	return utf8.RuneCount(l.s)
}

func NewBuffer() *Buffer {
	return &Buffer{
		Lines: []*Line{new(Line)},
	}
}

func (b *Buffer) NextAddress(a Address) Address {
	if a.Col < len(b.Lines[a.Row].s) {
		a.Col++
	} else if a.Row < len(b.Lines)-1 {
		a.Col = 0
		a.Row++
	}
	return a
}

func (b *Buffer) PrevAddress(a Address) Address {
	if a.Col > 0 {
		a.Col--
	} else if a.Row > 0 {
		a.Row--
		a.Col = len(b.Lines[a.Row].s)
	}
	return a
}

func (b *Buffer) Contents() []byte {
	var buf bytes.Buffer
	for _, l := range b.Lines {
		buf.Write(l.s)
		buf.WriteByte('\n')
	}
	// trim the extra newline
	return buf.Bytes()[:buf.Len()-1]
}

func (b *Buffer) GetSel(sel Selection) string {
	if sel.IsEmpty() {
		return ""
	}

	if sel.From.Row == sel.To.Row {
		row := sel.From.Row
		from := byteCount(b.Lines[row].s, sel.From.Col)
		to := byteCount(b.Lines[row].s, sel.To.Col)
		return string(b.Lines[row].s[from:to])
	}

	from := byteCount(b.Lines[sel.From.Row].s, sel.From.Col)
	ret := string(b.Lines[sel.From.Row].s[from:]) + "\n"
	for i := sel.From.Row + 1; i < sel.To.Row; i++ {
		ret += string(b.Lines[i].s) + "\n"
	}
	to := byteCount(b.Lines[sel.To.Row].s, sel.To.Col)
	ret += string(b.Lines[sel.To.Row].s[:to])
	return ret
}

func (b *Buffer) ClearSel(sel Selection) Selection {
	if sel.IsEmpty() {
		return sel
	}

	row1, row2 := sel.From.Row, sel.To.Row
	colBytes1 := byteCount(b.Lines[row1].s, sel.From.Col)
	colBytes2 := byteCount(b.Lines[row2].s, sel.To.Col)

	// make a new line from trimmed row1 and row2
	line := b.Lines[row1].s[:colBytes1]
	b.Lines[row1].s = append(line, b.Lines[row2].s[colBytes2:]...)

	if row2 > row1 {
		// delete remaining Lines
		b.Lines = append(b.Lines[:row1+1], b.Lines[row2+1:]...)
	}
	return Selection{sel.From, sel.From}
}

// InsertString inserts s into the buffer at a, adding new Lines if s
// contains newline characters.
func (b *Buffer) InsertString(a Address, s string) Address {
	inputLines := strings.Split(s, "\n")
	if len(inputLines) == 1 {
		// fast path for inserts with no newline
		return b.insertStringSingle(a, s)
	}

	// grow b.Lines as necessary
	for i := 0; i < len(inputLines)-1; i++ {
		b.Lines = append(b.Lines, &Line{})
	}
	copy(b.Lines[a.Row+len(inputLines)-1:], b.Lines[a.Row:])

	// add all completely new lines
	for i := 1; i < len(inputLines)-1; i++ {
		b.Lines[a.Row+i] = &Line{
			s: []byte(inputLines[i]),
		}
	}

	// last line is new, but is constructed from the last input line
	// and part of the original row
	part1 := inputLines[len(inputLines)-1]
	part2 := ""
	split := byteCount(b.Lines[a.Row].s, a.Col)
	if split < len(b.Lines[a.Row].s) {
		part2 = string(b.Lines[a.Row].s[split:])
	}
	b.Lines[a.Row+len(inputLines)-1] = &Line{
		s: []byte(part1 + part2),
	}

	// finally, append the first line of input to the remainder
	// of the original line
	b.Lines[a.Row].s = append(b.Lines[a.Row].s[:split], inputLines[0]...)

	a.Row += len(inputLines) - 1
	a.Col = utf8.RuneCountInString(inputLines[len(inputLines)-1])
	return a
}

// InsertBytes
func (b *Buffer) InsertBytes(a Address, s []byte) Address {
	inputLines := bytes.Split(s, []byte("\n"))
	if len(inputLines) == 1 {
		// fast path for inserts with no newline
		return b.insertBytesSingle(a, s)
	}

	// grow b.Lines as necessary
	for i := 0; i < len(inputLines)-1; i++ {
		b.Lines = append(b.Lines, &Line{})
	}
	copy(b.Lines[a.Row+len(inputLines)-1:], b.Lines[a.Row:])

	// add all completely new lines
	for i := 1; i < len(inputLines)-1; i++ {
		b.Lines[a.Row+i] = &Line{
			s: inputLines[i],
		}
	}

	// last line is new, but is constructed from the last input line
	// and part of the original row
	part1 := inputLines[len(inputLines)-1]
	part2 := []byte{}
	split := byteCount(b.Lines[a.Row].s, a.Col)
	if split < len(b.Lines[a.Row].s) {
		part2 = b.Lines[a.Row].s[split:]
	}
	b.Lines[a.Row+len(inputLines)-1] = &Line{
		s: append(part1, part2...),
	}

	// finally, append the first line of input to the remainder
	// of the original line
	b.Lines[a.Row].s = append(b.Lines[a.Row].s[:split], inputLines[0]...)

	a.Row += len(inputLines) - 1
	a.Col = utf8.RuneCount(inputLines[len(inputLines)-1])
	return a
}

func (b *Buffer) insertStringSingle(a Address, s string) Address {
	l := b.Lines[a.Row].s

	c := byteCount(l, a.Col)  // split the line at c
	l = append(l, s...)       // grow l by len(s)
	copy(l[c+len(s):], l[c:]) // shift second part over

	// insert s
	for i := 0; i < len(s); i++ {
		l[c+i] = s[i]
	}
	b.Lines[a.Row].s = l

	a.Col += utf8.RuneCountInString(s)
	return a
}

func (b *Buffer) insertBytesSingle(a Address, s []byte) Address {
	l := b.Lines[a.Row].s

	c := byteCount(l, a.Col)  // split the line at c
	l = append(l, s...)       // grow l by len(s)
	copy(l[c+len(s):], l[c:]) // shift second part over

	// insert s
	for i := 0; i < len(s); i++ {
		l[c+i] = s[i]
	}
	b.Lines[a.Row].s = l

	a.Col += utf8.RuneCount(s)
	return a
}

func byteCount(s []byte, col int) (count int) {
	for i := 0; i < col; i++ {
		_, n := utf8.DecodeRune(s[count:])
		count += n
	}
	return
}

// AutoSelect selects some text around a. Based on acme's double click selection rules.
func (b *Buffer) AutoSelect(addr Address) Selection {
	// select bracketed text
	if sel, ok := b.selDelimited(addr, "{[(<", "}])>"); ok {
		return sel
	}

	sel := Selection{addr, addr}
	line := []rune(string(b.Lines[addr.Row].s)) // TODO: try to avoid this

	// select line
	if addr.Col == len(line) || addr.Col == 0 {
		sel.From.Col = 0
		if addr.Row+1 < len(b.Lines) {
			sel.To.Row++
			sel.To.Col = 0
		} else {
			sel.To.Col = len(line)
		}
		return sel
	}

	// select quoted text
	const quotes = "\"'`"
	if sel, ok := b.selDelimited(addr, quotes, quotes); ok {
		return sel
	}

	// Select a word. If we're on a non-alphanumeric, attempt to select a word to
	// the left of the click; otherwise expand across alphanumerics in both directions.
	for col := addr.Col; col > 0 && isAlnum(line[col-1]); col-- {
		sel.From.Col--
	}
	if isAlnum(line[addr.Col]) {
		for col := addr.Col; col < len(line) && isAlnum(line[col]); col++ {
			sel.To.Col++
		}
	}
	return sel
}

func isAlnum(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c)
}

// returns true if a selection was attempted, successfully or not
func (b *Buffer) selDelimited(addr Address, delims1, delims2 string) (Selection, bool) {
	sel := Selection{addr, addr}

	var delim int
	var line = []rune(string(b.Lines[addr.Row].s)) // TODO: try to avoid this
	var next func(Address) Address
	var rightwards bool
	if addr.Col > 0 {
		if delim = strings.IndexRune(delims1, line[addr.Col-1]); delim != -1 {
			// scan to the right, from a left delimiter
			next = b.NextAddress
			rightwards = true

			// the user double-clicked to the right of a left delimiter; move addr
			// to the delimiter itself
			addr.Col--
		}
	}
	if next == nil && addr.Col < len(line) {
		if delim = strings.IndexRune(delims2, line[addr.Col]); delim != -1 {
			// scan to the left, from a right delimiter
			// swap delimiters so that delim1 refers to the first one we encountered
			delims1, delims2 = delims2, delims1
			next = b.PrevAddress
		}
	}
	if next == nil {
		return sel, false
	}

	stack := 0
	match := addr
	prev := Address{-1, -1}
	for match != prev {
		prev = match
		match = next(match)
		line := []rune(string(b.Lines[match.Row].s)) // TODO: avoid
		if match.Col > len(line)-1 {
			continue
		}
		c := line[match.Col]
		if c == rune(delims2[delim]) && stack == 0 {
			if rightwards {
				sel.From, sel.To = addr, match
			} else {
				sel.From, sel.To = match, addr
			}
			sel.From.Col++ // move the head of the selection past the left delimiter
			return sel, true
		} else if c == 0 {
			return sel, true
		}
		if delims1 != delims2 && c == rune(delims1[delim]) {
			stack++
		}
		if delims1 != delims2 && c == rune(delims2[delim]) {
			stack--
		}
	}
	return sel, true
}
