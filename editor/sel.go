package editor

import (
	"strings"
	"unicode"

	"sigint.ca/graphics/editor/internal/text"
)

func (ed *Editor) sel(a1, a2 text.Address) {
	ed.dot.From = a1
	ed.dot.To = a2
	ed.dirtyLines(ed.dot.From.Row, ed.dot.To.Row+1)
}

func (ed *Editor) selAll() {
	last := len(ed.lines) - 1
	ed.sel(text.Address{0, 0}, text.Address{last, len(ed.lines[last].s)})
}

// expandSel selects some text around a. Based on acme's double click selection rules.
func (ed *Editor) expandSel(a text.Address) {
	ed.dot.From, ed.dot.To = a, a
	line := ed.lines[a.Row].s

	// select bracketed text
	if ed.selDelimited("{[(<", "}])>") {
		ed.dirtyLines(ed.dot.From.Row, ed.dot.To.Row+1)
		return
	}

	// select line
	if a.Col == len(line) || a.Col == 0 {
		ed.dot.From.Col = 0
		if a.Row+1 < len(ed.lines) {
			ed.dot.To.Row++
			ed.dot.To.Col = 0
		} else {
			ed.dot.To.Col = len(line)
		}
		return
	}

	// select quoted text
	const quotes = "\"'`"
	if ed.selDelimited(quotes, quotes) {
		ed.dirtyLines(ed.dot.From.Row, ed.dot.To.Row+1)
		return
	}

	// Select a word. If we're on a non-alphanumeric, attempt to select a word to
	// the left of the click; otherwise expand across alphanumerics in both directions.
	for col := a.Col; col > 0 && isAlnum(line[col-1]); col-- {
		ed.dot.From.Col--
	}
	if isAlnum(line[a.Col]) {
		for col := a.Col; col < len(line) && isAlnum(line[col]); col++ {
			ed.dot.To.Col++
		}
	}
}

func isAlnum(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c)
}

// returns true if a selection was attempted, successfully or not
func (ed *Editor) selDelimited(delims1, delims2 string) bool {
	addr := ed.dot.From
	var delim int
	var line = ed.lines[addr.Row].s
	var next func(text.Address) text.Address
	var rightwards bool
	if addr.Col > 0 {
		if delim = strings.IndexRune(delims1, line[addr.Col-1]); delim != -1 {
			// scan to the right, from a left delimiter
			next = ed.nextAddress
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
			next = ed.prevAddress
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
		line := ed.lines[match.Row].s
		if match.Col > len(line)-1 {
			continue
		}
		c := line[match.Col]
		if c == rune(delims2[delim]) && stack == 0 {
			if rightwards {
				ed.dot.From, ed.dot.To = addr, match
			} else {
				ed.dot.From, ed.dot.To = match, addr
			}
			ed.dot.From.Col++ // move the head of the selection past the left delimiter
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

func (ed *Editor) nextAddress(a text.Address) text.Address {
	if a.Col < len(ed.lines[a.Row].s) {
		a.Col++
	} else if a.Row < len(ed.lines)-1 {
		a.Col = 0
		a.Row++
	}
	return a
}

func (ed *Editor) prevAddress(a text.Address) text.Address {
	if a.Col > 0 {
		a.Col--
	} else if a.Row > 0 {
		a.Row--
		a.Col = len(ed.lines[a.Row].s)
	}
	return a
}
