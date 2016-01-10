package editor

import (
	"image"
	"unicode"

	"sigint.ca/graphics/editor/internal/hist"
	"sigint.ca/graphics/editor/internal/text"

	"golang.org/x/mobile/event/key"
)

// TODO: clean up history processing somehow
func (ed *Editor) handleKeyEvent(e key.Event) {
	var preChunk, postChunk hist.Chunk

	// prepare the first part of the history transformation
	if ed.dot.IsEmpty() {
		// adjust for uncommitted input
		row, col := ed.dot.From.Row, ed.dot.From.Col-len(ed.uncommitted)
		preChunk.Sel = text.Sel(row, col, row, col)
	} else {
		preChunk.Sel = ed.dot
		preChunk.Text = ed.buf.GetSel(preChunk.Sel)
	}
	postChunk = preChunk

	// handle a single typed rune
	if isGraphic(e.Rune) && e.Modifiers&key.ModMeta == 0 {
		ed.input(e.Rune)
		return // don't commit history on each single-rune input
	}

	uncommitted := ed.uncommitted
	ed.uncommitted = nil

	// handle all other key events
	switch {
	case e.Code == key.CodeDeleteBackspace:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted[:len(uncommitted)-1])
			postChunk.Sel.To.Col += len(uncommitted) - 1
		} else {
			preChunk.Sel.From.Col--                     // select the character before the cursor
			preChunk.Text = ed.buf.GetSel(preChunk.Sel) // TODO: avoid this double calculation (use defer?)
			postChunk.Sel = text.Selection{preChunk.Sel.From, preChunk.Sel.From}
		}

		ed.backspace()
	case e.Code == key.CodeReturnEnter:
		if len(uncommitted) > 0 {
			postChunk.Text = string(append(uncommitted, '\n'))
		} else {
			postChunk.Text = "\n"
		}
		postChunk.Sel.To.Row++
		postChunk.Sel.To.Col = 0

		ed.newline()
	case e.Code == key.CodeUpArrow:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		ed.scroll(image.Pt(0, -18*ed.lineHeight))
	case e.Code == key.CodeLeftArrow:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		ed.left()
	case e.Code == key.CodeRightArrow:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		ed.right()
	case e.Code == key.CodeDownArrow:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		ed.scroll(image.Pt(0, 18*ed.lineHeight))
	case e.Modifiers == key.ModMeta && e.Code == key.CodeC:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		ed.snarf()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeV:

		ed.paste()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeX:
		ed.snarf()
		ed.buf.ClearSel(ed.dot)
	case e.Modifiers == key.ModMeta && e.Code == key.CodeA:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
		}

		ed.selAll()
	case e.Modifiers == key.ModMeta|key.ModShift && e.Code == key.CodeZ:
		ed.redo()
		return
	case e.Modifiers == key.ModMeta && e.Code == key.CodeZ:
		if len(uncommitted) > 0 {
			postChunk.Text = string(uncommitted)
			postChunk.Sel.To.Col += len(uncommitted)
			ed.history.Current().Pre = preChunk
			ed.history.Current().Post = postChunk
			ed.history.Commit()
		}

		ed.undo()
		return
	}

	if preChunk != postChunk {
		ed.history.Current().Pre = preChunk
		ed.history.Current().Post = postChunk
		ed.history.Commit()
	}
}

func isGraphic(r rune) bool {
	switch r {
	case '\t':
		return true
	}
	return unicode.IsGraphic(r)
}

func (ed *Editor) input(r rune) {
	ed.uncommitted = append(ed.uncommitted, r) // to be committed to history later
	ed.putString(string(r))
	ed.dot.From = ed.dot.To
}

func (ed *Editor) backspace() {
	ed.dot.From = ed.buf.PrevAddress(ed.dot.From)
	ed.dot = ed.buf.ClearSel(ed.dot)
}

func (ed *Editor) left() {
	a := ed.buf.PrevAddress(ed.dot.From)
	ed.dot.From, ed.dot.To = a, a
}

func (ed *Editor) right() {
	a := ed.buf.NextAddress(ed.dot.To)
	ed.dot.From, ed.dot.To = a, a
}

func (ed *Editor) newline() {
	ed.putString("\n")
	ed.dot.From = ed.dot.To
}
