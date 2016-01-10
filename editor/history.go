package editor

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
