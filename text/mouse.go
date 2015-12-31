package text

import (
	"image"
	"sort"
	"time"

	"golang.org/x/mobile/event/mouse"
)

const dClickPause = 500 * time.Millisecond

func (b *Buffer) handleMouseEvent(e mouse.Event) {
	pos := image.Pt(int(e.X), int(e.Y)).Add(b.clipr.Min) // adjust for scrolling
	button := e.Button

	oldpos := b.mPos
	b.mButton = button
	b.mPos = pos

	switch button {
	case mouse.ButtonLeft:
		if e.Direction == mouse.DirPress {
			// click
			a := b.pt2address(pos)
			olda := b.pt2address(oldpos)
			b.mSweepOrigin = a
			b.click(a, olda, button)
			b.commitAction()
		} else if e.Direction == mouse.DirNone {
			// sweep
			// possibly scroll by sweeping past the edge of the window
			if pos.Y <= b.clipr.Min.Y {
				b.scroll(image.Pt(0, -b.lineHeight))
				pos.Y -= b.lineHeight
			} else if pos.Y >= b.clipr.Max.Y {
				b.scroll(image.Pt(0, b.lineHeight))
				pos.Y += b.lineHeight
			}

			a := b.pt2address(pos)
			olda := b.pt2address(oldpos)
			if a != olda {
				b.sweep(olda, a)
			}
		}
	case mouse.ButtonWheelDown:
		b.scroll(image.Pt(0, -b.lineHeight))
	case mouse.ButtonWheelUp:
		b.scroll(image.Pt(0, b.lineHeight))
	}
}

func (b *Buffer) pt2address(pt image.Point) address {
	// (0,0) if pt is above the buffer
	if pt.Y < 0 {
		return address{}
	}

	var pos address
	pos.row = pt.Y / b.lineHeight

	// end of the last line if pos is below the last line
	if pos.row > len(b.lines)-1 {
		pos.row = len(b.lines) - 1
		pos.col = len(b.lines[pos.row].s)
		return pos
	}

	line := b.lines[pos.row]
	// the column number is found by looking for the smallest px element
	// which is larger than pt.X, and returning the column number before that.
	// If no px elements are larger than pt.X, then return the last column on
	// the line.
	if pt.X <= line.px[0] {
		pos.col = 0
	} else if pt.X > line.px[len(line.px)-1] {
		pos.col = len(line.px) - 1
	} else {
		n := sort.Search(len(line.px), func(i int) bool {
			return line.px[i] > pt.X
		})
		pos.col = n - 1
	}
	return pos
}

func (b *Buffer) click(a, olda address, button mouse.Button) {
	switch button {
	case mouse.ButtonLeft:
		b.dirtyLines(b.dot.head.row, b.dot.tail.row+1)
		b.dot.head, b.dot.tail = a, a
		b.dirtyLine(a.row)

		if time.Since(b.lastClickTime) < dClickPause && a == olda {
			// double click
			b.expandSel(a)
			b.lastClickTime = time.Time{}
		} else {
			b.lastClickTime = time.Now()
		}
	}
}

func (b *Buffer) sweep(from, to address) {
	// mark all the rows between to and from as dirty
	// (to and from can be more than one row apart, if they are sweeping quickly)
	r1, r2 := to.row, from.row
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	b.dirtyLines(r1, r2+1)

	// set the selection
	if to.lessThan(b.mSweepOrigin) {
		b.dot = selection{to, b.mSweepOrigin}
	} else if to != b.mSweepOrigin {
		b.dot = selection{b.mSweepOrigin, to}
	} else {
		b.dirtyLine(to.row)
		b.dot = selection{b.mSweepOrigin, b.mSweepOrigin}
	}
}
