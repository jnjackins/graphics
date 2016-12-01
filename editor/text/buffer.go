package text

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"

	"sigint.ca/graphics/editor/address"
)

// TODO: rename to Doc?
type Buffer struct {
	Lines []*Line
}

func NewBuffer() *Buffer {
	return &Buffer{
		Lines: []*Line{new(Line)},
	}
}

func (b *Buffer) NextSimple(a address.Simple) address.Simple {
	if a.Col < utf8.RuneCount(b.Lines[a.Row].s) {
		a.Col++
	} else if a.Row < len(b.Lines)-1 {
		a.Col = 0
		a.Row++
	}
	return a
}

func (b *Buffer) PrevSimple(a address.Simple) address.Simple {
	if a.Col > 0 {
		a.Col--
	} else if a.Row > 0 {
		a.Row--
		a.Col = utf8.RuneCount(b.Lines[a.Row].s)
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

func (b *Buffer) fixAddr(a address.Simple) address.Simple {
	if a.Row < 0 {
		a.Row = 0
	} else if a.Row > len(b.Lines)-1 {
		a.Row = len(b.Lines) - 1
	}

	if a.Col < 0 {
		a.Col = 0
	} else if a.Col > utf8.RuneCount(b.Lines[a.Row].s) {
		a.Col = utf8.RuneCount(b.Lines[a.Row].s)
	}
	return a
}

func (b *Buffer) GetSel(sel address.Selection) string {
	if sel.IsEmpty() {
		return ""
	}

	if sel.From.Row == sel.To.Row {
		row := sel.From.Row
		from := b.Lines[row].elemFromCol(sel.From.Col)
		to := b.Lines[row].elemFromCol(sel.To.Col)
		return string(b.Lines[row].s[from:to])
	}

	ret := make([]rune, 0, 20*(sel.To.Row-sel.From.Row))
	ret = append(ret, bytes.Runes(b.Lines[sel.From.Row].s)[sel.From.Col:]...)
	ret = append(ret, '\n')
	for i := sel.From.Row + 1; i < sel.To.Row; i++ {
		ret = append(ret, bytes.Runes(b.Lines[i].s)...)
		ret = append(ret, '\n')
	}
	ret = append(ret, bytes.Runes(b.Lines[sel.To.Row].s)[:sel.To.Col]...)
	return string(ret)
}

func (b *Buffer) ClearSel(sel address.Selection) address.Selection {
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
	return address.Selection{sel.From, sel.From}
}

func (b *Buffer) LastAddress() address.Simple {
	lastLine := len(b.Lines) - 1
	lastChar := utf8.RuneCount(b.Lines[lastLine].s)
	return address.Simple{Row: lastLine, Col: lastChar}
}

// InsertString inserts s into the buffer at a, adding new Lines if s
// contains newline characters.
func (b *Buffer) InsertString(addr address.Simple, s string) address.Simple {
	addr = b.fixAddr(addr)

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
func (b *Buffer) AutoSelect(addr address.Simple) address.Selection {
	// selections to attempt, in order:
	//  - bracketed text
	//  - the entire line
	//  - quoted text
	//  - a word

	if sel, ok := b.selDelimited(addr, "{[(<", "}])>"); ok {
		return sel
	}

	if addr.Col == utf8.RuneCount(b.Lines[addr.Row].s) || addr.Col == 0 {
		return b.SelLine(addr)
	}

	const quotes = "\"'`"
	if sel, ok := b.selDelimited(addr, quotes, quotes); ok {
		return sel
	}

	return b.SelWord(addr)
}

// SelLine selects a line.
func (b *Buffer) SelLine(addr address.Simple) address.Selection {
	sel := address.Selection{addr, addr}
	sel.From.Col = 0
	if addr.Row+1 < len(b.Lines) {
		sel.To.Row++
		sel.To.Col = 0
	} else {
		sel.To.Col = utf8.RuneCount(b.Lines[addr.Row].s)
	}
	return sel
}

// SelWord selects a word. If we're on a non-alphanumeric, attempt to select a word to
// the left of the click; otherwise expand across alphanumerics in both directions.
func (b *Buffer) SelWord(addr address.Simple) address.Selection {
	return b.SelFunc(addr, isWordSep)
}

func isWordSep(c rune) bool {
	return !(unicode.IsLetter(c) || unicode.IsNumber(c) || c == '_')
}

// SelFunc is analogous to strings.FieldsFunc, returning a selection specified
// by sepFn. SepFn should return true when sep is the desired separator.
func (b *Buffer) SelFunc(addr address.Simple, sepFn func(sep rune) bool) address.Selection {
	sel := address.Selection{addr, addr}

	line := bytes.Runes(b.Lines[addr.Row].s)
	for col := addr.Col; col > 0 && !sepFn(line[col-1]); col-- {
		sel.From.Col--
	}
	for col := addr.Col; col < len(line) && !sepFn(line[col]); col++ {
		sel.To.Col++
	}
	return sel
}

// SelDelimited returns true if a selection was attempted, successfully or not.
func (b *Buffer) selDelimited(addr address.Simple, leftDelims, rightDelims string) (address.Selection, bool) {
	sel := address.Selection{addr, addr}

	var delim int
	var line = bytes.Runes(b.Lines[addr.Row].s)
	var next func(address.Simple) address.Simple
	var rightwards bool
	if addr.Col > 0 {
		if delim = strings.IndexRune(leftDelims, line[addr.Col-1]); delim != -1 {
			// scan to the right, from a left delimiter
			next = b.NextSimple
			rightwards = true

			// the user double-clicked to the right of a left delimiter; move addr
			// to the delimiter itself
			addr.Col--
		}
	}
	if next == nil && addr.Col < len(line) {
		if delim = strings.IndexRune(rightDelims, line[addr.Col]); delim != -1 {
			// scan to the left, from a right delimiter
			// swap delimiters so that delim1 refers to the first one we encountered
			leftDelims, rightDelims = rightDelims, leftDelims
			next = b.PrevSimple
		}
	}
	if next == nil {
		return sel, false
	}

	stack := 0
	match := addr
	prev := address.Simple{-1, -1}
	for match != prev {
		prev = match
		match = next(match)
		line := bytes.Runes(b.Lines[match.Row].s)
		if match.Col > len(line)-1 {
			continue
		}
		c := line[match.Col]
		if c == rune(rightDelims[delim]) && stack == 0 {
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
		if leftDelims != rightDelims && c == rune(leftDelims[delim]) {
			stack++
		}
		if leftDelims != rightDelims && c == rune(rightDelims[delim]) {
			stack--
		}
	}
	return sel, true
}
