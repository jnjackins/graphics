package text

import "image"

// each editor action is a deletion followed by an insertion.
type action struct {
	deletionBounds  Selection
	deletionText    string
	insertionBounds Selection
	insertionText   string
	prev, next      *action
}

func (b *Buffer) initCurrentAction() {
	if b.currentAction == nil {
		b.currentAction = new(action)
	}
}

func (b *Buffer) redo() {
	if b.lastAction.next != nil {
		a := b.lastAction.next
		b.lastAction = b.lastAction.next

		// Replace the deleted text with the inserted text. This must be
		// done as two separate steps, because one of the deletion or insertion
		// may be a no-op, i.e. the bounds aren't set.
		b.dot = a.deletionBounds
		b.deleteSel(false)
		if a.insertionText != "" {
			b.dot = Selection{a.insertionBounds.Head, a.insertionBounds.Head}
			b.load(a.insertionText, false)
		}

		b.dirtyLines(0, len(b.lines))
		b.autoScroll()
	}
}

func (b *Buffer) undo() {
	if b.currentAction != nil {
		b.commitAction()
	}
	if b.lastAction.prev != nil {
		a := b.lastAction
		b.lastAction = b.lastAction.prev

		// Replace the inserted text with the deleted text. This must be
		// done as two separate steps, because one of the deletion or insertion
		// may be a no-op, i.e. the bounds aren't set.
		b.dot = a.insertionBounds
		b.deleteSel(false)
		if a.deletionText != "" {
			b.dot = Selection{a.deletionBounds.Head, a.deletionBounds.Head}
			b.load(a.deletionText, false)
		}

		b.dirtyLines(0, len(b.lines))
		b.autoScroll()
	}
}

// commitAction finalizes b.currentAction and adds it to the list,
// becoming the new b.lastAction.
func (b *Buffer) commitAction() {
	if b.currentAction == nil {
		return
	}
	if b.lastAction != nil {
		b.lastAction.next = b.currentAction
	}
	b.currentAction.prev = b.lastAction
	b.currentAction.next = nil // should be anyway, but make sure
	b.lastAction = b.currentAction
	b.currentAction = nil
}

// autoScroll does nothing if b.dot.Head is currently in view, or
// scrolls so that it is 20% down from the top of the screen if it is not.
func (b *Buffer) autoScroll() {
	headpx := b.dot.Head.Row * b.font.height
	if headpx < b.clipr.Min.Y || headpx > b.clipr.Max.Y {
		padding := int(0.20 * float64(b.clipr.Dy()))
		padding -= padding % b.font.height
		scrollpt := image.Pt(0, b.dot.Head.Row*b.font.height-padding)
		b.clipr = image.Rectangle{scrollpt, scrollpt.Add(b.clipr.Size())}
		b.scroll(image.ZP) // this doesn't scroll, but fixes b.clipr if it is out-of-bounds
	}
}
