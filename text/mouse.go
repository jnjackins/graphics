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
			a := b.pt2Address(pos)
			olda := b.pt2Address(oldpos)
			b.mSweepOrigin = a
			b.click(a, olda, button)
			b.commitAction()
		} else if e.Direction == mouse.DirNone {
			// sweep
			// possibly scroll by sweeping past the edge of the window
			if pos.Y <= b.clipr.Min.Y {
				b.scroll(image.Pt(0, -b.font.height))
				pos.Y -= b.font.height
			} else if pos.Y >= b.clipr.Max.Y {
				b.scroll(image.Pt(0, b.font.height))
				pos.Y += b.font.height
			}

			a := b.pt2Address(pos)
			olda := b.pt2Address(oldpos)
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

func (b *Buffer) pt2Address(pt image.Point) Address {
	// (0,0) if pt is above the buffer
	if pt.Y < 0 {
		return Address{}
	}

	var pos Address
	pos.Row = pt.Y / b.font.height

	// end of the last line if pos is below the last line
	if pos.Row > len(b.lines)-1 {
		pos.Row = len(b.lines) - 1
		pos.Col = len(b.lines[pos.Row].s)
		return pos
	}

	line := b.lines[pos.Row]
	// the column number is found by looking for the smallest px element
	// which is larger than pt.X, and returning the column number before that.
	// If no px elements are larger than pt.X, then return the last column on
	// the line.
	if pt.X <= line.px[0] {
		pos.Col = 0
	} else if pt.X > line.px[len(line.px)-1] {
		pos.Col = len(line.px)
	} else {
		n := sort.Search(len(line.px), func(i int) bool {
			return line.px[i] > pt.X
		})
		pos.Col = n - 1
	}
	return pos
}

func (b *Buffer) click(a, olda Address, button mouse.Button) {
	switch button {
	case mouse.ButtonLeft:
		b.dirtyLines(b.Dot.Head.Row, b.Dot.Tail.Row+1)
		b.Dot.Head, b.Dot.Tail = a, a
		b.dirtyLine(a.Row)

		if time.Since(b.lastClickTime) < dClickPause && a == olda {
			// double click
			b.expandSel(a)
			b.lastClickTime = time.Time{}
		} else {
			b.lastClickTime = time.Now()
		}
	}
}

func (b *Buffer) sweep(from, to Address) {
	// mark all the rows between to and from as dirty
	// (to and from can be more than one row apart, if they are sweeping quickly)
	r1, r2 := to.Row, from.Row
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	b.dirtyLines(r1, r2+1)

	// set the selection
	if to.lessThan(b.mSweepOrigin) {
		b.Dot = Selection{to, b.mSweepOrigin}
	} else if to != b.mSweepOrigin {
		b.Dot = Selection{b.mSweepOrigin, to}
	} else {
		b.dirtyLine(to.Row)
		b.Dot = Selection{b.mSweepOrigin, b.mSweepOrigin}
	}
}
