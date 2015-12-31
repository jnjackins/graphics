package text

// each editor action is a deletion followed by an insertion.
type action struct {
	deletion   *change
	insertion  *change
	prev, next *action
}

type change struct {
	bounds selection
	text   []byte
}

func (b *Buffer) redo() {
	if b.lastAction.next != nil {
		a := b.lastAction.next
		b.lastAction = b.lastAction.next

		if a.deletion != nil {
			b.dot = a.deletion.bounds
			b.deleteSel(false)
		}
		if a.insertion != nil {
			b.dot = selection{a.insertion.bounds.head, a.insertion.bounds.head}
			b.loadBytes(a.insertion.text, false)
		}

		b.dirtyLines(0, len(b.lines))
		b.autoScroll()
	}
}

func (b *Buffer) undo() {
	if b.lastAction.prev != nil {
		a := b.lastAction
		b.lastAction = b.lastAction.prev

		if a.insertion != nil {
			b.dot = a.insertion.bounds
			b.deleteSel(false)
		}
		if a.deletion != nil {
			b.dot = selection{a.deletion.bounds.head, a.deletion.bounds.head}
			b.loadBytes(a.deletion.text, false)
		}

		b.dirtyLines(0, len(b.lines))
		b.autoScroll()
	}
}

// commitAction finalizes b.currentAction and adds it to the list, to become the
// new b.lastAction.
func (b *Buffer) commitAction() bool {
	if b.currentAction.deletion == nil && b.currentAction.insertion == nil {
		return false
	}
	if b.lastAction != nil {
		b.lastAction.next = b.currentAction
	}
	b.currentAction.prev = b.lastAction
	b.currentAction.next = nil
	b.lastAction = b.currentAction
	b.currentAction = new(action)
	return true
}
