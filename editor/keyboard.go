package editor

import (
	"image"
	"unicode"
	"unicode/utf8"

	"sigint.ca/graphics/editor/internal/address"

	"golang.org/x/mobile/event/key"
)

// SendKeyEvent sends a key event to be interpreted by the Editor.
func (ed *Editor) SendKeyEvent(e key.Event) {
	ed.handleKeyEvent(e)
}

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
			ed.uncommitted.Pre.Sel.From = ed.buf.PrevSimple(ed.uncommitted.Pre.Sel.From)
			ed.uncommitted.Pre.Text = ed.buf.GetSel(ed.uncommitted.Pre.Sel)
		}
		ed.dot.From = ed.buf.PrevSimple(ed.dot.From)
		ed.dot = ed.buf.ClearSel(ed.dot)
		ed.commitTransformation()

	case e.Code == key.CodeReturnEnter:
		prefix := ""
		if ed.opts.AutoIndent {
			prefix = ed.getIndentation()
		}
		ed.putString("\n" + prefix)
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
		a := ed.buf.PrevSimple(ed.dot.From)
		ed.dot.From, ed.dot.To = a, a

	case e.Code == key.CodeRightArrow:
		ed.commitTransformation()
		a := ed.buf.NextSimple(ed.dot.To)
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
		ed.dot.From = address.Simple{0, 0}
		ed.dot.To = address.Simple{last, ed.buf.Lines[last].RuneCount()}

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

func (ed *Editor) getIndentation() string {
	prefix := make([]rune, 0)
	line := ed.buf.Lines[ed.dot.From.Row].String()
	for _, r := range line {
		if unicode.IsSpace(r) {
			prefix = append(prefix, r)
		} else {
			break
		}
	}
	return string(prefix)
}

func isGraphic(r rune) bool {
	switch r {
	case '\t':
		return true
	}
	return unicode.IsGraphic(r)
}
