package text

import (
	"image"
	"log"
	"unicode"

	"sigint.ca/graphics/keys"
)

func (b *Buffer) handleKey(r rune) {
	key := r
	switch key {
	case keys.Backspace:
		b.backspace()
		b.commitAction()
	case keys.Return:
		b.newline()
		b.commitAction()
	case keys.Up:
		b.scroll(image.Pt(0, -18*b.font.height))
		b.commitAction()
	case keys.Left:
		b.left()
		b.commitAction()
	case keys.Right:
		b.right()
		b.commitAction()
	case keys.Down:
		b.scroll(image.Pt(0, 18*b.font.height))
		b.commitAction()
	case keys.Copy:
		b.snarf()
		b.commitAction()
	case keys.Paste:
		b.paste()
		b.commitAction()
	case keys.Cut:
		b.snarf()
		b.deleteSel(true)
		b.commitAction()
	case keys.Redo:
		b.commitAction()
		b.redo()
	case keys.Undo:
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
