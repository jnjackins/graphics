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

	pos := image.Pt(int(e.X), int(e.Y)).Add(ed.visible().Min) // adjust for scrolling
	a := ed.pt2address(pos)

	switch e.Button {
	case mouse.ButtonLeft:
		ed.commitTransformation()

		if e.Direction == mouse.DirPress {
			// click
			ed.dirty = true
			ed.sweepOrigin = a
			ed.click(a, e.Button)
		} else if e.Direction == mouse.DirNone {
			// sweep
			vis := ed.visible()
			if a == ed.sweepLast && pos.In(vis) {
				return
			}
			if pos.Y <= vis.Min.Y && vis.Min.Y > 0 {
				ed.scroll(image.Pt(0, ed.font.height))
			} else if pos.Y >= vis.Max.Y && vis.Max.Y < (len(ed.buf.Lines)-1)*ed.font.height {
				ed.scroll(image.Pt(0, -ed.font.height))
			}

			ed.dirty = true
			ed.sweepLast = a

			ed.sweep(a)
		}
	}
}

func (ed *Editor) pt2address(pt image.Point) text.Address {
	// (0,0) if pt is above the buffer
	if pt.Y < 0 {
		return text.Address{}
	}

	var addr text.Address
	addr.Row = pt.Y / ed.font.height

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

func (ed *Editor) click(a text.Address, button mouse.Button) {
	switch button {
	case mouse.ButtonLeft:
		prev := ed.dot
		ed.dot.From, ed.dot.To = a, a

		if time.Since(ed.lastClickTime) < dClickPause && ed.dot == prev {
			// double click
			ed.dot = ed.buf.AutoSelect(a)
			ed.lastClickTime = time.Time{}
		} else {
			ed.lastClickTime = time.Now()
		}
	}
}

func (ed *Editor) sweep(to text.Address) {
	if to.LessThan(ed.sweepOrigin) {
		ed.dot = text.Selection{to, ed.sweepOrigin}
	} else if to != ed.sweepOrigin {
		ed.dot = text.Selection{ed.sweepOrigin, to}
	} else {
		ed.dot = text.Selection{ed.sweepOrigin, ed.sweepOrigin}
	}
}
