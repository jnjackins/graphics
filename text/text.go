package text

import (
	"strings"
	"unicode"
)

const (
	leftbrackets  = "{[(<"
	rightbrackets = "}])>"
	quotes        = "'`\""
)

type line struct {
	s     []rune // TODO make this string
	px    []int  // x-coord of the rightmost pixels of each rune in s
	dirty bool   // true if the line needs to be redrawn (px needs to be repopulated)
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

// load replaces the current selection with s.
func (b *Buffer) load(s string, recordAction bool) {
	b.deleteSel(true)

	col1, row1 := b.dot.Head.Col, b.dot.Head.Row

	inputlines := strings.Split(s, "\n")
	row2 := row1 + len(inputlines) - 1

	// save the rest of the line after the the cursor
	remainder := string(b.lines[row1].s[col1:])

	// insert the first line of the new text, fixing px so that the
	// selection rect can be drawn correctly
	b.lines[row1].s = append(b.lines[row1].s[:col1], []rune(inputlines[0])...)
	b.lines[row1].px = b.font.getPx(b.margin.X, string(b.lines[row1].s))

	// create new lines as needed
	if len(inputlines) > 1 {
		inputlines = inputlines[1:]
		newlines := make([]*line, len(inputlines))
		for i := 0; i < len(inputlines); i++ {
			s := inputlines[i]
			newlines[i] = &line{
				s:  []rune(s),
				px: b.font.getPx(b.margin.X, s),
			}
		}
		if row1+1 < len(b.lines) {
			newlines = append(newlines, b.lines[row1+1:]...)
		}
		b.lines = append(b.lines[:row1+1], newlines...)
		b.dirtyLines(row1, len(b.lines))
	} else {
		b.dirtyLine(row1)
	}

	// add the remander of the line following the deleted text to the end
	// of the last line we modified or added
	b.lines[row2].s = append(b.lines[row2].s, []rune(remainder)...)

	var col2 int
	if row2 > row1 {
		col2 = len(inputlines[len(inputlines)-1])
	} else {
		col2 = col1 + len(s)
	}
	b.dot.Tail = Address{row2, col2}

	if recordAction {
		b.initCurrentAction()
		b.currentAction.insertionBounds = b.dot
		b.currentAction.insertionText = b.contents(b.dot)
	}
}

func (b *Buffer) contents(sel Selection) string {
	a1, a2 := sel.Head, sel.Tail
	if a1.Row == a2.Row {
		return string(b.lines[a1.Row].s[a1.Col:a2.Col])
	} else {
		sel := string(b.lines[a1.Row].s[a1.Col:]) + "\n"
		for i := a1.Row + 1; i < a2.Row; i++ {
			sel += string(b.lines[i].s) + "\n"
		}
		sel += string(b.lines[a2.Row].s[:a2.Col])
		return sel
	}
}

func (b *Buffer) deleteSel(recordAction bool) {
	if b.dot.Head == b.dot.Tail {
		return
	}

	if recordAction {
		b.initCurrentAction()
		b.currentAction.deletionBounds = b.dot
		b.currentAction.deletionText = b.contents(b.dot)
	}

	col1, row1, col2, row2 := b.dot.Head.Col, b.dot.Head.Row, b.dot.Tail.Col, b.dot.Tail.Row
	line := b.lines[row1].s[:col1]
	b.lines[row1].s = append(line, b.lines[row2].s[col2:]...)
	b.dirtyLine(row1)
	if row2 > row1 {
		b.lines = append(b.lines[:row1+1], b.lines[row2+1:]...)
		b.dirtyLines(row1+1, len(b.lines))

		// make sure we clean up the garbage left after the (new) final line
		b.clear = b.img.Bounds()
		b.clear.Min.Y = b.font.height * (len(b.lines) - 1)
	}
	b.dot.Tail = b.dot.Head
}

func isAlnum(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c)
}

// expandSel selects some text around a. Based on acme's double click selection rules.
func (b *Buffer) expandSel(a Address) {
	b.dot.Head, b.dot.Tail = a, a
	line := b.lines[a.Row].s

	// select bracketed text
	if b.selDelimited(leftbrackets, rightbrackets) {
		b.dirtyLines(b.dot.Head.Row, b.dot.Tail.Row+1)
		return
	}

	// select line
	if a.Col == len(line) || a.Col == 0 {
		b.dot.Head.Col = 0
		if a.Row+1 < len(b.lines) {
			b.dot.Tail.Row++
			b.dot.Tail.Col = 0
		} else {
			b.dot.Tail.Col = len(line)
		}
		return
	}

	// select quoted text
	if b.selDelimited(quotes, quotes) {
		b.dirtyLines(b.dot.Head.Row, b.dot.Tail.Row+1)
		return
	}

	// Select a word. If we're on a non-alphanumeric, attempt to select a word to
	// the left of the click; otherwise expand across alphanumerics in both directions.
	for col := a.Col; col > 0 && isAlnum(line[col-1]); col-- {
		b.dot.Head.Col--
	}
	if isAlnum(line[a.Col]) {
		for col := a.Col; col < len(line) && isAlnum(line[col]); col++ {
			b.dot.Tail.Col++
		}
	}
}

// returns true if a selection was attempted, successfully or not
func (b *Buffer) selDelimited(delims1, delims2 string) bool {
	left, right := b.dot.Head, b.dot.Tail
	line := b.lines[left.Row].s
	var delim int
	var next func() Address

	// First see if we can scan to the right.
	if left.Col > 0 {
		if delim = strings.IndexRune(delims1, line[left.Col-1]); delim != -1 {
			// scan from left delimiter
			next = func() Address {
				if right.Col+1 > len(line) {
					right.Row++
					if right.Row < len(b.lines) {
						line = b.lines[right.Row].s
					}
					right.Col = 0
				} else {
					right.Col++
				}
				return right
			}
		}
	}

	// Otherwise, see if we can scan to the left.
	var leftwards bool
	if next == nil && left.Col < len(line) {
		if delim = strings.IndexRune(delims2, line[left.Col]); delim != -1 {
			// scan from right delimiter
			leftwards = true
			// swap delimiters so that delim1 refers to the first one we encountered
			delims1, delims2 = delims2, delims1
			next = func() Address {
				if left.Col-1 < 0 {
					left.Row--
					if left.Row >= 0 {
						left.Col = len(b.lines[left.Row].s)
					}
				} else {
					left.Col--
				}
				return left
			}
		}
	}

	// Either we're not on a delimiter or there's nowhere to scan. Bail.
	if next == nil {
		return false
	}

	// We're on a valid delimiter and have a next function. Scan for the matching delimiter.
	stack := 0
	for {
		p := next()
		if p.Row < 0 || p.Row >= len(b.lines) {
			return true
		} else if p.Col >= len(b.lines[p.Row].s) {
			continue
		}
		c := b.lines[p.Row].s[p.Col]
		if c == rune(delims2[delim]) && stack == 0 {
			b.dot.Head, b.dot.Tail = left, right
			if leftwards {
				b.dot.Head.Col++
			}
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
}
