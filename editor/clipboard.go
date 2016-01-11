package editor

func (ed *Editor) snarf() {
	ed.clipboard.Put([]byte(ed.buf.GetSel(ed.dot)))
}

func (ed *Editor) paste() {
	s, err := ed.clipboard.Get()
	if err == nil {
		ed.putString(string(s))
	}
}
