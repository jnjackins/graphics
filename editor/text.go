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
func (ed *Editor) loadRune(r rune) {
	ed.dot = ed.clear(ed.dot)
	if r == '\n' {
		ed.loadLines([][]rune{{}})
		return
	}
	ed.loadLine([]rune{r})
}

// loadBytes replaces the current selection with s, and handles arbitrary
// utf8 input, including newlines.
func (ed *Editor) loadBytes(s []byte) {
	ed.dot = ed.clear(ed.dot)

	lines := bytes.Split(s, []byte("\n"))
	input := make([][]rune, len(lines))
	for i, line := range lines {
		input[i] = bytes.Runes(line)
	}

	if len(input) == 1 {
		ed.loadLine(input[0])
	} else {
		ed.loadLines(input)
	}
}

func (ed *Editor) loadLines(input [][]rune) {
	row, col := ed.dot.From.Row, ed.dot.From.Col

	newLines := make([]*line, len(input))

	// the beginning and end of the current line are attached to the first and last of the
	// lines that are being loaded
	newLines[0] = &line{s: append(ed.lines[row].s[:col], input[0]...)}
	newLines[0].adv = ed.font.measure(ed.margin.X, newLines[0].s)
	last := len(newLines) - 1
	newLines[last] = &line{s: append(input[len(input)-1], ed.lines[row].s[col:]...)}
	newLines[last].adv = ed.font.measure(ed.margin.X, newLines[last].s)

	// entirely new lines
	for i := 1; i < len(newLines)-1; i++ {
		newLines[i] = &line{s: input[i]}
	}

	// put everything together
	pre, post := ed.lines[:row], ed.lines[row+1:]
	ed.lines = append(append(pre, newLines...), post...)

	// fix selection; ed.dot.From is already fine
	ed.dot.To.Row = row + len(newLines) - 1
	ed.dot.To.Col = len(input[len(input)-1])
	ed.dirtyLines(row, len(ed.lines))
	ed.autoScroll()
}

func (ed *Editor) loadLine(s []rune) {
	addr := ed.dot.From
	ed.lines[addr.Row].s = insertRunes(ed.lines[addr.Row].s, s, addr.Col)

	// TODO: why do we have to measure here? should be measured when drawn
	ed.lines[addr.Row].adv = ed.font.measure(ed.margin.X, ed.lines[addr.Row].s)

	ed.dot.To.Col += len(s)
	ed.dirtyLine(addr.Row)
}

func insertRunes(dst, src []rune, pos int) []rune {
	return append(append(dst[:pos], src...), dst[pos:]...)
}

func (ed *Editor) contents(sel text.Selection) string {
	a1, a2 := sel.From, sel.To
	if a1.Row == a2.Row {
		return string(ed.lines[a1.Row].s[a1.Col:a2.Col])
	} else {
		sel := string(ed.lines[a1.Row].s[a1.Col:]) + "\n"
		for i := a1.Row + 1; i < a2.Row; i++ {
			sel += string(ed.lines[i].s) + "\n"
		}
		sel += string(ed.lines[a2.Row].s[:a2.Col])
		return sel
	}
}

func (ed *Editor) clear(sel text.Selection) text.Selection {
	if sel.IsEmpty() {
		return sel
	}
	col1, row1, col2, row2 := sel.From.Col, sel.From.Row, sel.To.Col, sel.To.Row
	line := ed.lines[row1].s[:col1]
	ed.lines[row1].s = append(line, ed.lines[row2].s[col2:]...)
	ed.dirtyLine(row1)
	if row2 > row1 {
		ed.lines = append(ed.lines[:row1+1], ed.lines[row2+1:]...)
		ed.dirtyLines(row1+1, len(ed.lines))

		// make sure we clean up the garbage left after the (new) final line
		ed.clearr = ed.img.Bounds()
		ed.clearr.Min.Y = ed.font.height * (len(ed.lines) - 1)
		ed.autoScroll()
		ed.shrinkImg()
	}
	sel.To = sel.From
	return sel
}
