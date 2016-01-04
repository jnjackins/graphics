package text

// each editor action is a deletion followed by an insertion.
type action struct {
	del, ins   *change
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

		if a.del != nil {
			b.dot = a.del.bounds
			b.deleteSel(false)
		}
		if a.ins != nil {
			b.dot = selection{a.ins.bounds.head, a.ins.bounds.head}
			b.loadBytes(a.ins.text, false)
		}

		b.dirtyLines(0, len(b.lines))
		b.autoScroll()
	}
}

func (b *Buffer) undo() {
	if b.lastAction.prev != nil {
		a := b.lastAction
		b.lastAction = b.lastAction.prev

		if a.ins != nil {
			b.dot = a.ins.bounds
			b.deleteSel(false)
		}
		if a.del != nil {
			b.dot = selection{a.del.bounds.head, a.del.bounds.head}
			b.loadBytes(a.del.text, false)
		}

		b.dirtyLines(0, len(b.lines))
		b.autoScroll()
	}
}

// commitAction finalizes b.currentAction and adds it to the list, to become the
// new b.lastAction.
func (b *Buffer) commitAction() bool {
	if b.currentAction.del == nil && b.currentAction.ins == nil {
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

func (b *Buffer) clearHist() {
	b.lastAction.prev = nil
	b.lastAction.next = nil
	b.savedAction = b.lastAction
}
