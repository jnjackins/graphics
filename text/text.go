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

type Selection struct {
	Head, Tail Address // the beginning and end points of the selection
}

func (b *Buffer) input(r rune) {
	b.deleteSel()
	row, col := b.dot.Head.Row, b.dot.Head.Col
	b.lines[row].dirty = true

	if col == len(b.lines[row].s) {
		b.lines[row].s = append(b.lines[row].s, r)
	} else {
		line := string(b.lines[row].s)
		line = line[:col] + string(r) + line[col:]
		b.lines[row].s = []rune(line)
	}
	b.dot.Head.Col++
	b.dot.Tail = b.dot.Head
}

// load replaces the current selection with the contents of s. The cursor is left
// at the beginning of the new text, and the position at the end of the new text
// is returned.
func (b *Buffer) load(s string) {
	b.deleteSel()
	col1, row1 := b.dot.Head.Col, b.dot.Head.Row
	b.lines[row1].dirty = true

	inputlines := strings.Split(s, "\n")
	row2 := row1 + len(inputlines) - 1

	// save the rest of the line after the the cursor
	remainder := string(b.lines[row1].s[col1:])

	// insert the first line of the new text (or all of the new text, if
	// it is < 1 line
	b.lines[row1].s = append(b.lines[row1].s[:col1], []rune(inputlines[0])...)

	// create new lines as needed
	if len(inputlines) > 1 {
		inputlines = inputlines[1:]
		newlines := make([]*line, len(inputlines))
		for i := 0; i < len(inputlines); i++ {
			newlines[i] = &line{
				s:     []rune(inputlines[i]),
				dirty: true,
			}
		}
		if row1+1 < len(b.lines) {
			newlines = append(newlines, b.lines[row1+1:]...)
		}
		b.lines = append(b.lines[:row1+1], newlines...)
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
}

func (b *Buffer) deleteSel() {
	if b.dot.Head == b.dot.Tail {
		return
	}

	col1, row1, col2, row2 := b.dot.Head.Col, b.dot.Head.Row, b.dot.Tail.Col, b.dot.Tail.Row
	b.lines[row1].dirty = true
	b.lines[row2].dirty = true

	line := b.lines[row1].s
	line = b.lines[row1].s[:col1]
	b.lines[row1].s = append(line, b.lines[row2].s[col2:]...)
	if row2 > row1 {
		b.lines = append(b.lines[:row1+1], b.lines[row2+1:]...)
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
		return
	}

	// select word
	for col := a.Col; col > 0 && isAlnum(line[col-1]); col-- {
		b.dot.Head.Col--
	}

	// if we're on a non-alphanumeric, attempt to select only to the left.
	if isAlnum(line[a.Col]) {
		for col := a.Col; col < len(line) && isAlnum(line[col]); col++ {
			b.dot.Tail.Col++
		}
	}
}

const (
	selLeft = iota
	selRight
	selNone
)

// returns true if a selection was attempted, successfully or not
func (b *Buffer) selDelimited(delims1, delims2 string) bool {
	var dir int

	left, right := b.dot.Head, b.dot.Tail
	line := b.lines[left.Row].s
	var delim int
	var next func() Address
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
			dir = selRight
		}
	}
	if next == nil && left.Col < len(line) {
		if delim = strings.IndexRune(delims2, line[left.Col]); delim != -1 {
			// scan from right delimiter
			// swap delimiters so that delim1 refers to the first one we encountered
			tmp := delims1
			delims1 = delims2
			delims2 = tmp
			next = func() Address {
				if left.Col-1 < 0 {
					left.Row--
					if left.Row >= 0 {
						line = b.lines[left.Row].s
					}
					left.Col = len(line)
				} else {
					left.Col--
				}
				return left
			}
			dir = selLeft
		}
	}
	if next == nil {
		return false
	}
	stack := 0
	for {
		p := next()
		if p.Row < 0 || p.Row >= len(b.lines) {
			dir = selNone
			return true
		} else if p.Col >= len(b.lines[p.Row].s) {
			continue
		}
		c := b.lines[p.Row].s[p.Col]
		if c == rune(delims2[delim]) && stack == 0 {
			b.dot.Head, b.dot.Tail = left, right
			if dir == selLeft {
				b.dot.Head.Col++
			}
			return true
		} else if c == 0 {
			dir = selNone
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
