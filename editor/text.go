package editor

import (
	"sigint.ca/graphics/editor/internal/address"
	"sigint.ca/graphics/editor/internal/hist"
)

// Load replaces the contents of the Editor's text buffer with s, and resets the Editor's history.
func (ed *Editor) Load(s []byte) {
	last := len(ed.buf.Lines) - 1
	all := address.Selection{To: address.Simple{last, ed.buf.Lines[last].RuneCount()}}
	ed.dot = ed.buf.ClearSel(all)
	ed.buf.InsertString(address.Simple{0, 0}, string(s))
	ed.history = new(hist.History)
	ed.uncommitted = nil
	ed.dirty = true
}

// Contents returns the contents of the Editor's text buffer.
func (ed *Editor) Contents() []byte {
	return ed.buf.Contents()
}

// FindNext searches for s in the Editor's text buffer, and selects the first match
// starting from the current selection, possibly wrapping around to the beginning
// of the buffer. If there are no matches, the selection is unchanged.
// TODO: go back to dedicated implementation
func (ed *Editor) FindNext(s string) bool {
	if sel, ok := ed.buf.Find(ed.dot.To, s); ok {
		ed.dot = sel
		ed.autoscroll()
		ed.dirty = true
		return true
	}
	return false
}

// JumpTo sets the selection to the specified address, as define in sam(1).
func (ed *Editor) JumpTo(addr string) bool {
	if sel, ok := ed.buf.JumpTo(ed.dot.To, addr); ok {
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
	ed.buf.ClearSel(ed.dot)
	addr := ed.buf.InsertString(ed.dot.From, s)
	ed.dot.To = addr
}
