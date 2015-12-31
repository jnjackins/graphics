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

type address struct {
	row, col int
}

func (a1 address) lessThan(a2 address) bool {
	return a1.row < a2.row || (a1.row == a2.row && a1.col < a2.col)
}

func (b *Buffer) nextaddress(a address) address {
	if a.col < len(b.lines[a.row].s) {
		a.col++
	} else if a.row < len(b.lines)-1 {
		a.col = 0
		a.row++
	}
	return a
}

func (b *Buffer) prevaddress(a address) address {
	if a.col > 0 {
		a.col--
	} else if a.row > 0 {
		a.row--
		a.col = len(b.lines[a.row].s)
	}
	return a
}

type selection struct {
	head, tail address // the beginning and end points of the selection
}

// loadRune inserts a printable utf8 encoded character, replacing the current
// selection.
// TODO: history?
func (b *Buffer) loadRune(r rune, recordHist bool) {
	b.deleteSel(recordHist)
	if r == '\n' {
		b.loadLines([][]rune{{}})
		return
	}
	b.loadLine([]rune{r})
}

// loadBytes replaces the current selection with s, and handles arbitrary
// utf8 input, including newlines.
func (b *Buffer) loadBytes(s []byte, recordHist bool) {
	b.deleteSel(recordHist)

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
			b.currentAction.insertion = &change{bounds: b.dot, text: b.contents(b.dot)}
		} else {
			// append to b.currentAction if the user simply typed another rune
			ins.bounds.tail = b.dot.tail
			ins.text = append(ins.text, b.contents(b.dot)...)
		}
	}
}

func (b *Buffer) loadLines(input [][]rune) {
	row, col := b.dot.head.row, b.dot.head.col

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

	// fix selection; b.dot.head is already fine
	b.dot.tail.row = row + len(lNew) - 1
	b.dot.tail.col = len(input[len(input)-1])
	b.dirtyLines(row, len(b.lines))
	b.autoScroll()
}

func (b *Buffer) loadLine(s []rune) {
	row, col := b.dot.head.row, b.dot.head.col
	before := b.lines[row].s[:col]
	after := b.lines[row].s[col:]
	b.lines[row].s = append(append(before, s...), after...)
	b.lines[row].px = b.font.measure(b.margin.X, b.lines[row].s)
	b.dot.tail.col += len(s)
	b.dirtyLine(row)
}

func (b *Buffer) contents(sel selection) []byte {
	a1, a2 := sel.head, sel.tail
	if a1.row == a2.row {
		return []byte(string(b.lines[a1.row].s[a1.col:a2.col]))
	} else {
		sel := string(b.lines[a1.row].s[a1.col:]) + "\n"
		for i := a1.row + 1; i < a2.row; i++ {
			sel += string(b.lines[i].s) + "\n"
		}
		sel += string(b.lines[a2.row].s[:a2.col])
		return []byte(sel)
	}
}

func (b *Buffer) deleteSel(recordHist bool) {
	if b.dot.head == b.dot.tail {
		return
	}
	if recordHist {
		b.currentAction.deletion = &change{bounds: b.dot, text: b.contents(b.dot)}
	}
	col1, row1, col2, row2 := b.dot.head.col, b.dot.head.row, b.dot.tail.col, b.dot.tail.row
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
	b.dot.tail = b.dot.head
}

func isAlnum(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c)
}

func (b *Buffer) sel(a1, a2 address) {
	b.dot.head = a1
	b.dot.tail = a2
	b.dirtyLines(b.dot.head.row, b.dot.tail.row+1)
}

// expandSel selects some text around a. Based on acme's double click selection rules.
func (b *Buffer) expandSel(a address) {
	b.dot.head, b.dot.tail = a, a
	line := b.lines[a.row].s

	// select bracketed text
	if b.selDelimited(leftbrackets, rightbrackets) {
		b.dirtyLines(b.dot.head.row, b.dot.tail.row+1)
		return
	}

	// select line
	if a.col == len(line) || a.col == 0 {
		b.dot.head.col = 0
		if a.row+1 < len(b.lines) {
			b.dot.tail.row++
			b.dot.tail.col = 0
		} else {
			b.dot.tail.col = len(line)
		}
		return
	}

	// select quoted text
	if b.selDelimited(quotes, quotes) {
		b.dirtyLines(b.dot.head.row, b.dot.tail.row+1)
		return
	}

	// Select a word. If we're on a non-alphanumeric, attempt to select a word to
	// the left of the click; otherwise expand across alphanumerics in both directions.
	for col := a.col; col > 0 && isAlnum(line[col-1]); col-- {
		b.dot.head.col--
	}
	if isAlnum(line[a.col]) {
		for col := a.col; col < len(line) && isAlnum(line[col]); col++ {
			b.dot.tail.col++
		}
	}
}

// returns true if a selection was attempted, successfully or not
func (b *Buffer) selDelimited(delims1, delims2 string) bool {
	addr := b.dot.head
	var delim int
	var line = b.lines[addr.row].s
	var next func(address) address
	var rightwards bool
	if addr.col > 0 {
		if delim = strings.IndexRune(delims1, line[addr.col-1]); delim != -1 {
			// scan to the right, from a left delimiter
			next = b.nextaddress
			rightwards = true

			// the user double-clicked to the right of a left delimiter; move addr
			// to the delimiter itself
			addr.col--
		}
	}
	if next == nil && addr.col < len(line) {
		if delim = strings.IndexRune(delims2, line[addr.col]); delim != -1 {
			// scan to the left, from a right delimiter
			// swap delimiters so that delim1 refers to the first one we encountered
			delims1, delims2 = delims2, delims1
			next = b.prevaddress
		}
	}
	if next == nil {
		return false
	}

	stack := 0
	match := addr
	prev := address{-1, -1}
	for match != prev {
		prev = match
		match = next(match)
		line := b.lines[match.row].s
		if match.col > len(line)-1 {
			continue
		}
		c := line[match.col]
		if c == rune(delims2[delim]) && stack == 0 {
			if rightwards {
				b.dot.head, b.dot.tail = addr, match
			} else {
				b.dot.head, b.dot.tail = match, addr
			}
			b.dot.head.col++ // move the head of the selection past the left delimiter
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
