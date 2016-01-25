package editor

// putString replaces the current selection with s, and selects
// the results.
func (ed *Editor) putString(s string) {
	ed.buf.ClearSel(ed.dot)
	addr := ed.buf.InsertString(ed.dot.From, s)
	ed.dot.To = addr
}
