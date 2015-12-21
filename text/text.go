package text

import (
	"bytes"
	"strings"
	"unicode"
)

const (
	leftbrackets  = "{[(<"
	rightbrackets = "}])>"
	quotes        = "'`\""
)

type line struct {
	s     []rune
	px    []int // x-coord of the rightmost pixels of each rune in s
	dirty bool  // true if the line needs to be redrawn (px needs to be repopulated)
}

type Address struct {
	Row, Col int
}

func (a1 Address) lessThan(a2 Address) bool {
	return a1.Row < a2.Row || (a1.Row == a2.Row && a1.Col < a2.Col)
}

func (b *Buffer) nextAddress(a Address) Address {
	if a.Col < len(b.lines[a.Row].s) {
		a.Col++
	} else if a.Row < len(b.lines)-1 {
		a.Col = 0
		a.Row++
	}
	return a
}

func (b *Buffer) prevAddress(a Address) Address {
	if a.Col > 0 {
		a.Col--
	} else if a.Row > 0 {
		a.Row--
		a.Col = len(b.lines[a.Row].s)
	}
	return a
}

type Selection struct {
	Head, Tail Address // the beginning and end points of the selection
}

func (b *Buffer) loadBytes(s []byte, recordHist bool) {
	b.deleteSel(true)

	lines := bytes.Split(s, []byte("\n"))
	input := make([][]rune, len(lines))
	for i, line := range lines {
		input[i] = bytes.Runes(line)
	}

	if len(input) == 1 {
		b.loadLine(input[0])
	} else {
		b.loadLines(input)
	}

	if recordHist {
		ins := b.currentAction.insertion
		if ins == nil {
			b.currentAction.insertion = &change{bounds: b.Dot, text: b.contents(b.Dot)}
		} else {
			// append to b.currentAction if the user simply typed another rune
			ins.bounds.Tail = b.Dot.Tail
			ins.text = append(ins.text, b.contents(b.Dot)...)
		}
	}
}

func (b *Buffer) loadRune(r rune, recordHist bool) {
	b.deleteSel(true)

	if r == '\n' {
		b.loadLines([][]rune{{'\n'}})
		return
	}
	b.loadLine([]rune{r})
}

// load replaces the current selection with s.
func (b *Buffer) loadLines(input [][]rune) {
	row, col := b.Dot.Head.Row, b.Dot.Head.Col

	// unchanged lines
	lPreceding := b.lines[:row]
	lFollowing := b.lines[row+1:]

	lNew := make([]*line, len(input))

	// the beginning and end of the current line are attached to the first and last of the
	// lines that are being loaded
	lNew[0] = &line{s: append(b.lines[row].s[:col], input[0]...)}
	lNew[0].px = b.font.measure(b.margin.X, lNew[0].s)
	last := len(lNew) - 1
	lNew[last] = &line{s: append(input[len(input)-1], b.lines[row].s[col:]...)}
	lNew[last].px = b.font.measure(b.margin.X, lNew[last].s)

	// entirely new lines
	for i := 1; i < len(lNew)-1; i++ {
		lNew[i] = &line{s: input[i]}
	}

	// put everything together
	b.lines = append(lPreceding, append(lNew, lFollowing...)...)

	// fix selection; b.Dot.Head is already fine
	b.Dot.Tail.Row = row + len(lNew) - 1
	b.Dot.Tail.Col = len(input[len(input)-1])
	b.dirtyLines(row, len(b.lines))
	b.autoScroll()
}

// loadLine inserts a string with no line breaks at b.Dot, assuming an empty selection.
func (b *Buffer) loadLine(s []rune) {
	row, col := b.Dot.Head.Row, b.Dot.Head.Col
	before := b.lines[row].s[:col]
	after := b.lines[row].s[col:]
	b.lines[row].s = append(append(before, s...), after...)
	b.lines[row].px = b.font.measure(b.margin.X, b.lines[row].s)
	b.Dot.Tail.Col += len(s)
	b.dirtyLine(row)
}

