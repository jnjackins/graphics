package editor

import (
	"bytes"

	"sigint.ca/graphics/editor/internal/text"
)

type line struct {
	s     []rune
	adv   []int // advances (widths in pixels) of each rune in s
	dirty bool  // true if the line needs to be redrawn (px needs to be repopulated)
}

// loadRune inserts a single printable utf8 encoded character, replacing
// any selection
func (b *Buffer) loadRune(r rune) {
	b.dot = b.clear(b.dot)
	if r == '\n' {
		b.loadLines([][]rune{{}})
		return
	}
	b.loadLine([]rune{r})
}

// loadBytes replaces the current selection with s, and handles arbitrary
// utf8 input, including newlines.
func (b *Buffer) loadBytes(s []byte) {
	b.dot = b.clear(b.dot)

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
}

func (b *Buffer) loadLines(input [][]rune) {
	row, col := b.dot.From.Row, b.dot.From.Col

	newLines := make([]*line, len(input))

	// the beginning and end of the current line are attached to the first and last of the
	// lines that are being loaded
	newLines[0] = &line{s: append(b.lines[row].s[:col], input[0]...)}
	newLines[0].adv = b.font.measure(b.margin.X, newLines[0].s)
	last := len(newLines) - 1
	newLines[last] = &line{s: append(input[len(input)-1], b.lines[row].s[col:]...)}
	newLines[last].adv = b.font.measure(b.margin.X, newLines[last].s)

	// entirely new lines
	for i := 1; i < len(newLines)-1; i++ {
		newLines[i] = &line{s: input[i]}
	}

	// put everything together
	pre, post := b.lines[:row], b.lines[row+1:]
	b.lines = append(append(pre, newLines...), post...)

	// fix selection; b.dot.From is already fine
	b.dot.To.Row = row + len(newLines) - 1
	b.dot.To.Col = len(input[len(input)-1])
	b.dirtyLines(row, len(b.lines))
	b.autoScroll()
}

func (b *Buffer) loadLine(s []rune) {
	addr := b.dot.From
	b.lines[addr.Row].s = insertRunes(b.lines[addr.Row].s, s, addr.Col)

	// TODO: why do we have to measure here? should be measured when drawn
	b.lines[addr.Row].adv = b.font.measure(b.margin.X, b.lines[addr.Row].s)

	b.dot.To.Col += len(s)
	b.dirtyLine(addr.Row)
}

func insertRunes(dst, src []rune, pos int) []rune {
	return append(append(dst[:pos], src...), dst[pos:]...)
}

func (b *Buffer) contents(sel text.Selection) string {
	a1, a2 := sel.From, sel.To
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

func (b *Buffer) clear(sel text.Selection) text.Selection {
	if sel.IsEmpty() {
		return sel
	}
	col1, row1, col2, row2 := sel.From.Col, sel.From.Row, sel.To.Col, sel.To.Row
	line := b.lines[row1].s[:col1]
	b.lines[row1].s = append(line, b.lines[row2].s[col2:]...)
	b.dirtyLine(row1)
	if row2 > row1 {
		b.lines = append(b.lines[:row1+1], b.lines[row2+1:]...)
		b.dirtyLines(row1+1, len(b.lines))

		// make sure we clean up the garbage left after the (new) final line
		b.clearr = b.img.Bounds()
		b.clearr.Min.Y = b.font.height * (len(b.lines) - 1)
		b.autoScroll()
		b.shrinkImg()
	}
	sel.To = sel.From
	return sel
}
