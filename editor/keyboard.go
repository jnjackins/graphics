package editor

import (
	"image"
	"unicode"

	"sigint.ca/graphics/editor/internal/hist"
	"sigint.ca/graphics/editor/internal/text"

	"golang.org/x/mobile/event/key"
)

// TODO: clean up history processing somehow
func (b *Buffer) handleKeyEvent(e key.Event) {
	var preChunk, postChunk hist.Chunk

	// prepare the first part of the history transformation
	if b.dot.IsEmpty() {
		// adjust for uncommitted input
		row, col := b.dot.From.Row, b.dot.From.Col-len(b.uncommitted)
		preChunk.Sel = text.Sel(row, col, row, col)
	} else {
		preChunk.Sel = b.dot
		preChunk.Text = b.contents(b.dot)
	}
	postChunk = preChunk

	// handle a single typed rune
	if isGraphic(e.Rune) && e.Modifiers&key.ModMeta == 0 {
		b.input(e.Rune)
		return // don't commit history on each single-rune input
	}

	uncommitted := b.uncommitted
	b.uncommitted = nil

	// handle all other key events
	switch {
	case e.Code == key.CodeDeleteBackspace:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted[:len(uncommitted)-1])
			postChunk.Sel.To.Col += len(uncommitted) - 1
		} else {
			preChunk.Sel.From.Col-- // select the character before the cursor
			postChunk.Sel = text.Selection{preChunk.Sel.From, preChunk.Sel.From}
		}

		b.backspace()
	case e.Code == key.CodeReturnEnter:
		if len(uncommitted) > 0 {
			postChunk.Text = string(append(uncommitted, '\n'))
		} else {
			postChunk.Text = "\n"
		}
		postChunk.Sel.To.Row++
		postChunk.Sel.To.Col = 0

		b.newline()
	case e.Code == key.CodeUpArrow:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		b.scroll(image.Pt(0, -18*b.lineHeight))
	case e.Code == key.CodeLeftArrow:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		b.left()
	case e.Code == key.CodeRightArrow:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		b.right()
	case e.Code == key.CodeDownArrow:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		b.scroll(image.Pt(0, 18*b.lineHeight))
	case e.Modifiers == key.ModMeta && e.Code == key.CodeC:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		b.snarf()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeV:

		b.paste()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeX:
		b.snarf()
		b.dot = b.clear(b.dot)
	case e.Modifiers == key.ModMeta && e.Code == key.CodeA:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		b.selAll()
	case e.Modifiers == key.ModMeta|key.ModShift && e.Code == key.CodeZ:
		b.redo()
		return
	case e.Modifiers == key.ModMeta && e.Code == key.CodeZ:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
			b.history.Current().Pre = preChunk
			b.history.Current().Post = postChunk
			b.history.Commit()
		}

		b.undo()
		return
	}

	if preChunk != postChunk {
		b.history.Current().Pre = preChunk
		b.history.Current().Post = postChunk
		b.history.Commit()
	}
}

func isGraphic(r rune) bool {
	switch r {
	case '\t':
		return true
	}
	return unicode.IsGraphic(r)
}

func (b *Buffer) input(r rune) {
	b.uncommitted = append(b.uncommitted, r) // to be committed to history later
	b.loadRune(r)
	b.dot.From = b.dot.To
}

func (b *Buffer) backspace() {
	b.dot.From = b.prevAddress(b.dot.From)
	b.dot = b.clear(b.dot)
}

func (b *Buffer) left() {
	b.dirtyLines(b.dot.From.Row, b.dot.To.Row+1)
	a := b.prevAddress(b.dot.From)
	b.dot.From, b.dot.To = a, a
	b.dirtyLine(b.dot.From.Row) // new dot may be in a higher row
}

func (b *Buffer) right() {
	b.dirtyLines(b.dot.From.Row, b.dot.To.Row+1)
	a := b.nextAddress(b.dot.To)
	b.dot.From, b.dot.To = a, a
	b.dirtyLine(b.dot.From.Row) // new dot may be in a lower row
}

func (b *Buffer) newline() {
	b.loadBytes([]byte("\n"))
	b.dot.From = b.dot.To
}
