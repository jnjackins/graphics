package editor

import (
	"image"
	"sort"
	"time"

	"sigint.ca/graphics/editor/internal/text"

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
			b.history.Commit()
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

func (b *Buffer) pt2address(pt image.Point) text.Address {
	// (0,0) if pt is above the buffer
	if pt.Y < 0 {
		return text.Address{}
	}

	var addr text.Address
	addr.Row = pt.Y / b.lineHeight

	// end of the last line if addr is below the last line
	if addr.Row > len(b.lines)-1 {
		addr.Row = len(b.lines) - 1
		addr.Col = len(b.lines[addr.Row].s)
		return addr
	}

	line := b.lines[addr.Row]
	// the column number is found by looking for the smallest px element
	// which is larger than pt.X, and returning the column number before that.
	// If no px elements are larger than pt.X, then return the last column on
	// the line.
	if pt.X <= line.adv[0] {
		addr.Col = 0
	} else if pt.X > line.adv[len(line.adv)-1] {
		addr.Col = len(line.adv) - 1
	} else {
		n := sort.Search(len(line.adv), func(i int) bool {
			return line.adv[i] > pt.X
		})
		addr.Col = n - 1
	}
	return addr
}

func (b *Buffer) click(a, olda text.Address, button mouse.Button) {
	switch button {
	case mouse.ButtonLeft:
		b.dirtyLines(b.dot.From.Row, b.dot.To.Row+1)
		b.dot.From, b.dot.To = a, a
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

func (b *Buffer) sweep(from, to text.Address) {
	// mark all the rows between to and from as dirty
	// (to and from can be more than one row apart, if they are sweeping quickly)
	r1, r2 := to.Row, from.Row
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	b.dirtyLines(r1, r2+1)

	// set the selection
	if to.LessThan(b.mSweepOrigin) {
		b.dot = text.Selection{to, b.mSweepOrigin}
	} else if to != b.mSweepOrigin {
		b.dot = text.Selection{b.mSweepOrigin, to}
	} else {
		b.dirtyLine(to.Row)
		b.dot = text.Selection{b.mSweepOrigin, b.mSweepOrigin}
	}
}
