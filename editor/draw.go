package editor

import (
	"image"
	"image/draw"

	"sigint.ca/graphics/editor/internal/text"
)

func (ed *Editor) redraw() {
	draw.Draw(ed.img, ed.Bounds(), ed.bgcol, image.ZP, draw.Src)

	from, to := ed.dirtyRows()
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
			ed.drawSel(row)
		}

		// draw font overtop
		pt := image.Pt(ed.getxpx(text.Address{row, 0}), ed.getypx(row))
		ed.font.draw(ed.img, pt, line)
	}

	// draw cursor
	if ed.dot.IsEmpty() {
		cursor := ed.cursor(ed.font.height)
		// subtract a pixel from x coordinate to match acme
		pt := image.Pt(ed.getxpx(ed.dot.From)-1, ed.getypx(ed.dot.From.Row))
		draw.Draw(ed.img, cursor.Bounds().Add(pt), cursor, image.ZP, draw.Over)
	}
}

func (ed *Editor) drawSel(row int) {
	x1 := ed.getxpx(text.Address{row, 0})
	if row == ed.dot.From.Row {
		x1 = ed.getxpx(ed.dot.From)
	}
	x2 := ed.Bounds().Dx()
	if row == ed.dot.To.Row {
		x2 = ed.getxpx(ed.dot.To)
	}
	min := image.Pt(x1, ed.getypx(row))
	max := image.Pt(x2, ed.getypx(row+1))
	r := image.Rectangle{min, max}
	draw.Draw(ed.img, r, ed.selcol, image.ZP, draw.Src)
}

func (ed *Editor) visible() image.Rectangle {
	return image.Rectangle{
		Min: ed.scrollPt,
		Max: ed.scrollPt.Add(ed.Bounds().Size()),
	}
}

func (ed *Editor) dirtyRows() (from, to int) {
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
	ymax := (len(ed.buf.Lines) - 1) * ed.font.height
	if ed.visible().Min.Y > ymax {
		ed.scrollPt.Y = ymax
	}
}

// returns x (pixels) for a given address
func (ed *Editor) getxpx(a text.Address) (x int) {
	l := ed.buf.Lines[a.Row]
	if len(l.Adv) == 0 {
		x = 0
	} else if a.Col >= len(l.Adv) {
		x = int(l.Adv[len(l.Adv)-1])
	} else {
		x = int(l.Adv[a.Col])
	}
	return x - ed.visible().Min.X + ed.margin.X
}

// returns y (pixels) for a given row
func (ed *Editor) getypx(row int) (y int) {
	return (row * ed.font.height) - ed.visible().Min.Y + ed.margin.Y
}
