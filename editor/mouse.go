package editor

import (
	"image"
	"time"

	"sigint.ca/graphics/editor/internal/text"

	"golang.org/x/mobile/event/mouse"
)

const (
	b1 = 1 << uint(mouse.ButtonLeft)
	b2 = 1 << uint(mouse.ButtonMiddle)
	b3 = 1 << uint(mouse.ButtonRight)
)

const dClickPause = 500 * time.Millisecond
const twitch = 3 // pixels

type mouseState struct {
	buttons       uint32 // a bit field of mouse buttons currently pressed
	chording      bool
	lastClickTime time.Time // used to detect a double-click

	pt image.Point
	a  text.Address

	sweepOrigin image.Point  // the origin of a sweep
	sweepLast   text.Address // the last column that was swept
}

func (ed *Editor) handleMouseEvent(e mouse.Event) {
	if e.Button == mouse.ButtonNone {
		return
	}

	// a mouse event commits any pending transformation
	ed.commitTransformation()

	ed.m.pt = image.Pt(int(e.X), int(e.Y)).Add(ed.visible().Min) // adjust for scrolling
	ed.m.a = ed.getAddress(ed.m.pt)

	if e.Direction == mouse.DirPress {
		ed.click(e)
	} else if e.Direction == mouse.DirNone {
		ed.sweep(e)
	} else if e.Direction == mouse.DirRelease {
		ed.release(e)
	}
}

func (ed *Editor) click(e mouse.Event) {
	a, pt := ed.m.a, ed.m.pt

	ed.m.buttons |= 1 << uint(e.Button)
	ed.m.sweepOrigin = pt

	switch ed.m.buttons {
	case b1:
		prev := ed.dot
		ed.dot.From, ed.dot.To = a, a

		// check for double-click
		if time.Since(ed.m.lastClickTime) < dClickPause && ed.dot == prev {
			ed.dot = ed.buf.AutoSelect(a)
			ed.m.lastClickTime = time.Time{}
		} else {
			ed.m.lastClickTime = time.Now()
		}

	case b2, b3:
		ed.dot = ed.buf.SelWord(a)

	case b1 | b2:
		// cut
		ed.m.chording = true
		ed.initTransformation()
		ed.snarf()
		ed.dot = ed.buf.ClearSel(ed.dot)
		ed.commitTransformation()

	case b1 | b3:
		// paste
		ed.m.chording = true
		ed.initTransformation()
		ed.paste()
		ed.commitTransformation()
	}

	ed.dirty = true
}

// sweep sweeps (selects) text.
func (ed *Editor) sweep(e mouse.Event) {
	a, pt := ed.m.a, ed.m.pt

	vis := ed.visible()
	if a == ed.m.sweepLast && pt.In(vis) {
		return
	}
	if ed.m.chording == true {
		return
	}
	if isTwitch(pt, ed.m.sweepOrigin) {
		return
	}

	if pt.Y <= vis.Min.Y && vis.Min.Y > 0 {
		ed.scroll(image.Pt(0, ed.font.height))
	} else if pt.Y >= vis.Max.Y && vis.Max.Y < (len(ed.buf.Lines)-1)*ed.font.height {
		ed.scroll(image.Pt(0, -ed.font.height))
	}

	ed.m.sweepLast = a

	origin := ed.getAddress(ed.m.sweepOrigin)
	if a.LessThan(origin) {
		ed.dot = text.Selection{a, origin}
	} else if a != origin {
		ed.dot = text.Selection{origin, a}
	} else {
		ed.dot = text.Selection{origin, origin}
	}

	ed.dirty = true
}

// isTwitch reports whether p1 is within 1 twitch distance of p2.
func isTwitch(p1, p2 image.Point) bool {
	size := image.Pt(twitch, twitch)
	r := image.Rectangle{p2.Sub(size), p2.Add(size)}
	return p1.In(r)
}

func (ed *Editor) release(e mouse.Event) {
	switch ed.m.buttons {
	case b2:
		if !ed.m.chording && ed.B2Action != nil {
			ed.B2Action(ed.buf.GetSel(ed.dot))
			ed.dirty = true
		}
	case b3:
		if !ed.m.chording && ed.B3Action != nil {
			ed.B3Action(ed.buf.GetSel(ed.dot))
			ed.dirty = true
		}
	}

	ed.m.buttons &^= 1 << uint(e.Button)
	if ed.m.buttons&(b1|b2|b3) == 0 {
		ed.m.chording = false
	}
}
