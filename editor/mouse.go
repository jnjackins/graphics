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
	if e.Direction == mouse.DirRelease {
		return
	}

	ed.dirty = true

	// a mouse event triggers a history commit, in case there is some
	// uncommitted input
	// TODO: this assumes mouse events to not cause any text transformations,
	// and will need to change when cut/paste via mouse chords is added.
	ed.initTransformation()
	ed.commitTransformation()

	pos := image.Pt(int(e.X), int(e.Y)).Add(ed.visible().Min) // adjust for scrolling
	button := e.Button

	oldpos := ed.mPos
	ed.mPos = pos

	switch button {
	case mouse.ButtonLeft:
		if e.Direction == mouse.DirPress {
			// click
			a := ed.pt2address(pos)
			olda := ed.pt2address(oldpos)
			ed.mSweepOrigin = a
			ed.click(a, olda, button)
		} else if e.Direction == mouse.DirNone {
			// sweep
			// possibly scroll by sweeping past the edge of the window
			if pos.Y <= ed.visible().Min.Y {
				ed.scroll(image.Pt(0, ed.lineHeight))
				pos.Y -= ed.lineHeight
			} else if pos.Y >= ed.visible().Max.Y {
				ed.scroll(image.Pt(0, -ed.lineHeight))
				pos.Y += ed.lineHeight
			}

			a := ed.pt2address(pos)
			olda := ed.pt2address(oldpos)
			if a != olda {
				ed.sweep(olda, a)
			}
		}
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
	if addr.Row > len(ed.buf.Lines)-1 {
		addr.Row = len(ed.buf.Lines) - 1
		addr.Col = ed.buf.Lines[addr.Row].RuneCount()
		return addr
	}

	line := ed.buf.Lines[addr.Row]
	// the column number is found by looking for the smallest px element
	// which is larger than pt.X, and returning the column number before that.
	// If no px elements are larger than pt.X, then return the last column on
	// the line.
	if len(line.Adv) == 0 || pt.X <= int(line.Adv[0]) {
		addr.Col = 0
	} else if pt.X > int(line.Adv[len(line.Adv)-1]) {
		addr.Col = len(line.Adv) - 1
	} else {
		n := sort.Search(len(line.Adv), func(i int) bool {
			return int(line.Adv[i]) > pt.X
		})
		addr.Col = n - 1
	}
	return addr
}

func (ed *Editor) click(a, olda text.Address, button mouse.Button) {
	switch button {
	case mouse.ButtonLeft:
		ed.dot.From, ed.dot.To = a, a

		if time.Since(ed.lastClickTime) < dClickPause && a == olda {
			// double click
			ed.dot = ed.buf.AutoSelect(a)
			ed.lastClickTime = time.Time{}
		} else {
			ed.lastClickTime = time.Now()
		}
	}
}

func (ed *Editor) sweep(from, to text.Address) {
	if to.LessThan(ed.mSweepOrigin) {
		ed.dot = text.Selection{to, ed.mSweepOrigin}
	} else if to != ed.mSweepOrigin {
		ed.dot = text.Selection{ed.mSweepOrigin, to}
	} else {
		ed.dot = text.Selection{ed.mSweepOrigin, ed.mSweepOrigin}
	}
}