func (b *Buffer) contents(sel Selection) []byte {
	a1, a2 := sel.Head, sel.Tail
	if a1.Row == a2.Row {
		return []byte(string(b.lines[a1.Row].s[a1.Col:a2.Col]))
	} else {
		sel := string(b.lines[a1.Row].s[a1.Col:]) + "\n"
		for i := a1.Row + 1; i < a2.Row; i++ {
			sel += string(b.lines[i].s) + "\n"
		}
		sel += string(b.lines[a2.Row].s[:a2.Col])
		return []byte(sel)
	}
}

func (b *Buffer) deleteSel(recordAction bool) {
	if b.Dot.Head == b.Dot.Tail {
		return
	}
	if recordAction {
		b.currentAction.deletion = &change{bounds: b.Dot, text: b.contents(b.Dot)}
	}
	col1, row1, col2, row2 := b.Dot.Head.Col, b.Dot.Head.Row, b.Dot.Tail.Col, b.Dot.Tail.Row
	line := b.lines[row1].s[:col1]
	b.lines[row1].s = append(line, b.lines[row2].s[col2:]...)
	b.dirtyLine(row1)
	if row2 > row1 {
		b.lines = append(b.lines[:row1+1], b.lines[row2+1:]...)
		b.dirtyLines(row1+1, len(b.lines))

		// make sure we clean up the garbage left after the (new) final line
		b.clear = b.img.Bounds()
		b.clear.Min.Y = b.font.height * (len(b.lines) - 1)
		b.autoScroll()
		b.shrinkImg()
	}
	b.Dot.Tail = b.Dot.Head
}

func isAlnum(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c)
}

// expandSel selects some text around a. Based on acme's double click selection rules.
func (b *Buffer) expandSel(a Address) {
	b.Dot.Head, b.Dot.Tail = a, a
	line := b.lines[a.Row].s

	// select bracketed text
	if b.selDelimited(leftbrackets, rightbrackets) {
		b.dirtyLines(b.Dot.Head.Row, b.Dot.Tail.Row+1)
		return
	}

	// select line
	if a.Col == len(line) || a.Col == 0 {
		b.Dot.Head.Col = 0
		if a.Row+1 < len(b.lines) {
			b.Dot.Tail.Row++
			b.Dot.Tail.Col = 0
		} else {
			b.Dot.Tail.Col = len(line)
		}
		return
	}

	// select quoted text
	if b.selDelimited(quotes, quotes) {
		b.dirtyLines(b.Dot.Head.Row, b.Dot.Tail.Row+1)
		return
	}

	// Select a word. If we're on a non-alphanumeric, attempt to select a word to
	// the left of the click; otherwise expand across alphanumerics in both directions.
	for col := a.Col; col > 0 && isAlnum(line[col-1]); col-- {
		b.Dot.Head.Col--
	}
	if isAlnum(line[a.Col]) {
		for col := a.Col; col < len(line) && isAlnum(line[col]); col++ {
			b.Dot.Tail.Col++
		}
	}
}

// returns true if a selection was attempted, successfully or not
func (b *Buffer) selDelimited(delims1, delims2 string) bool {
	addr := b.Dot.Head
	var delim int
	var line = b.lines[addr.Row].s
	var next func(Address) Address
	var rightwards bool
	if addr.Col > 0 {
		if delim = strings.IndexRune(delims1, line[addr.Col-1]); delim != -1 {
			// scan to the right, from a left delimiter
			next = b.nextAddress
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
			next = b.prevAddress
		}
	}
	if next == nil {
		return false
	}

	stack := 0
	match := addr
	prev := Address{-1, -1}
	for match != prev {
		prev = match
		match = next(match)
		line := b.lines[match.Row].s
		if match.Col > len(line)-1 {
			continue
		}
		c := line[match.Col]
		if c == rune(delims2[delim]) && stack == 0 {
			if rightwards {
				b.Dot.Head, b.Dot.Tail = addr, match
			} else {
				b.Dot.Head, b.Dot.Tail = match, addr
			}
			b.Dot.Head.Col++ // move the head of the selection past the left delimiter
			return true
		} else if c == 0 {
			return true
		}
		if delims1 != delims2 && c == rune(delims1[delim]) {
			stack++
		}
		if delims1 != delims2 && c == rune(delims2[delim]) {
			stack--
		}
	}
	return true
}
