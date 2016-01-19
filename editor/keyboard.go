package editor

import (
	"image"
	"unicode"
	"unicode/utf8"

	"golang.org/x/mobile/event/key"
)

func (ed *Editor) handleKeyEvent(e key.Event) {
	if e.Direction == key.DirRelease {
		// ignore key up events
		return
	}

	ed.dirty = true

	// prepare for a change in the editor's history
	ed.initTransformation()

	// handle a single typed rune
	if isGraphic(e.Rune) && e.Modifiers&key.ModMeta == 0 {
		ed.uncommitted.Post.Text += string(e.Rune)
		ed.input(e.Rune)

		// exit early - history isn't updated for each keystroke
		return
	}

	// handle all other key events
	switch {
	case e.Code == key.CodeDeleteBackspace:
		ed.backspace()
	case e.Code == key.CodeReturnEnter:
		ed.newline()
	case e.Code == key.CodeUpArrow:
		ed.scroll(image.Pt(0, 18*ed.font.height))
	case e.Code == key.CodeLeftArrow:
		ed.left()
	case e.Code == key.CodeRightArrow:
		ed.right()
	case e.Code == key.CodeDownArrow:
		ed.scroll(image.Pt(0, -18*ed.font.height))
	case e.Modifiers == key.ModMeta && e.Code == key.CodeC:
		ed.snarf()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeV:
		ed.paste()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeX:
		ed.snarf()
		ed.dot = ed.buf.ClearSel(ed.dot)
	case e.Modifiers == key.ModMeta && e.Code == key.CodeA:
		ed.selAll()
	case e.Modifiers == key.ModMeta|key.ModShift && e.Code == key.CodeZ:
		// if there is a new transformation, allow it to be committed before trying to redo
		defer ed.redo()
	case e.Modifiers == key.ModMeta && e.Code == key.CodeZ:
		// if there is a new transformation, allow it to be committed before trying to undo
		defer ed.undo()
	}

	// commit any change to the editor's history
	ed.commitTransformation()
}

func isGraphic(r rune) bool {
	switch r {
	case '\t':
		return true
	}
	return unicode.IsGraphic(r)
}

func (ed *Editor) input(r rune) {
	ed.putString(string(r))
	ed.dot.From = ed.dot.To
}

func (ed *Editor) backspace() {
	// special history handling specific to backspace
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
}

func (ed *Editor) left() {
	// commit early: the transformation should not be affected
	// by the new dot after moving performing this action
	ed.commitTransformation()

	a := ed.buf.PrevAddress(ed.dot.From)
	ed.dot.From, ed.dot.To = a, a
}

func (ed *Editor) right() {
	// commit early: the transformation should not be affected
	// by the new dot after moving performing this action
	ed.commitTransformation()

	a := ed.buf.NextAddress(ed.dot.To)
	ed.dot.From, ed.dot.To = a, a
}

func (ed *Editor) newline() {
	// special history handling specific to newline
	ed.uncommitted.Post.Text += "\n"

	ed.putString("\n")
	ed.dot.From = ed.dot.To
}
