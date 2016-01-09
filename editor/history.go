package editor

func (ed *Editor) undo() {
	ch, ok := ed.history.Undo()
	if !ok {
		return
	}
	ed.sel(ch.Sel.From, ch.Sel.To)
	ed.loadBytes([]byte(ch.Text))
}

func (ed *Editor) redo() {
	ch, ok := ed.history.Redo()
	if !ok {
		return
	}
	ed.sel(ch.Sel.From, ch.Sel.To)
	ed.loadBytes([]byte(ch.Text))
}
