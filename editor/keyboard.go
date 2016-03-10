package editor

import (
	"image"
	"unicode"
	"unicode/utf8"

	"sigint.ca/graphics/editor/address"

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
		if ed.Dot.IsEmpty() {
			ed.Dot.From.Col -= utf8.RuneCountInString(ed.uncommitted.Post.Text)
		} else {
			ed.Dot = ed.Buffer.ClearSel(ed.Dot)
		}
		ed.commitTransformation()

	case e.Code == key.CodeDeleteBackspace, e.Modifiers == key.ModControl && e.Code == key.CodeH:
		ed.backspace(1)
		ed.commitTransformation()

	// word kill
	case e.Modifiers == key.ModControl && e.Code == key.CodeW:
		if ed.Dot.From.Col == 0 {
			ed.backspace(1)
		} else {
			line := ed.Buffer.Lines[ed.Dot.From.Row].Runes()
			var n, dot int
			for dot = ed.Dot.From.Col; dot > 0 && !isWordChar(line[dot-1]); dot-- {
				n++
			}
			for ; dot > 0 && isWordChar(line[dot-1]); dot-- {
				n++
			}
			ed.backspace(n)
		}
		ed.commitTransformation()

	// line kill
	case e.Modifiers == key.ModControl && e.Code == key.CodeU:
		if ed.Dot.From.Col == 0 {
			ed.backspace(1)
		} else {
			ed.backspace(ed.Dot.From.Col)
		}
		ed.commitTransformation()

	case e.Code == key.CodeReturnEnter, e.Modifiers == key.ModControl && e.Code == key.CodeJ:
		prefix := ""
		if ed.opts.AutoIndent {
			prefix = ed.getIndentation()
		}
		ed.putString("\n" + prefix)
		ed.Dot.From = ed.Dot.To
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
		a := ed.Buffer.PrevSimple(ed.Dot.From)
		ed.Dot.From, ed.Dot.To = a, a

	case e.Code == key.CodeRightArrow:
		ed.commitTransformation()
		a := ed.Buffer.NextSimple(ed.Dot.To)
		ed.Dot.From, ed.Dot.To = a, a

	case e.Modifiers == key.ModControl && e.Code == key.CodeA:
		ed.commitTransformation()
		ed.Dot.From.Col = 0
		ed.Dot.To = ed.Dot.From

	case e.Modifiers == key.ModControl && e.Code == key.CodeE:
		ed.commitTransformation()
		ed.Dot.From.Col = ed.Buffer.Lines[ed.Dot.From.Row].RuneCount()
		ed.Dot.To = ed.Dot.From

	case e.Modifiers == key.ModMeta && e.Code == key.CodeC:
		ed.commitTransformation()
		ed.snarf()

	case e.Modifiers == key.ModMeta && e.Code == key.CodeV:
		ed.paste()
		ed.commitTransformation()

	case e.Modifiers == key.ModMeta && e.Code == key.CodeX:
		ed.snarf()
		ed.Dot = ed.Buffer.ClearSel(ed.Dot)
		ed.commitTransformation()

	case e.Modifiers == key.ModMeta && e.Code == key.CodeA:
		ed.commitTransformation()
		last := len(ed.Buffer.Lines) - 1
		ed.Dot.From = address.Simple{0, 0}
		ed.Dot.To = address.Simple{last, ed.Buffer.Lines[last].RuneCount()}

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
			ed.Dot.From = ed.Dot.To

			// don't commit - history is not updated for each rune of input
		}
	}
}

func (ed *Editor) backspace(n int) {
	// first, trim from uncommitted characters
	for n > 0 && ed.uncommitted.Post.Text != "" {
		_, rSize := utf8.DecodeLastRuneInString(ed.uncommitted.Post.Text)
		newSize := len(ed.uncommitted.Post.Text) - rSize
		ed.uncommitted.Post.Text = ed.uncommitted.Post.Text[:newSize]
		ed.Dot.From = ed.Buffer.PrevSimple(ed.Dot.From)
		n--
	}
	for n > 0 {
		// ed.uncommitted.Pre.Sel.From must also include the rune preceding dot
		ed.uncommitted.Pre.Sel.From = ed.Buffer.PrevSimple(ed.uncommitted.Pre.Sel.From)
		ed.uncommitted.Pre.Text = ed.Buffer.GetSel(ed.uncommitted.Pre.Sel)
		ed.Dot.From = ed.Buffer.PrevSimple(ed.Dot.From)
		n--
	}
	ed.Dot = ed.Buffer.ClearSel(ed.Dot)
}

func (ed *Editor) getIndentation() string {
	prefix := make([]rune, 0)
	line := ed.Buffer.Lines[ed.Dot.From.Row].String()
	for _, r := range line {
		if unicode.IsSpace(r) {
			prefix = append(prefix, r)
		} else {
			break
		}
	}
	return string(prefix)
}

func isWordChar(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c) || c == '_'
}

func isGraphic(r rune) bool {
	switch r {
	case '\t':
		return true
	}
	return unicode.IsGraphic(r)
}
