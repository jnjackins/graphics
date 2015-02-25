package text

import (
	"image"
	"unicode"
)

func (b *Buffer) handleKey(r rune) {
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

	// up
	case 14:
		b.scroll(image.Pt(0, -18*b.font.height))

	// left
	case 17:
		b.left()

	// right
	case 18:
		b.right()

	// down
	case 128:
		b.scroll(image.Pt(0, 18*b.font.height))

	default:
		if unicode.IsGraphic(r) {
			b.input(r)
		}
	}
}

func (b *Buffer) input(r rune) {
	b.dirtyImg = true
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

func (b *Buffer) backspace() {
	b.dirtyImg = true
	b.deleteSel()
	head := b.dot.Head
	if head.Col > 0 {
		b.lines[head.Row].dirty = true
		head.Col--
		line := b.lines[head.Row].s
		line = append(line[:head.Col], line[head.Col+1:]...)
		b.lines[head.Row].s = line
	} else if head.Row > 0 {
		head.Row--
		head.Col = len(b.lines[head.Row].s)

		// delete the old line
		b.lines[head.Row].s = append(b.lines[head.Row].s, b.lines[head.Row+1].s...)
		b.lines = append(b.lines[:head.Row+1], b.lines[head.Row+2:]...)

		// redraw everything past here
		for _, line := range b.lines[head.Row:] {
			line.dirty = true
		}
		// make sure we clean up the garbage left after the (new) final line
		b.clear = b.img.Bounds()
		b.clear.Min.Y = b.font.height * (len(b.lines) - 1)
	}
	b.dot.Head, b.dot.Tail = head, head
}

func (b *Buffer) left() {
	b.dirtyImg = true
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
	b.dirtyImg = true
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
	b.dirtyImg = true
	b.deleteSel()
	row, col := b.dot.Head.Row, b.dot.Head.Col
	nl := &line{
		s:     make([]rune, 0, 100),
		dirty: true,
	}

	// since we need to shift everything down, all past here will need
	// to be redrawn.
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
