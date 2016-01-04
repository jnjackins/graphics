package text

import (
	"image"
	"unicode"

	"golang.org/x/mobile/event/key"
)

func (b *Buffer) handleKeyEvent(e key.Event) {
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
	case e.Modifiers == key.ModMeta && e.Code == key.CodeA:
		last := len(b.lines) - 1
		b.sel(address{0, 0}, address{last, len(b.lines[last].s)})
	case e.Modifiers == key.ModMeta|key.ModShift && e.Code == key.CodeZ:
		// redo
		b.commitAction()
		b.redo()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeZ:
		// undo
		b.commitAction()
		b.undo()
	default:
		if unicode.IsGraphic(e.Rune) && e.Modifiers&key.ModMeta == 0 {
			b.input(e.Rune)
		}
	}
}

func (b *Buffer) input(r rune) {
	b.loadRune(r, true)
	b.dot.head = b.dot.tail
}

func (b *Buffer) backspace() {
	b.dot.head = b.prevaddress(b.dot.head)
	b.deleteSel(false) // don't update the current action

	if b.currentAction.ins != nil {
		// This is the only case where the insertion must happen before the deletion.
		// Update b.currentAction manually here to make it an insertion only.
		b.currentAction.ins.bounds.tail.col--
		text := b.currentAction.ins.text
		b.currentAction.ins.text = text[:len(text)-1]
	}
}

func (b *Buffer) left() {
	b.dirtyLines(b.dot.head.row, b.dot.tail.row+1)
	a := b.prevaddress(b.dot.head)
	b.dot.head, b.dot.tail = a, a
	b.dirtyLine(b.dot.head.row) // new dot may be in a higher row
}

func (b *Buffer) right() {
	b.dirtyLines(b.dot.head.row, b.dot.tail.row+1)
	a := b.nextaddress(b.dot.tail)
	b.dot.head, b.dot.tail = a, a
	b.dirtyLine(b.dot.head.row) // new dot may be in a lower row
}

func (b *Buffer) newline() {
	b.loadBytes([]byte("\n"), true)
	b.dot.head = b.dot.tail
}
