package text

import (
	"strings"
	"unicode"
)

func (b *Buffer) sel(a1, a2 address) {
	b.dot.head = a1
	b.dot.tail = a2
	b.dirtyLines(b.dot.head.row, b.dot.tail.row+1)
}

func (b *Buffer) selAll() {
	last := len(b.lines) - 1
	b.sel(address{0, 0}, address{last, len(b.lines[last].s)})
}

// expandSel selects some text around a. Based on acme's double click selection rules.
func (b *Buffer) expandSel(a address) {
	b.dot.head, b.dot.tail = a, a
	line := b.lines[a.row].s

	// select bracketed text
	if b.selDelimited("{[(<", "}])>") {
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
	const quotes = "\"'`"
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

func isAlnum(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c)
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
			next = b.nextAddress
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
			next = b.prevAddress
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

func (b *Buffer) nextAddress(a address) address {
	if a.col < len(b.lines[a.row].s) {
		a.col++
	} else if a.row < len(b.lines)-1 {
		a.col = 0
		a.row++
	}
	return a
}

func (b *Buffer) prevAddress(a address) address {
	if a.col > 0 {
		a.col--
	} else if a.row > 0 {
		a.row--
		a.col = len(b.lines[a.row].s)
	}
	return a
}
