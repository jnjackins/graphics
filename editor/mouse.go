package editor

import (
	"image"
	"time"

	"sigint.ca/graphics/editor/address"

	"golang.org/x/mobile/event/mouse"
)

// SendMouseEvent sends a mouse event to be interpreted by the Editor.
func (ed *Editor) SendMouseEvent(e mouse.Event) {
	if e.Button == mouse.ButtonScroll {
		ed.handleScrollEvent(e)
		return
	}
	ed.handleMouseEvent(e)
}

const (
	b1 = 1 << uint(mouse.ButtonLeft)
	b2 = 1 << uint(mouse.ButtonMiddle)
	b3 = 1 << uint(mouse.ButtonRight)
)

const dClickPause = 500 * time.Millisecond
const twitch = 3 // pixels

type mouseState struct {
	buttons       uint32    // a bit field of mouse buttons currently pressed
	chording      bool      // a chord has been initiated
	scrolling     bool      // the scroll bar is being manipulated
	lastClickTime time.Time // used to detect a double-click

	pt image.Point
	a  address.Simple

	sweepOrigin image.Point    // the origin of a sweep
	sweepLast   address.Simple // the last column that was swept
}

func (ed *Editor) handleScrollEvent(e mouse.Event) {
	if !e.PreciseScrolling {
		e.ScrollDelta.X *= ed.fontHeight
		e.ScrollDelta.Y *= ed.fontHeight
	}
	oldPt := ed.scrollPt
	ed.scroll(e.ScrollDelta)
	if ed.scrollPt != oldPt {
		ed.dirty = true
	}
}

func (ed *Editor) handleMouseEvent(e mouse.Event) {
	if e.Button == mouse.ButtonNone {
		return
	}

	// a mouse event commits any pending transformation
	ed.commitTransformation()

	ed.m.pt = e.Pos.Add(ed.visible().Min) // adjust for scrolling
	ed.m.a = ed.getAddress(ed.m.pt)

	if e.Direction == mouse.DirRelease {
		ed.release(e)
	} else if e.Direction == mouse.DirPress && e.Pos.In(ed.sbRect()) || ed.m.scrolling {
		ed.clickSb(e)
	} else if e.Direction == mouse.DirPress {
		ed.click(e)
	} else if e.Direction == mouse.DirNone {
		ed.sweep(e)
	}
}

// scrollbar click
func (ed *Editor) clickSb(e mouse.Event) {
	ed.m.scrolling = true

	height := float64(ed.visible().Dy())
	percent := (float64(e.Pos.Y) + float64(ed.r.Min.Y)) / height

	switch e.Button {
	case mouse.ButtonLeft:
		d := int(height * percent)
		if d < 0 {
			d = 0
		}
		ed.scroll(image.Pt(0, d))

	case mouse.ButtonRight:
		d := int(-height * percent)
		if d > 0 {
			d = 0
		}
		ed.scroll(image.Pt(0, d))

	case mouse.ButtonMiddle:
		ed.scrollPt.Y = int(float64(ed.docHeight()) * percent)
		ed.scroll(image.ZP) // fix potential invalid scrollPt

	}
	ed.dirty = true
}

func (ed *Editor) click(e mouse.Event) {
	a, pt := ed.m.a, ed.m.pt

	ed.m.buttons |= 1 << uint(e.Button)
	ed.m.sweepOrigin = pt

	dprintf("click: ed.m.buttons: %v\n", ed.m.buttons)
	switch ed.m.buttons {
	case b1:
		prev := ed.Dot
		ed.Dot.From, ed.Dot.To = a, a

		// check for double-click
		if time.Since(ed.m.lastClickTime) < dClickPause && ed.Dot == prev {
			ed.Dot = ed.Buffer.AutoSelect(a)
			ed.m.lastClickTime = time.Time{}
		} else {
			ed.m.lastClickTime = time.Now()
		}

	case b2, b3:
		if ed.Dot.IsEmpty() || !a.In(ed.Dot) {
			ed.Dot = ed.Buffer.SelWord(a)
		}

	case b1 | b2:
		// cut
		ed.m.chording = true
		ed.initTransformation()
		ed.snarf()
		ed.Dot = ed.Buffer.ClearSel(ed.Dot)
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
		ed.scroll(image.Pt(0, ed.fontHeight))
	} else if pt.Y >= vis.Max.Y && vis.Max.Y < ed.docHeight() {
		ed.scroll(image.Pt(0, -ed.fontHeight))
	}

	ed.m.sweepLast = a

	origin := ed.getAddress(ed.m.sweepOrigin)
	if a.LessThan(origin) {
		ed.Dot = address.Selection{a, origin}
	} else if a != origin {
		ed.Dot = address.Selection{origin, a}
	} else {
		ed.Dot = address.Selection{origin, origin}
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
	ed.m.scrolling = false

	switch ed.m.buttons {
	case b2:
		if !ed.m.chording && ed.B2Action != nil {
			ed.B2Action(ed.Buffer.GetSel(ed.Dot))
			ed.dirty = true
		}
	case b3:
		if !ed.m.chording && ed.B3Action != nil {
			ed.B3Action(ed.Buffer.GetSel(ed.Dot))
			ed.dirty = true
		}
	}

	ed.m.buttons &^= 1 << uint(e.Button)
	if ed.m.buttons&(b1|b2|b3) == 0 {
		dprintf("release: ed.mchording=false (was %v)\n", ed.m.chording)
		ed.m.chording = false
	}
	dprintf("release: ed.m.buttons = %v\n", ed.m.buttons)
}
