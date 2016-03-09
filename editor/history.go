package editor

import (
	"sigint.ca/graphics/editor/address"
	"sigint.ca/graphics/editor/internal/hist"
)

// SendUndo attempts to apply the Editor's previous history state, if it exists.
func (ed *Editor) SendUndo() {
	// commit any lingering uncommitted changes
	ed.initTransformation()
	ed.commitTransformation()
	ed.undo()
}

// SendRedo attempts to apply the Editor's next history state, if it exists.
func (ed *Editor) SendRedo() {
	// commit any lingering uncommitted changes
	ed.initTransformation()
	ed.commitTransformation()
	ed.redo()
}

// CanUndo reports whether the Editor has a previous history state which can be applied.
func (ed *Editor) CanUndo() bool {
	return ed.history.CanUndo() ||
		ed.uncommitted != nil && len(ed.uncommitted.Post.Text) > 0
}

// CanRedo reports whether the Editor has a following history state which can be applied.
func (ed *Editor) CanRedo() bool {
	return ed.history.CanRedo()
}

// SetSaved instructs the Editor that the current contents should be
// considered saved. After calling SetSaved, the client can call
// Saved to see if the Editor has unsaved content.
func (ed *Editor) SetSaved() {
	if ed.uncommitted != nil {
		ed.commitTransformation()
	}
	ed.savePoint = ed.history.Current()
}

// Saved reports whether the Editor has been modified since the last
// time SetSaved was called.
func (ed *Editor) Saved() bool {
	return ed.history.Current() == ed.savePoint &&
		(ed.uncommitted == nil || ed.uncommitted.Post.Text == "")
}

func (ed *Editor) undo() {
	ch, ok := ed.history.Undo()
	if !ok {
		return
	}
	ed.Dot = ch.Sel
	ed.putString(ch.Text)
	ed.dirty = true
}

func (ed *Editor) redo() {
	ch, ok := ed.history.Redo()
	if !ok {
		return
	}
	ed.Dot = ch.Sel
	ed.putString(ch.Text)
	ed.dirty = true
}

// initTransformation sets uncommitted.Pre to the current selection.
func (ed *Editor) initTransformation() {
	if ed.uncommitted == nil {
		ed.uncommitted = new(hist.Transformation)
		ed.uncommitted.Pre = hist.Chunk{
			Sel:  ed.Dot,
			Text: ed.Buffer.GetSel(ed.Dot),
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
		ed.uncommitted.Post.Text = ed.Buffer.GetSel(ed.Dot)
		ed.uncommitted.Post.Sel = ed.Dot
	} else {
		ed.uncommitted.Post.Sel = address.Selection{
			ed.uncommitted.Pre.Sel.From,
			ed.uncommitted.Pre.Sel.From,
		}
		for range ed.uncommitted.Post.Text {
			// TODO: use measure and add?
			ed.uncommitted.Post.Sel.To = ed.Buffer.NextSimple(ed.uncommitted.Post.Sel.To)
		}
	}

	if ed.uncommitted.Pre != ed.uncommitted.Post {
		ed.history.Current().Pre = ed.uncommitted.Pre
		ed.history.Current().Post = ed.uncommitted.Post
		ed.history.Commit()
	}
	ed.uncommitted = nil
}
