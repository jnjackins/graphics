package text

import "bytes"

type line struct {
	s     []rune
	adv   []int // advances (widths in pixels) of each rune in s
	dirty bool  // true if the line needs to be redrawn (px needs to be repopulated)
}

type selection struct {
	head, tail address // the beginning and end points of the selection
}

type address struct {
	row, col int
}

// loadRune inserts a single printable utf8 encoded character, replacing the current
// selection.
func (b *Buffer) loadRune(r rune, recordHist bool) {
	b.deleteSel(recordHist)
	if r == '\n' {
		b.loadLines([][]rune{{}})
		return
	}
	b.loadLine([]rune{r})

	if recordHist {
		ins := b.currentAction.ins
		if ins == nil {
			ins = &change{bounds: b.dot}
		}
		ins.bounds.tail = b.dot.tail
		ins.text = append(ins.text, b.contents(b.dot)...)
	}
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
		ins := b.currentAction.ins
		if ins == nil {
			ins = &change{bounds: b.dot}
		}
		ins.bounds.tail = b.dot.tail
		ins.text = append(ins.text, b.contents(b.dot)...)
	}
}

func (b *Buffer) loadLines(input [][]rune) {
	row, col := b.dot.head.row, b.dot.head.col

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

	// fix selection; b.dot.head is already fine
	b.dot.tail.row = row + len(newLines) - 1
	b.dot.tail.col = len(input[len(input)-1])
	b.dirtyLines(row, len(b.lines))
	b.autoScroll()
}

func (b *Buffer) loadLine(s []rune) {
	pos := b.dot.head
	b.lines[pos.row].s = insertRunes(b.lines[pos.row].s, s, pos.col)

	// TODO: why do we have to measure here? should be measured when drawn
	b.lines[pos.row].adv = b.font.measure(b.margin.X, b.lines[pos.row].s)

	b.dot.tail.col += len(s)
	b.dirtyLine(pos.row)
}

func insertRunes(dst, src []rune, pos int) []rune {
	return append(append(dst[:pos], src...), dst[pos:]...)
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
		b.currentAction.del = &change{bounds: b.dot, text: b.contents(b.dot)}
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
