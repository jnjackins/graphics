package editor

import (
	"image"
	"image/draw"
	"sort"

	"sigint.ca/graphics/editor/internal/text"
)

func (ed *Editor) redraw() {
	draw.Draw(ed.img, ed.Bounds(), ed.bgcol, image.ZP, draw.Src)

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
			ed.drawSelRect(row)
		}

		// draw font overtop
		pt := ed.getPixelsRel(text.Address{Row: row, Col: 0})
		ed.font.draw(ed.img, pt, line)
	}

	// draw cursor
	if ed.dot.IsEmpty() {
		cursor := ed.cursor(ed.font.height)
		pt := ed.getPixelsRel(ed.dot.From)
		pt.X-- // match acme
		draw.Draw(ed.img, cursor.Bounds().Add(pt), cursor, image.ZP, draw.Over)
	}
}

func (ed *Editor) drawSelRect(row int) {
	var r image.Rectangle

	if row == ed.dot.From.Row {
		r.Min = ed.getPixelsRel(ed.dot.From)
	} else {
		r.Min = ed.getPixelsRel(text.Address{Row: row, Col: 0})
	}
	if row == ed.dot.To.Row {
		r.Max = ed.getPixelsRel(ed.dot.To)
	} else {
		r.Max = ed.getPixelsRel(ed.dot.To)
		r.Max.X = ed.Bounds().Dx()
	}
	r.Max.Y += ed.font.height

	draw.Draw(ed.img, r, ed.selcol, image.ZP, draw.Src)
}

func (ed *Editor) visible() image.Rectangle {
	return image.Rectangle{
		Min: ed.scrollPt,
		Max: ed.scrollPt.Add(ed.Bounds().Size()),
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
	max := ed.getPixelsAbs(text.Address{Row: len(ed.buf.Lines) - 1})
	if ed.visible().Min.Y > max.Y {
		ed.scrollPt.Y = max.Y
	}
}

func (ed *Editor) autoscroll() {
	visible := ed.visible()
	pt := ed.getPixelsAbs(text.Address{Row: ed.dot.From.Row})
	if pt.Y > visible.Min.Y && pt.Y < visible.Max.Y {
		return
	}

	ed.scrollPt.Y = pt.Y - int(.2*float64(visible.Dy()))

	// scroll fixes boundary conditions, since we manually set ed.scrollPt
	ed.scroll(image.ZP)
}

func (ed *Editor) getPixelsAbs(a text.Address) image.Point {
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

	return image.Pt(x, y).Add(ed.margin)
}

func (ed *Editor) getPixelsRel(a text.Address) image.Point {
	return ed.getPixelsAbs(a).Sub(ed.visible().Min)
}

func (ed *Editor) getAddress(pt image.Point) text.Address {
	pt = pt.Sub(ed.margin)

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
			return int(line.Adv[i]) > pt.X+1
		})
		addr.Col = n - 1
	}
	return addr
}
