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

func NewBuffer() *Buffer {
	return &Buffer{
		Lines: []*Line{new(Line)},
	}
}

func (b *Buffer) NextAddress(a Address) Address {
	if a.Col < b.Lines[a.Row].RuneCount() {
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
		a.Col = b.Lines[a.Row].RuneCount()
	}
	return a
}

func (b *Buffer) Contents() []byte {
	var buf bytes.Buffer
	for _, l := range b.Lines {
		buf.Write(l.Bytes())
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
		from := b.Lines[row].elemFromCol(sel.From.Col)
		to := b.Lines[row].elemFromCol(sel.To.Col)
		return string(b.Lines[row].s[from:to])
	}

	from := b.Lines[sel.From.Row].elemFromCol(sel.From.Col)
	ret := string(b.Lines[sel.From.Row].s[from:]) + "\n"
	for i := sel.From.Row + 1; i < sel.To.Row; i++ {
		ret += string(b.Lines[i].s) + "\n"
	}
	to := b.Lines[sel.To.Row].elemFromCol(sel.To.Col)
	ret += string(b.Lines[sel.To.Row].s[:to])
	return ret
}

func (b *Buffer) ClearSel(sel Selection) Selection {
	if sel.IsEmpty() {
		return sel
	}

	row1, row2 := sel.From.Row, sel.To.Row
	elem1 := b.Lines[row1].elemFromCol(sel.From.Col)
	elem2 := b.Lines[row2].elemFromCol(sel.To.Col)

	// make a new line from trimmed row1 and row2
	line := b.Lines[row1].s[:elem1]
	b.Lines[row1].s = append(line, b.Lines[row2].s[elem2:]...)

	if row2 > row1 {
		// delete remaining Lines
		b.Lines = append(b.Lines[:row1+1], b.Lines[row2+1:]...)
	}
	return Selection{sel.From, sel.From}
}

// InsertString inserts s into the buffer at a, adding new Lines if s
// contains newline characters.
func (b *Buffer) InsertString(addr Address, s string) Address {
	inputLines := strings.Split(s, "\n")
	if len(inputLines) == 1 {
		// fast path for inserts with no newline
		addr.Col = b.Lines[addr.Row].insertString(addr.Col, s)
		return addr
	}

	// grow b.Lines as necessary
	for i := 0; i < len(inputLines)-1; i++ {
		b.Lines = append(b.Lines, &Line{})
	}
	copy(b.Lines[addr.Row+len(inputLines)-1:], b.Lines[addr.Row:])

	// add all completely new lines
	for i := 1; i < len(inputLines)-1; i++ {
		b.Lines[addr.Row+i] = newLineFromString(inputLines[i])
	}

	// last line is new, but is constructed from the last input line
	// and part of the original row
	part1 := inputLines[len(inputLines)-1]
	part2 := ""
	split := b.Lines[addr.Row].elemFromCol(addr.Col)
	if split < len(b.Lines[addr.Row].s) {
		part2 = string(b.Lines[addr.Row].s[split:])
	}
	b.Lines[addr.Row+len(inputLines)-1] = newLineFromString(part1 + part2)

	// finally, append the first line of input to the remainder
	// of the original line
	b.Lines[addr.Row] = newLineFromString(string(b.Lines[addr.Row].s[:split]) + inputLines[0])

	addr.Row += len(inputLines) - 1
	addr.Col = utf8.RuneCountInString(inputLines[len(inputLines)-1])
	return addr
}

// AutoSelect selects some text around a. Based on acme's double click selection rules.
func (b *Buffer) AutoSelect(addr Address) Selection {
	// select bracketed text
	if sel, ok := b.selDelimited(addr, "{[(<", "}])>"); ok {
		return sel
	}

	sel := Selection{addr, addr}
	line := b.Lines[addr.Row].Runes()

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
	var line = b.Lines[addr.Row].Runes()
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
		line := b.Lines[match.Row].Runes()
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
