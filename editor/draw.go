package editor

import (
	"image"
	"image/draw"
	"math"
	"sort"

	"sigint.ca/graphics/editor/address"
)

// Draw draws the editor onto dst within the bounding rectangle dr, and returns
// the height in pixels of the text that was drawn.
func (ed *Editor) Draw(dst *image.RGBA, dr image.Rectangle) int {
	ed.dirty = false
	return ed.draw(dst, dr)
}

// Dirty reports whether the next call to Draw will result in a different image than the previous call.
func (ed *Editor) Dirty() bool {
	return ed.dirty
}

func (ed *Editor) draw(dst *image.RGBA, dr image.Rectangle) int {
	ed.r = dr
	draw.Draw(dst, ed.r, ed.opts.BG1, image.ZP, draw.Src)
	ed.drawSb(dst)

	from, to := ed.visibleRows()
	for row := from; row < to; row++ {
		line := ed.buffer.Lines[row]

		// draw selection rectangles
		if !ed.dot.IsEmpty() && (row >= ed.dot.From.Row && row <= ed.dot.To.Row) {
			// If some text has just been inserted (e.g. via paste from clipboard),
			// it may not have been measured yet - selection rects need to be drawn
			// before text. This only applies to when row == ed.dot.From.Row or
			// row == ed.dot.To.Row (otherwise the entire line is selected and character
			// advances are not necessary).
			// TODO: only measure if line text has changed
			if row == ed.dot.From.Row || row == ed.dot.To.Row {
				ed.measureString(line.String())
			}
			ed.drawSelRect(dst, row)
		}

		// draw font overtop
		pt := ed.getPixelsRel(address.Simple{Row: row, Col: 0})
		ed.drawString(dst, pt, line.String(), ed.opts.Text)
	}

	// draw cursor
	if ed.dot.IsEmpty() {
		cursor := ed.opts.Cursor(ed.fontHeight)
		pt := ed.getPixelsRel(ed.dot.From)
		pt.X-- // match acme
		draw.Draw(dst, cursor.Bounds().Add(pt), cursor, image.ZP, draw.Over)
	}

	return (to - from) * ed.fontHeight
}

func (ed *Editor) drawSelRect(dst *image.RGBA, row int) {
	var r image.Rectangle

	if row == ed.dot.From.Row {
		r.Min = ed.getPixelsRel(ed.dot.From)
	} else {
		r.Min = ed.getPixelsRel(address.Simple{Row: row, Col: 0})
	}

	if row == ed.dot.To.Row {
		r.Max = ed.getPixelsRel(ed.dot.To)
	} else {
		r.Max = ed.getPixelsRel(address.Simple{Row: row, Col: 0})
		r.Max.X = ed.r.Dx()
	}
	r.Max.Y += ed.fontHeight

	draw.Draw(dst, r, ed.opts.Sel, image.ZP, draw.Src)
}

func (ed *Editor) docHeight() int {
	return (len(ed.buffer.Lines) - 1) * ed.fontHeight
}

func (ed *Editor) visible() image.Rectangle {
	return image.Rectangle{
		Min: ed.scrollPt,
		Max: ed.scrollPt.Add(ed.r.Size()),
	}
}

func (ed *Editor) visibleRows() (from, to int) {
	from = ed.visible().Min.Y / ed.fontHeight
	to = ed.visible().Max.Y/ed.fontHeight + 2
	if to > len(ed.buffer.Lines) {
		to = len(ed.buffer.Lines)
	}
	return
}

func (ed *Editor) scroll(pt image.Point) {
	ed.scrollPt.Y -= pt.Y

	// check boundaries
	if ed.visible().Min.Y < 0 {
		ed.scrollPt.Y = 0
	}
	max := ed.getPixelsAbs(address.Simple{Row: len(ed.buffer.Lines) - 1})
	if ed.visible().Min.Y > max.Y {
		ed.scrollPt.Y = max.Y
	}
}

func (ed *Editor) autoscroll() {
	visible := ed.visible()
	pt := ed.getPixelsAbs(address.Simple{Row: ed.dot.From.Row})
	if pt.Y > visible.Min.Y && pt.Y+ed.fontHeight < visible.Max.Y {
		return
	}

	ed.scrollPt.Y = pt.Y - int(.2*float64(visible.Dy()))

	// scroll fixes boundary conditions, since we manually set ed.scrollPt
	ed.scroll(image.ZP)
}

func (ed *Editor) getPixelsAbs(a address.Simple) image.Point {
	var x, y int
	s := ed.buffer.Lines[a.Row].String()

	if len(s) == 0 {
		// fast path
		x = 0
	} else {
		adv := ed.measureString(s)
		if a.Col >= len(adv) {
			x = adv[len(adv)-1].Round()
		} else {
			x = adv[a.Col].Round()
		}
	}

	y = a.Row * ed.fontHeight

	pt := image.Pt(x+ed.margin, y)
	if ed.opts.ScrollBar {
		pt.X += ed.sbwidth
	}
	return pt
}

func (ed *Editor) getPixelsRel(a address.Simple) image.Point {
	return ed.getPixelsAbs(a).Sub(ed.visible().Min)
}

func (ed *Editor) getAddress(pt image.Point) address.Simple {
	pt.X -= ed.margin
	if ed.opts.ScrollBar {
		pt.X -= ed.sbwidth
	}

	// (0,0) if pt is above the buffer
	if pt.Y < 0 {
		return address.Simple{}
	}

	var addr address.Simple
	addr.Row = pt.Y / ed.fontHeight

	// end of the last line if addr is below the last line
	if addr.Row > len(ed.buffer.Lines)-1 {
		addr.Row = len(ed.buffer.Lines) - 1
		addr.Col = ed.buffer.Lines[addr.Row].RuneCount()
		return addr
	}

	adv := ed.measureString(ed.buffer.Lines[addr.Row].String())
	// the column number is found by looking for the smallest px element
	// which is larger than pt.X, and returning the column number before that.
	// If no px elements are larger than pt.X, then return the last column on
	// the line.
	if len(adv) == 0 || pt.X <= adv[0].Round() {
		addr.Col = 0
	} else if pt.X > adv[len(adv)-1].Round() {
		addr.Col = len(adv) - 1
	} else {
		n := sort.Search(len(adv), func(i int) bool {
			return adv[i].Round() > pt.X+1
		})
		addr.Col = n - 1
	}
	return addr
}

func (ed *Editor) sbRect() image.Rectangle {
	if !ed.opts.ScrollBar {
		return image.ZR
	}
	return image.Rect(0, 0, ed.sbwidth, ed.visible().Dy())
}

func (ed *Editor) drawSb(dst *image.RGBA) {
	if !ed.opts.ScrollBar {
		return
	}
	draw.Draw(dst, ed.sbRect(), ed.opts.BG2, image.ZP, draw.Src)
	slider := sliderRect(ed.visible(), ed.docHeight(), ed.sbwidth)
	draw.Draw(dst, slider, ed.opts.BG1, image.ZP, draw.Src)
}

func sliderRect(visible image.Rectangle, docHeight, width int) image.Rectangle {
	barHeight := float64(visible.Dy())
	h := barHeight * float64(visible.Dy()) / float64(docHeight)
	sliderHeight := int(math.Max(h, 10))
	pos := int(barHeight * float64(visible.Min.Y) / float64(docHeight))
	pos -= 3 // show a wee bit of slider when we're scrolled to the bottom, like acme
	return image.Rect(0, pos, width-1, pos+sliderHeight)
}
