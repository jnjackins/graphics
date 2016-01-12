package editor

import "sigint.ca/graphics/editor/internal/hist"

func (ed *Editor) undo() {
	ch, ok := ed.history.Undo()
	if !ok {
		return
	}
	ed.dot = ch.Sel
	ed.putString(ch.Text)
}

func (ed *Editor) redo() {
	ch, ok := ed.history.Redo()
	if !ok {
		return
	}
	ed.dot = ch.Sel
	ed.putString(ch.Text)
}

func (ed *Editor) initTransformation() {
	if ed.uncommitted == nil {
		ed.uncommitted = new(hist.Transformation)
		ed.uncommitted.Pre = hist.Chunk{
			Sel:  ed.dot,
			Text: ed.buf.GetSel(ed.dot),
		}
		// this sets up Post to match Pre in the case that Pre is empty,
		// so that no-op transformations are caught and discarded
		ed.uncommitted.Post.Sel = ed.dot
	}
}

func (ed *Editor) commitTransformation() {
	if ed.uncommitted.Post.Text == "" {
		ed.uncommitted.Post.Text = ed.buf.GetSel(ed.dot)
		ed.uncommitted.Post.Sel = ed.dot
	} else {
		ed.uncommitted.Post.Sel = ed.uncommitted.Pre.Sel
		for range ed.uncommitted.Post.Text {
			// TODO: this is inefficient
			ed.uncommitted.Post.Sel.To = ed.buf.NextAddress(ed.uncommitted.Post.Sel.To)
		}
	}

	if ed.uncommitted.Pre != ed.uncommitted.Post {
		ed.history.Current().Pre = ed.uncommitted.Pre
		ed.history.Current().Post = ed.uncommitted.Post
		ed.history.Commit()
	}
	ed.uncommitted = nil
}
