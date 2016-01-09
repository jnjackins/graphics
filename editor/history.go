package editor

func (b *Buffer) undo() {
	ch, ok := b.history.Undo()
	if !ok {
		return
	}
	b.sel(ch.Sel.From, ch.Sel.To)
	b.loadBytes([]byte(ch.Text))
}

func (b *Buffer) redo() {
	ch, ok := b.history.Redo()
	if !ok {
		return
	}
	b.sel(ch.Sel.From, ch.Sel.To)
	b.loadBytes([]byte(ch.Text))
}
