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
	pos = pos.Sub(b.pos)       // asdjust for placement of the buffer
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
			b.dirtyLine(a.Row)
			b.commitAction()
		} else if sweep {
			// possibly scroll by sweeping past the edge of the window
			if pos.Y <= b.clipr.Min.Y {
				b.scroll(image.Pt(0, -b.font.height))
				pos.Y -= b.font.height
			} else if pos.Y >= b.clipr.Max.Y {
				b.scroll(image.Pt(0, b.font.height))
				pos.Y += b.font.height
			}

			a := b.pt2Address(pos)
			oldA := b.pt2Address(oldpos)
			if a != oldA {
				b.sweep(oldA, a)
			}
		}
	case b4:
		b.scroll(image.Pt(0, -b.font.height))
	case b5:
		b.scroll(image.Pt(0, b.font.height))
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

	// TODO: seems wasteful to iterate over the whole line here
	line := b.lines[pos.Row]
	for i := 1; i < len(line.px); i++ {
		if line.px[i] > pt.X {
			pos.Col = i - 1
			return pos
		}
	}
	pos.Col = len(line.s)
	return pos
}

func (b *Buffer) click(pos Address, buttons int) {
	switch buttons {
	case b1:
		b.dirtyLines(b.Dot.Head.Row, b.Dot.Tail.Row+1)
		b.Dot.Head, b.Dot.Tail = pos, pos
		b.dirtyLine(pos.Row)
		if b.dClicking == true && pos == b.Dot.Head && pos == b.Dot.Tail {
			b.dClick(pos)
			b.dClicking = false
			b.dClickTimer.Stop()
		} else {
			b.dClicking = true
			if b.dClickTimer != nil {
				b.dClickTimer.Stop()
			}
			// TODO: this is racy
			b.dClickTimer = time.AfterFunc(dClickPause, func() { b.dClicking = false })
		}
	}
}

func (b *Buffer) dClick(a Address) {
	b.expandSel(a)
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
