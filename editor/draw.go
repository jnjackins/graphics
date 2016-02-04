package editor

import (
	"image"
	"image/draw"
	"math"
	"sort"

	"sigint.ca/graphics/editor/internal/address"
)

const sbwidth = 20

func (ed *Editor) draw(dst *image.RGBA) {
	ed.r = dst.Bounds()
	draw.Draw(dst, ed.r, ed.opts.BG1, image.ZP, draw.Src)
	ed.drawSb(dst)

	from, to := ed.visibleRows()
	for row := from; row < to; row++ {
		line := ed.buf.Lines[row]

		// draw selection rectangles
		if !ed.dot.IsEmpty() && (row >= ed.dot.From.Row && row <= ed.dot.To.Row) {
			// If some text has just been inserted (e.g. via paste from clipboard),
			// it may not have been measured yet - selection rects need to be drawn
			// before text. This only applies to when row == ed.dot.From.Row or
			// row == ed.dot.To.Row (otherwise the entire line is selected and character
			// advances are not necessary).
			// TODO: only measure if line text has changed
			if row == ed.dot.From.Row || row == ed.dot.To.Row {
				ed.font.measure(line)
			}
			ed.drawSelRect(dst, row)
		}

		// draw font overtop
		pt := ed.getPixelsRel(address.Simple{Row: row, Col: 0})
		ed.font.draw(dst, pt, line, ed.opts.Text)
	}

	// draw cursor
	if ed.dot.IsEmpty() {
		cursor := ed.opts.Cursor(ed.font.height)
		pt := ed.getPixelsRel(ed.dot.From)
		pt.X-- // match acme
		draw.Draw(dst, cursor.Bounds().Add(pt), cursor, image.ZP, draw.Over)
	}
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
	r.Max.Y += ed.font.height

	draw.Draw(dst, r, ed.opts.Sel, image.ZP, draw.Src)
}

func (ed *Editor) docHeight() int {
	return (len(ed.buf.Lines) - 1) * ed.font.height
}

func (ed *Editor) visible() image.Rectangle {
	return image.Rectangle{
		Min: ed.scrollPt,
		Max: ed.scrollPt.Add(ed.r.Size()),
	}
}

func (ed *Editor) visibleRows() (from, to int) {
	from = ed.visible().Min.Y / ed.font.height
	to = ed.visible().Max.Y/ed.font.height + 2
	if to > len(ed.buf.Lines) {
		to = len(ed.buf.Lines)
	}
	return
}

func (ed *Editor) scroll(pt image.Point) {
	ed.scrollPt.Y -= pt.Y

	// check boundaries
	if ed.visible().Min.Y < 0 {
		ed.scrollPt.Y = 0
	}
	max := ed.getPixelsAbs(address.Simple{Row: len(ed.buf.Lines) - 1})
	if ed.visible().Min.Y > max.Y {
		ed.scrollPt.Y = max.Y
	}
}

func (ed *Editor) autoscroll() {
	visible := ed.visible()
	pt := ed.getPixelsAbs(address.Simple{Row: ed.dot.From.Row})
	if pt.Y > visible.Min.Y && pt.Y+ed.font.height < visible.Max.Y {
		return
	}

	ed.scrollPt.Y = pt.Y - int(.2*float64(visible.Dy()))

	// scroll fixes boundary conditions, since we manually set ed.scrollPt
	ed.scroll(image.ZP)
}

func (ed *Editor) getPixelsAbs(a address.Simple) image.Point {
	var x, y int
	l := ed.buf.Lines[a.Row]

	if len(l.Adv) == 0 {
		x = 0
	} else if a.Col >= len(l.Adv) {
		x = int(l.Adv[len(l.Adv)-1])
	} else {
		x = int(l.Adv[a.Col])
	}

	y = a.Row * ed.font.height

	pt := image.Pt(x, y).Add(ed.opts.Margin)
	if ed.opts.ScrollBar {
		pt.X += sbwidth
	}
	return pt
}

func (ed *Editor) getPixelsRel(a address.Simple) image.Point {
	return ed.getPixelsAbs(a).Sub(ed.visible().Min)
}

func (ed *Editor) getAddress(pt image.Point) address.Simple {
	pt = pt.Sub(ed.opts.Margin)
	if ed.opts.ScrollBar {
		pt.X -= sbwidth
	}

	// (0,0) if pt is above the buffer
	if pt.Y < 0 {
		return address.Simple{}
	}

	var addr address.Simple
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
			return int(line.Adv[i]) > pt.X+1
		})
		addr.Col = n - 1
	}
	return addr
}

func (ed *Editor) sbRect() image.Rectangle {
	if !ed.opts.ScrollBar {
		return image.ZR
	}
	return image.Rect(0, 0, sbwidth, ed.visible().Dy())
}

func (ed *Editor) drawSb(dst *image.RGBA) {
	if !ed.opts.ScrollBar {
		return
	}
	draw.Draw(dst, ed.sbRect(), ed.opts.BG2, image.ZP, draw.Src)
	slider := sliderRect(ed.visible(), ed.docHeight(), sbwidth)
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
