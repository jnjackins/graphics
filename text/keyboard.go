package text

import (
	"image"
	"log"
	"unicode"

	"golang.org/x/mobile/event/key"
)

func (b *Buffer) handleKey(e key.Event) {
	switch {
	case e.Code == key.CodeDeleteBackspace:
		b.backspace()
		b.commitAction()
	case e.Code == key.CodeReturnEnter:
		b.newline()
		b.commitAction()
	case e.Code == key.CodeTab:
		b.input('\t')
	case e.Code == key.CodeUpArrow:
		b.scroll(image.Pt(0, -18*b.lineHeight))
		b.commitAction()
	case e.Code == key.CodeLeftArrow:
		b.left()
		b.commitAction()
	case e.Code == key.CodeRightArrow:
		b.right()
		b.commitAction()
	case e.Code == key.CodeDownArrow:
		b.scroll(image.Pt(0, 18*b.lineHeight))
		b.commitAction()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeC:
		// copy
		b.snarf()
		b.commitAction()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeV:
		// paste
		b.paste()
		b.commitAction()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeX:
		// cut
		b.snarf()
		b.deleteSel(true)
		b.commitAction()
	case e.Modifiers == key.ModMeta|key.ModShift && e.Code == key.CodeZ:
		// redo
		b.commitAction()
		b.redo()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeZ:
		// undo
		b.commitAction()
		b.undo()
	default:
		if unicode.IsGraphic(e.Rune) {
			b.input(e.Rune)
		} else {
			log.Printf("text: unhandled key: %d\n", e.Rune)
		}
	}
}

func (b *Buffer) input(r rune) {
	b.loadRune(r, true)
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
	b.loadBytes([]byte("\n"), true)
	b.Dot.Head = b.Dot.Tail
}
