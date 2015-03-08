package text

import "image"

// each editor action is a deletion followed by an insertion.
type action struct {
	deletion   *change
	insertion  *change
	prev, next *action
}

type change struct {
	bounds Selection
	text   string
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
			b.dot = Selection{a.insertion.bounds.Head, a.insertion.bounds.Head}
			b.load(a.insertion.text, false)
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
			b.dot = Selection{a.deletion.bounds.Head, a.deletion.bounds.Head}
			b.load(a.deletion.text, false)
		}

		b.dirtyLines(0, len(b.lines))
		b.autoScroll()
	}
}

// commitAction finalizes b.currentAction and adds it to the list,
// becoming the new b.lastAction.
func (b *Buffer) commitAction() {
	if b.currentAction.deletion == nil && b.currentAction.insertion == nil {
		return
	}
	if b.lastAction != nil {
		b.lastAction.next = b.currentAction
	}
	b.currentAction.prev = b.lastAction
	b.currentAction.next = nil
	b.lastAction = b.currentAction
	b.currentAction = new(action)
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
