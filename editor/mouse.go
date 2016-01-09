package editor

import (
	"image"
	"sort"
	"time"

	"sigint.ca/graphics/editor/internal/text"

	"golang.org/x/mobile/event/mouse"
)

const dClickPause = 500 * time.Millisecond

func (ed *Editor) handleMouseEvent(e mouse.Event) {
	pos := image.Pt(int(e.X), int(e.Y)).Add(ed.clipr.Min) // adjust for scrolling
	button := e.Button

	oldpos := ed.mPos
	ed.mButton = button
	ed.mPos = pos

	switch button {
	case mouse.ButtonLeft:
		if e.Direction == mouse.DirPress {
			// click
			a := ed.pt2address(pos)
			olda := ed.pt2address(oldpos)
			ed.mSweepOrigin = a
			ed.click(a, olda, button)
			ed.history.Commit()
		} else if e.Direction == mouse.DirNone {
			// sweep
			// possibly scroll by sweeping past the edge of the window
			if pos.Y <= ed.clipr.Min.Y {
				ed.scroll(image.Pt(0, -ed.lineHeight))
				pos.Y -= ed.lineHeight
			} else if pos.Y >= ed.clipr.Max.Y {
				ed.scroll(image.Pt(0, ed.lineHeight))
				pos.Y += ed.lineHeight
			}

			a := ed.pt2address(pos)
			olda := ed.pt2address(oldpos)
			if a != olda {
				ed.sweep(olda, a)
			}
		}
	case mouse.ButtonWheelDown:
		ed.scroll(image.Pt(0, -ed.lineHeight))
	case mouse.ButtonWheelUp:
		ed.scroll(image.Pt(0, ed.lineHeight))
	}
}

func (ed *Editor) pt2address(pt image.Point) text.Address {
	// (0,0) if pt is above the buffer
	if pt.Y < 0 {
		return text.Address{}
	}

	var addr text.Address
	addr.Row = pt.Y / ed.lineHeight

	// end of the last line if addr is below the last line
	if addr.Row > len(ed.lines)-1 {
		addr.Row = len(ed.lines) - 1
		addr.Col = len(ed.lines[addr.Row].s)
		return addr
	}

	line := ed.lines[addr.Row]
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

func (ed *Editor) click(a, olda text.Address, button mouse.Button) {
	switch button {
	case mouse.ButtonLeft:
		ed.dirtyLines(ed.dot.From.Row, ed.dot.To.Row+1)
		ed.dot.From, ed.dot.To = a, a
		ed.dirtyLine(a.Row)

		if time.Since(ed.lastClickTime) < dClickPause && a == olda {
			// double click
			ed.expandSel(a)
			ed.lastClickTime = time.Time{}
		} else {
			ed.lastClickTime = time.Now()
		}
	}
}

func (ed *Editor) sweep(from, to text.Address) {
	// mark all the rows between to and from as dirty
	// (to and from can be more than one row apart, if they are sweeping quickly)
	r1, r2 := to.Row, from.Row
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	ed.dirtyLines(r1, r2+1)

	// set the selection
	if to.LessThan(ed.mSweepOrigin) {
		ed.dot = text.Selection{to, ed.mSweepOrigin}
	} else if to != ed.mSweepOrigin {
		ed.dot = text.Selection{ed.mSweepOrigin, to}
	} else {
		ed.dirtyLine(to.Row)
		ed.dot = text.Selection{ed.mSweepOrigin, ed.mSweepOrigin}
	}
}
