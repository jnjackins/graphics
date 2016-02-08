package editor

import (
	"image"
	"time"

	"sigint.ca/graphics/editor/internal/address"

	"golang.org/x/mobile/event/mouse"
)

// SendMouseEvent sends a mouse event to be interpreted by the Editor.
func (ed *Editor) SendMouseEvent(e mouse.Event) {
	ed.handleMouseEvent(e)
}

// SendScrollEvent sends a scroll event to be interpreted by the Editor.
func (ed *Editor) SendScrollEvent(e mouse.ScrollEvent) {
	var pt image.Point
	if e.Precise {
		pt.X = int(e.Dx)
		pt.Y = int(e.Dy)
	} else {
		pt.X = int(e.Dx * float32(ed.font.height))
		pt.Y = int(e.Dy * float32(ed.font.height))
	}
	oldPt := ed.scrollPt
	ed.scroll(pt)
	if ed.scrollPt != oldPt {
		ed.dirty = true
	}
}

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
	a  address.Simple

	sweepOrigin image.Point    // the origin of a sweep
	sweepLast   address.Simple // the last column that was swept
}

func (ed *Editor) handleMouseEvent(e mouse.Event) {
	if e.Button == mouse.ButtonNone {
		return
	}

	// a mouse event commits any pending transformation
	ed.commitTransformation()

	pt := image.Pt(int(e.X), int(e.Y))
	ed.m.pt = pt.Add(ed.visible().Min) // adjust for scrolling
	ed.m.a = ed.getAddress(ed.m.pt)

	if pt.In(ed.sbRect()) && e.Direction != mouse.DirRelease {
		ed.clickSb(e)
	} else if e.Direction == mouse.DirPress {
		ed.click(e)
	} else if e.Direction == mouse.DirNone {
		ed.sweep(e)
	} else if e.Direction == mouse.DirRelease {
		ed.release(e)
	}
}

// scrollbar click
func (ed *Editor) clickSb(e mouse.Event) {
	height := float64(ed.visible().Dy())
	percent := (float64(e.Y) + float64(ed.r.Min.Y)) / height
	// disregard any chording; act on individual mouse.DirPress events
	switch e.Button {
	case mouse.ButtonLeft:
		ed.scroll(image.Pt(0, int(height*percent)))
	case mouse.ButtonMiddle:
		ed.scrollPt.Y = int(float64(ed.docHeight()) * percent)
	case mouse.ButtonRight:
		ed.scroll(image.Pt(0, int(-height*percent)))
	}
	ed.dirty = true
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
		if ed.dot.IsEmpty() || !a.In(ed.dot) {
			ed.dot = ed.buf.SelWord(a)
		}

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
	} else if pt.Y >= vis.Max.Y && vis.Max.Y < ed.docHeight() {
		ed.scroll(image.Pt(0, -ed.font.height))
	}

	ed.m.sweepLast = a

	origin := ed.getAddress(ed.m.sweepOrigin)
	if a.LessThan(origin) {
		ed.dot = address.Selection{a, origin}
	} else if a != origin {
		ed.dot = address.Selection{origin, a}
	} else {
		ed.dot = address.Selection{origin, origin}
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
