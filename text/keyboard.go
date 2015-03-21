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
		b.commitAction()

	// return
	case 10:
		b.newline()
		b.commitAction()

	// up
	case 14:
		b.scroll(image.Pt(0, -18*b.font.height))
		b.commitAction()

	// left
	case 17:
		b.left()
		b.commitAction()

	// right
	case 18:
		b.right()
		b.commitAction()

	// down
	case 128:
		b.scroll(image.Pt(0, 18*b.font.height))
		b.commitAction()

	// cmd-c
	case 61795:
		b.snarf()
		b.commitAction()

	// cmd-v
	case 61814:
		b.paste()
		b.commitAction()

	// cmd-x
	case 61816:
		b.snarf()
		b.deleteSel(true)
		b.commitAction()

	// cmd-y
	case 61817:
		b.commitAction()
		b.redo()

	// cmd-z
	case 61818:
		b.commitAction()
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
	b.load(string(r), true)
	b.Dot.Head = b.Dot.Tail
}

func (b *Buffer) backspace() {
	b.Dot.Head = b.prevAddress(b.Dot.Head)
	b.deleteSel(false) // don't update the current action

	if b.currentAction.insertion != nil {
		// This is the only case where the insertion must happen before the deletion.
		// Update b.currentAction manually here to make it an insertion only.
		b.currentAction.insertion.bounds.Tail.Col--
		text := b.currentAction.insertion.text
		b.currentAction.insertion.text = text[:len(text)-1]
	}
}

func (b *Buffer) left() {
	b.dirtyLines(b.Dot.Head.Row, b.Dot.Tail.Row+1)
	a := b.prevAddress(b.Dot.Head)
	b.Dot.Head, b.Dot.Tail = a, a
	b.dirtyLine(b.Dot.Head.Row) // new dot may be in a higher row
}

func (b *Buffer) right() {
	b.dirtyLines(b.Dot.Head.Row, b.Dot.Tail.Row+1)
	a := b.nextAddress(b.Dot.Tail)
	b.Dot.Head, b.Dot.Tail = a, a
	b.dirtyLine(b.Dot.Head.Row) // new dot may be in a lower row
}

func (b *Buffer) newline() {
	b.load("\n", true)
	b.Dot.Head = b.Dot.Tail
}
