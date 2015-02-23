package text

import "unicode"

func (b *Buffer) handleKey(r rune) {
	b.dirty = true
	key := r

	// fix left, right, and up on OSX
	if (key & 0xff00) == 0xf000 {
		key = key & 0xff
	}

	switch key {
	// backspace
	case 8:
		b.backspace()

	// return
	case 10:
		b.newline()

	// left
	case 17:
		b.left()

	// right
	case 18:
		b.right()

	default:
		if unicode.IsGraphic(r) {
			b.input(r)
		}
	}
}

func (b *Buffer) backspace() {
	b.deleteSel()
	head := b.dot.Head
	row, col := head.Row, head.Col
	if col > 0 {
		b.lines[row].dirty = true
		head.Col--
		line := b.lines[row].s
		line = append(line[:head.Col], line[head.Col+1:]...)
		b.lines[row].s = line
	} else if row > 0 {
		head.Row--
		row--

		// all following lines will need to be redraw at their new row,
		// and the end of the image cleared.
		//
		// TODO: perhaps we should shift the lines by redrawing a subrectangle
		// of b.img one b.font.height higher, instead of marking all the lower
		// lines as dirty?
		b.clear = b.img.Bounds()
		b.clear.Min.Y = b.font.height * (len(b.lines) - 1)
		for _, line := range b.lines[row:] {
			line.dirty = true
		}

		head.Col = len(b.lines[row].s)

		// delete the old line
		b.lines[row].s = append(b.lines[row].s, b.lines[row+1].s...)
		b.lines = append(b.lines[:row+1], b.lines[row+2:]...)
	}
	b.dot.Head, b.dot.Tail = head, head
}

func (b *Buffer) left() {
	head := b.dot.Head
	if head.Col > 0 {
		b.lines[head.Row].dirty = true
		head.Col--
	} else if head.Row > 0 {
		b.lines[head.Row].dirty = true
		head.Row--
		b.lines[head.Row].dirty = true
		head.Col = len(b.lines[head.Row].s)
	}
	b.dot.Head, b.dot.Tail = head, head
}

func (b *Buffer) right() {
	head := b.dot.Head
	if head.Col < len(b.lines[head.Row].s) {
		b.lines[head.Row].dirty = true
		head.Col++
	} else if head.Row < len(b.lines)-1 {
		b.lines[head.Row].dirty = true
		head.Row++
		b.lines[head.Row].dirty = true
		head.Col = 0
	}
	b.dot.Head, b.dot.Tail = head, head
}

func (b *Buffer) newline() {
	b.deleteSel()
	row, col := b.dot.Head.Row, b.dot.Head.Col
	nl := &line{
		s:     make([]rune, 0, 100),
		dirty: true,
	}

	// TODO: perhaps we should shift the lines by redrawing a subrectangle
	// of b.img one b.font.height lower, instead of marking all the lower
	// lines as dirty?
	for _, line := range b.lines[row:] {
		line.dirty = true
	}

	if col == len(b.lines[row].s) && row == len(b.lines)-1 {
		// easy case, dot is at the end of the final line
		b.lines = append(b.lines, nl)
	} else {
		// insert a new line at y+1
		b.lines = append(b.lines, nil)
		copy(b.lines[row+2:], b.lines[row+1:])
		b.lines[row+1] = nl

		// update strings for the y+1 and y
		b.lines[row+1].s = b.lines[row].s[col:]
		b.lines[row].s = b.lines[row].s[:col]
	}
	b.dot.Head.Col = 0
	b.dot.Head.Row++
	b.dot.Tail = b.dot.Head
}
