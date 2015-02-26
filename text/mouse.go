package text

import (
	"image"
	"time"
)

const (
	b1 = 1 << iota
	b2
	b3
	b4 // mouse wheel up
	b5 // mouse wheel down
)

const dClickPause = 500 * time.Millisecond

func (b *Buffer) handleMouseEvent(pos image.Point, buttons int) {
	pos = pos.Add(b.clipr.Min) // adjust for scrolling

	oldbuttons := b.mButtons
	oldpos := b.mPos
	b.mButtons = buttons
	b.mPos = pos

	switch buttons {
	case b1:
		click := oldbuttons == 0
		sweep := oldbuttons == buttons
		if click {
			a := b.pt2Address(pos)
			b.mSweepOrigin = a
			b.click(a, buttons)
			b.lines[a.Row].dirty = true
		} else if sweep {
			// possibly scroll by sweeping past the edge of the window
			if pos.Y == b.clipr.Min.Y {
				b.scroll(image.Pt(0, -b.font.height))
				pos.Y -= b.font.height
			} else if pos.Y == b.clipr.Max.Y {
				b.scroll(image.Pt(0, b.font.height))
				pos.Y += b.font.height
			}

			a := b.pt2Address(pos)
			oldA := b.pt2Address(oldpos)
			b.sweep(oldA, a)
		}
	case b4:
		b.scroll(image.Pt(0, -b.font.height))
	case b5:
		b.scroll(image.Pt(0, b.font.height))
	}
}

func (b *Buffer) pt2Address(pt image.Point) Address {
	pt = pt.Sub(b.img.Rect.Min)
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

	if pt.X < 0 {
		return Address{pos.Row, 0}
	}

	line := b.lines[pos.Row]
	for i := range line.px {
		if line.px[i]+b.img.Bounds().Min.X > pt.X {
			if i > 0 {
				pos.Col = i - 1
			}
			return pos
		}
	}
	pos.Col = len(line.s)
	return pos
}

func (b *Buffer) click(pos Address, buttons int) {
	b.dirtyImg = true
	switch buttons {
	case b1:
		for _, line := range b.lines[b.dot.Head.Row : b.dot.Tail.Row+1] {
			line.dirty = true
		}
		b.dot.Head, b.dot.Tail = pos, pos
		b.lines[pos.Row].dirty = true
		if b.dClicking == true && pos == b.dot.Head && pos == b.dot.Tail {
			b.dClick(pos)
			b.dClicking = false
			b.dClickTimer.Stop()
		} else {
			b.dClicking = true
			if b.dClickTimer != nil {
				b.dClickTimer.Stop()
			}
			b.dClickTimer = time.AfterFunc(dClickPause, func() { b.dClicking = false })
		}
	}
}

func (b *Buffer) dClick(a Address) {
	b.expandSel(a)
}

func (b *Buffer) sweep(from, to Address) {
	if from == to {
		return // no change in selection
	}
	b.dirtyImg = true

	// mark all the rows between to and from as dirty
	// (to and from can be more than one row apart, if they are sweeping quickly)
	r1, r2 := to.Row, from.Row
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	for _, line := range b.lines[r1 : r2+1] {
		line.dirty = true
	}

	// set the selection
	if to.lessThan(b.mSweepOrigin) {
		b.dot = Selection{to, b.mSweepOrigin}
	} else if to != b.mSweepOrigin {
		b.dot = Selection{b.mSweepOrigin, to}
	} else {
		b.lines[to.Row].dirty = true
		b.dot = Selection{b.mSweepOrigin, b.mSweepOrigin}
	}
}
