package editor

import (
	"sigint.ca/graphics/editor/internal/address"
	"sigint.ca/graphics/editor/internal/hist"
)

func (ed *Editor) undo() {
	ch, ok := ed.history.Undo()
	if !ok {
		return
	}
	ed.dot = ch.Sel
	ed.putString(ch.Text)
	ed.dirty = true
}

func (ed *Editor) redo() {
	ch, ok := ed.history.Redo()
	if !ok {
		return
	}
	ed.dot = ch.Sel
	ed.putString(ch.Text)
	ed.dirty = true
}

// initTransformation sets uncommitted.Pre to the current selection.
func (ed *Editor) initTransformation() {
	if ed.uncommitted == nil {
		ed.uncommitted = new(hist.Transformation)
		ed.uncommitted.Pre = hist.Chunk{
			Sel:  ed.dot,
			Text: ed.buf.GetSel(ed.dot),
		}
	}
}

// commitTransformation updates uncommitted.Post by either:
//
// a) setting it to the current selection, if uncommitted.Post.Text is empty
// b) updating uncommitted.Post.Sel to match uncommitted.Post.Text if it has been populated
//
// The transformation, if not a no-op, is then committed to history.
func (ed *Editor) commitTransformation() {
	if ed.uncommitted == nil {
		return
	}

	if ed.uncommitted.Post.Text == "" {
		ed.uncommitted.Post.Text = ed.buf.GetSel(ed.dot)
		ed.uncommitted.Post.Sel = ed.dot
	} else {
		ed.uncommitted.Post.Sel = address.Selection{
			ed.uncommitted.Pre.Sel.From,
			ed.uncommitted.Pre.Sel.From,
		}
		for range ed.uncommitted.Post.Text {
			// TODO: use measure and add?
			ed.uncommitted.Post.Sel.To = ed.buf.NextSimple(ed.uncommitted.Post.Sel.To)
		}
	}

	if ed.uncommitted.Pre != ed.uncommitted.Post {
		ed.history.Current().Pre = ed.uncommitted.Pre
		ed.history.Current().Post = ed.uncommitted.Post
		ed.history.Commit()
	}
	ed.uncommitted = nil
}
