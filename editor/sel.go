package editor

import (
	"strings"
	"unicode"

	"sigint.ca/graphics/editor/internal/text"
)

func (b *Buffer) sel(a1, a2 text.Address) {
	b.dot.From = a1
	b.dot.To = a2
	b.dirtyLines(b.dot.From.Row, b.dot.To.Row+1)
}

func (b *Buffer) selAll() {
	last := len(b.lines) - 1
	b.sel(text.Address{0, 0}, text.Address{last, len(b.lines[last].s)})
}

// expandSel selects some text around a. Based on acme's double click selection rules.
func (b *Buffer) expandSel(a text.Address) {
	b.dot.From, b.dot.To = a, a
	line := b.lines[a.Row].s

	// select bracketed text
	if b.selDelimited("{[(<", "}])>") {
		b.dirtyLines(b.dot.From.Row, b.dot.To.Row+1)
		return
	}

	// select line
	if a.Col == len(line) || a.Col == 0 {
		b.dot.From.Col = 0
		if a.Row+1 < len(b.lines) {
			b.dot.To.Row++
			b.dot.To.Col = 0
		} else {
			b.dot.To.Col = len(line)
		}
		return
	}

	// select quoted text
	const quotes = "\"'`"
	if b.selDelimited(quotes, quotes) {
		b.dirtyLines(b.dot.From.Row, b.dot.To.Row+1)
		return
	}

	// Select a word. If we're on a non-alphanumeric, attempt to select a word to
	// the left of the click; otherwise expand across alphanumerics in both directions.
	for col := a.Col; col > 0 && isAlnum(line[col-1]); col-- {
		b.dot.From.Col--
	}
	if isAlnum(line[a.Col]) {
		for col := a.Col; col < len(line) && isAlnum(line[col]); col++ {
			b.dot.To.Col++
		}
	}
}

func isAlnum(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c)
}

// returns true if a selection was attempted, successfully or not
func (b *Buffer) selDelimited(delims1, delims2 string) bool {
	addr := b.dot.From
	var delim int
	var line = b.lines[addr.Row].s
	var next func(text.Address) text.Address
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
	prev := text.Address{-1, -1}
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
				b.dot.From, b.dot.To = addr, match
			} else {
				b.dot.From, b.dot.To = match, addr
			}
			b.dot.From.Col++ // move the head of the selection past the left delimiter
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

func (b *Buffer) nextAddress(a text.Address) text.Address {
	if a.Col < len(b.lines[a.Row].s) {
		a.Col++
	} else if a.Row < len(b.lines)-1 {
		a.Col = 0
		a.Row++
	}
	return a
}

func (b *Buffer) prevAddress(a text.Address) text.Address {
	if a.Col > 0 {
		a.Col--
	} else if a.Row > 0 {
		a.Row--
		a.Col = len(b.lines[a.Row].s)
	}
	return a
}
