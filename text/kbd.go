package text

import (
	"image"
	"log"
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
		b.pushState()

	// return
	case 10:
		b.newline()
		b.pushState()

	// up
	case 14:
		b.scroll(image.Pt(0, -18*b.font.height))
		b.pushState()

	// left
	case 17:
		b.left()
		b.pushState()

	// right
	case 18:
		b.right()
		b.pushState()

	// down
	case 128:
		b.scroll(image.Pt(0, 18*b.font.height))
		b.pushState()

	// cmd-c
	case 61795:
		b.snarf()
		b.pushState()

	// cmd-v
	case 61814:
		b.paste()
		b.pushState()

	// cmd-x
	case 61816:
		b.snarf()
		b.deleteSel()
		b.pushState()

	// cmd-y
	case 61817:
		b.redo()

	// cmd-z
	case 61818:
		b.undo()

	default:
		if unicode.IsGraphic(r) || r == '\t' {
			b.input(r)
		} else {
			log.Printf("text: unhandled key: %d\n", r)
		}
	}
}

func (b *Buffer) input(r rune) {
	b.deleteSel()
	row, col := b.dot.Head.Row, b.dot.Head.Col
	b.dirtyLine(row)

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
	b.deleteSel()
	head := b.dot.Head
	if head.Col > 0 {
		b.dirtyLine(head.Row)
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
		b.dirtyLines(head.Row, len(b.lines))

		// make sure we clean up the garbage left after the (new) final line
		b.clear = b.img.Bounds()
		b.clear.Min.Y = b.font.height * (len(b.lines) - 1)
	}
	b.dot.Head, b.dot.Tail = head, head
}

func (b *Buffer) left() {
	head := b.dot.Head
	b.dirtyLine(head.Row)
	if head.Col > 0 {
		head.Col--
	} else if head.Row > 0 {
		head.Row--
		b.dirtyLine(head.Row)
		head.Col = len(b.lines[head.Row].s)
	}
	b.dot.Head, b.dot.Tail = head, head
}

func (b *Buffer) right() {
	head := b.dot.Head
	b.dirtyLine(head.Row)
	if head.Col < len(b.lines[head.Row].s) {
		head.Col++
	} else if head.Row < len(b.lines)-1 {
		head.Row++
		b.dirtyLine(head.Row)
		head.Col = 0
	}
	b.dot.Head, b.dot.Tail = head, head
}

func (b *Buffer) newline() {
	b.deleteSel()
	row, col := b.dot.Head.Row, b.dot.Head.Col
	nl := &line{s: []rune{}, px: []int{b.margin.X}}

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
	b.dirtyLines(row, len(b.lines))

	b.dot.Head.Col = 0
	b.dot.Head.Row++
	b.dot.Tail = b.dot.Head
}
