package editor

import (
	"image"
	"unicode"
	"unicode/utf8"

	"sigint.ca/graphics/editor/internal/text"

	"golang.org/x/mobile/event/key"
)

func (ed *Editor) handleKeyEvent(e key.Event) {
	if e.Direction == key.DirRelease {
		// ignore key up events
		return
	}

	ed.dirty = true

	// prepare for a change in the editor's history.
	ed.initTransformation()

	switch {
	case e.Code == key.CodeEscape:
		if ed.dot.IsEmpty() {
			ed.dot.From.Col -= utf8.RuneCountInString(ed.uncommitted.Post.Text)
		} else {
			ed.dot = ed.buf.ClearSel(ed.dot)
		}
		ed.commitTransformation()

	case e.Code == key.CodeDeleteBackspace:
		if ed.uncommitted.Post.Text != "" {
			// trim the final uncommitted character
			_, rSize := utf8.DecodeLastRuneInString(ed.uncommitted.Post.Text)
			newSize := len(ed.uncommitted.Post.Text) - rSize
			ed.uncommitted.Post.Text = ed.uncommitted.Post.Text[:newSize]
		} else {
			// ed.uncommitted.Pre.Sel.From must also include the rune preceding dot
			ed.uncommitted.Pre.Sel.From = ed.buf.PrevAddress(ed.uncommitted.Pre.Sel.From)
			ed.uncommitted.Pre.Text = ed.buf.GetSel(ed.uncommitted.Pre.Sel)
		}
		ed.dot.From = ed.buf.PrevAddress(ed.dot.From)
		ed.dot = ed.buf.ClearSel(ed.dot)
		ed.commitTransformation()

	case e.Code == key.CodeReturnEnter:
		ed.putString("\n")
		ed.dot.From = ed.dot.To
		ed.uncommitted.Post.Text += "\n"
		ed.commitTransformation()

	case e.Code == key.CodeUpArrow:
		ed.scroll(image.Pt(0, 18*ed.font.height))
		ed.commitTransformation()

	case e.Code == key.CodeDownArrow:
		ed.scroll(image.Pt(0, -18*ed.font.height))
		ed.commitTransformation()

	case e.Code == key.CodeLeftArrow:
		ed.commitTransformation()
		a := ed.buf.PrevAddress(ed.dot.From)
		ed.dot.From, ed.dot.To = a, a

	case e.Code == key.CodeRightArrow:
		ed.commitTransformation()
		a := ed.buf.NextAddress(ed.dot.To)
		ed.dot.From, ed.dot.To = a, a

	case e.Modifiers == key.ModMeta && e.Code == key.CodeC:
		ed.commitTransformation()
		ed.snarf()

	case e.Modifiers == key.ModMeta && e.Code == key.CodeV:
		ed.paste()
		ed.commitTransformation()

	case e.Modifiers == key.ModMeta && e.Code == key.CodeX:
		ed.snarf()
		ed.dot = ed.buf.ClearSel(ed.dot)
		ed.commitTransformation()

	case e.Modifiers == key.ModMeta && e.Code == key.CodeA:
		ed.commitTransformation()
		last := len(ed.buf.Lines) - 1
		ed.dot.From = text.Address{0, 0}
		ed.dot.To = text.Address{last, ed.buf.Lines[last].RuneCount()}

	case e.Modifiers == key.ModMeta|key.ModShift && e.Code == key.CodeZ:
		ed.commitTransformation()
		ed.redo()

	case e.Modifiers == key.ModMeta && e.Code == key.CodeZ:
		ed.commitTransformation()
		ed.undo()

	default:
		if isGraphic(e.Rune) && e.Modifiers&key.ModMeta == 0 {
			s := string(e.Rune)
			ed.uncommitted.Post.Text += s
			ed.putString(s)
			ed.dot.From = ed.dot.To

			// don't commit - history is not updated for each rune of input
		}
	}
}

func isGraphic(r rune) bool {
	switch r {
	case '\t':
		return true
	}
	return unicode.IsGraphic(r)
}
