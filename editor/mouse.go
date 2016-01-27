package editor

import (
	"image"
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
	a := ed.getAddress(pos)

	ed.commitTransformation()

	if e.Direction == mouse.DirPress {
		ed.dirty = true
		ed.sweepOrigin = a
		ed.click(a, e.Button)
	} else if e.Direction == mouse.DirNone {
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
	case mouse.ButtonMiddle, mouse.ButtonRight:
		ed.dot = ed.buf.SelWord(a)
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
