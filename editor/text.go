package editor

import "sigint.ca/graphics/editor/internal/text"

func (ed *Editor) putBytes(s []byte) {
	ed.buf.ClearSel(ed.dot)
	addr := ed.buf.InsertBytes(ed.dot.From, s)
	ed.dot.To = addr
}

func (ed *Editor) putString(s string) {
	ed.buf.ClearSel(ed.dot)
	addr := ed.buf.InsertString(ed.dot.From, s)
	ed.dot.To = addr
}

func (ed *Editor) selAll() {
	last := len(ed.buf.Lines) - 1
	ed.dot.From = text.Address{0, 0}
	ed.dot.To = text.Address{last, ed.buf.Lines[last].RuneCount()}
}
