package editor

import (
	"sigint.ca/graphics/editor/address"
	"sigint.ca/graphics/editor/internal/hist"
)

// Load replaces the contents of the Editor's text buffer with s, and resets the Editor's history.
func (ed *Editor) Load(s []byte) {
	last := len(ed.Buffer.Lines) - 1
	all := address.Selection{To: address.Simple{last, ed.Buffer.Lines[last].RuneCount()}}
	ed.Dot = ed.Buffer.ClearSel(all)
	ed.Dot.To = ed.Buffer.InsertString(address.Simple{}, string(s))
	ed.history = new(hist.History)
	ed.uncommitted = nil
	ed.dirty = true
}

// FindNext searches for s in the Editor's text buffer, and selects the first match
// starting from the current selection, possibly wrapping around to the beginning
// of the buffer. If there are no matches, the selection is unchanged.
func (ed *Editor) FindNext(s string) (address.Selection, bool) {
	if sel, ok := ed.Buffer.Find(ed.Dot.To, s); ok {
		ed.Dot = sel
		ed.autoscroll()
		ed.dirty = true
		return ed.Dot, true
	}
	return ed.Dot, false
}

// JumpTo sets the selection to the specified address, as define in sam(1).
func (ed *Editor) JumpTo(addr string) bool {
	if sel, ok := ed.Buffer.JumpTo(ed.Dot.To, addr); ok {
		ed.Dot = sel
		ed.autoscroll()
		ed.dirty = true
		return true
	}
	return false
}

// putString replaces the current selection with s, and selects
// the results.
func (ed *Editor) putString(s string) {
	ed.Buffer.ClearSel(ed.Dot)
	addr := ed.Buffer.InsertString(ed.Dot.From, s)
	ed.Dot.To = addr
}
