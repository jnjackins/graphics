package editor

import (
	"sigint.ca/graphics/editor/address"
	"sigint.ca/graphics/editor/internal/hist"
)

// Load replaces the contents of the Editor's text buffer with s, and resets the Editor's history.
func (ed *Editor) Load(s []byte) {
	last := len(ed.buffer.Lines) - 1
	all := address.Selection{To: address.Simple{last, ed.buffer.Lines[last].RuneCount()}}
	ed.dot = ed.buffer.ClearSel(all)
	ed.dot.To = ed.buffer.InsertString(address.Simple{}, string(s))
	ed.history = new(hist.History)
	ed.uncommitted = nil
	ed.dirty = true
}

// Contents returns the entire contents of the editor.
func (ed *Editor) Contents() []byte {
	return ed.buffer.Contents()
}

// Replace replaces the current selection with s, updating the Editor's history.
func (ed *Editor) Replace(s string) {
	ed.initTransformation()
	ed.putString(s)
	ed.commitTransformation()
	ed.autoscroll()
	ed.dirty = true
}

func (ed *Editor) GetDot() address.Selection {
	return ed.dot
}

func (ed *Editor) SetDot(a address.Selection) {
	if a != ed.dot {
		ed.initTransformation()
		ed.commitTransformation()
		ed.dot = a
	}
}

func (ed *Editor) GetDotContents() string {
	return ed.buffer.GetSel(ed.dot)
}

func (ed *Editor) LastAddress() address.Simple {
	return ed.buffer.LastAddress()
}

// FindNext searches for s in the Editor's text buffer, and selects the first match
// starting from the current selection, possibly wrapping around to the beginning
// of the buffer. If there are no matches, the selection is unchanged.
func (ed *Editor) FindNext(s string) (address.Selection, bool) {
	if sel, ok := ed.buffer.Find(ed.dot.To, s); ok {
		ed.dot = sel
		ed.autoscroll()
		ed.dirty = true
		return ed.dot, true
	}
	return ed.dot, false
}

// JumpTo sets the selection to the specified address, as define in sam(1).
func (ed *Editor) JumpTo(addr string) bool {
	if sel, ok := ed.buffer.JumpTo(ed.dot.To, addr); ok {
		ed.dot = sel
		ed.autoscroll()
		ed.dirty = true
		return true
	}
	return false
}

// putString replaces the current selection with s, and selects
// the results.
func (ed *Editor) putString(s string) {
	ed.buffer.ClearSel(ed.dot)
	addr := ed.buffer.InsertString(ed.dot.From, s)
	ed.dot.To = addr
}
