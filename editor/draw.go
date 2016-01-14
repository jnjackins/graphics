package editor

import (
	"image"
	"image/draw"

	"sigint.ca/graphics/editor/internal/text"
)

func (ed *Editor) redraw() {
	dirtyLines, r := ed.getDirty()
	if len(dirtyLines) == 0 {
		return
	}
	draw.Draw(ed.img, r, ed.bgcol, image.ZP, draw.Src)

	var grown bool

	// redraw lines
	for row, line := range dirtyLines {
		line.Dirty = false

		// the top left pixel of the line, relative to the image it's being drawn onto
		pt := image.Pt(0, row*ed.lineHeight).Add(ed.margin)

		// make sure ed.img is big enough to show this line and one full clipr past it
		if pt.Y+ed.clipr.Dy() >= ed.img.Bounds().Dy() {
			ed.growImg()
			grown = true
		}

		// clear the line, unless it is completely selected
		if ed.dot.IsEmpty() || row <= ed.dot.From.Row || row >= ed.dot.To.Row {
			// clear all the way to the left side of the image; the margin may have bits of cursor in it
			r := image.Rect(ed.img.Bounds().Min.X, pt.Y, pt.X+ed.img.Bounds().Dx(), pt.Y+ed.lineHeight)
			draw.Draw(ed.img, r, ed.bgcol, image.ZP, draw.Src)
		}

		// draw selection rectangles
		if !ed.dot.IsEmpty() && (row >= ed.dot.From.Row && row <= ed.dot.To.Row) {
			// If some text has just been inserted (e.g. via paste from clipboard),
			// it may not have been measured yet - selection rects need to be drawn
			// before text. This only applies to when row == ed.dot.From.Row or
			// row == ed.dot.To.Row (otherwise the entire line is selected and character
			// advances are not necessary).
			if row == ed.dot.From.Row || row == ed.dot.To.Row {
				// TODO: avoid measuring if we're sure the line hasn't changed since last draw
				line.Adv = ed.font.measure(ed.margin.X, line.String())
			}
			ed.drawSel(row)
		}

		// draw font overtop
		line.Adv = ed.font.draw(ed.img, pt, line.String())
	}
	if grown {
		ed.shrinkImg() // shrink back down to the correct size
	}

	// draw cursor
	if ed.dot.IsEmpty() {
		// subtract a pixel from x coordinate to match acme
		pt := image.Pt(ed.getxpx(ed.dot.From)-1, ed.getypx(ed.dot.From.Row))
		draw.Draw(ed.img, ed.cursor.Bounds().Add(pt), ed.cursor, image.ZP, draw.Src)
	}
}

// TODO
func (ed *Editor) getDirty() ([]*text.Line, image.Rectangle) {
	return ed.buf.Lines, ed.img.Bounds()
}

func (ed *Editor) drawSel(row int) {
	x1 := ed.margin.X
	if row == ed.dot.From.Row {
		x1 = ed.getxpx(ed.dot.From)
	}
	x2 := ed.img.Bounds().Dx()
	if row == ed.dot.To.Row {
		x2 = ed.getxpx(ed.dot.To)
	}
	min := image.Pt(x1, ed.getypx(row))
	max := image.Pt(x2, ed.getypx(row+1))
	r := image.Rectangle{min, max}
	draw.Draw(ed.img, r, ed.selcol, image.ZP, draw.Src)
}

func (ed *Editor) scroll(pt image.Point) {
	ed.clipr = ed.clipr.Sub(pt)

	// check boundaries
	min := ed.img.Bounds().Min
	max := ed.img.Bounds().Max
	max.Y = (len(ed.buf.Lines)-1)*ed.lineHeight + ed.clipr.Dy()
	if ed.clipr.Min.X < min.X {
		ed.clipr = image.Rect(min.X, ed.clipr.Min.Y, min.X+ed.clipr.Dx(), ed.clipr.Max.Y)
	}
	if ed.clipr.Min.Y < min.Y {
		ed.clipr = image.Rect(ed.clipr.Min.X, min.Y, ed.clipr.Max.X, min.Y+ed.clipr.Dy())
	}
	if ed.clipr.Max.X > max.X {
		ed.clipr = image.Rect(max.X-ed.clipr.Dx(), ed.clipr.Min.Y, max.X, ed.clipr.Max.Y)
	}
	if ed.clipr.Max.Y > max.Y {
		ed.clipr = image.Rect(ed.clipr.Min.X, max.Y-ed.clipr.Dy(), ed.clipr.Max.X, max.Y)
	}
}

// returns x (pixels) for a given address
func (ed *Editor) getxpx(a text.Address) int {
	l := ed.buf.Lines[a.Row]
	if a.Col >= len(l.Adv) {
		return int(l.Adv[len(l.Adv)-1])
	}
	return int(l.Adv[a.Col])
}

// returns y (pixels) for a given row
func (ed *Editor) getypx(row int) int {
	return row * ed.lineHeight
}

func (ed *Editor) growImg() {
	r := ed.img.Bounds()
	r.Max.Y += ed.img.Bounds().Dy() // new image is double the old
	newImg := image.NewRGBA(r)
	draw.Draw(newImg, newImg.Bounds(), ed.bgcol, image.ZP, draw.Src)
	draw.Draw(newImg, newImg.Bounds(), ed.img, image.ZP, draw.Src)
	ed.img = newImg
}

func (ed *Editor) shrinkImg() {
	r := ed.img.Bounds()
	height := (len(ed.buf.Lines)-1)*ed.lineHeight + ed.clipr.Dy()
	if r.Max.Y != height {
		r.Max.Y = height
		ed.img = ed.img.SubImage(r).(*image.RGBA)
	}
}

// autoScroll does nothing if ed.dot.From is currently in view, or
// scrolls so that it is 20% down from the top of the screen if it is not.
func (ed *Editor) autoScroll() {
	headpx := ed.dot.From.Row * ed.lineHeight
	if headpx < ed.clipr.Min.Y || headpx > ed.clipr.Max.Y-ed.lineHeight {
		padding := int(0.20 * float64(ed.clipr.Dy()))
		padding -= padding % ed.lineHeight
		scrollpt := image.Pt(0, ed.dot.From.Row*ed.lineHeight-padding)
		ed.clipr = image.Rectangle{scrollpt, scrollpt.Add(ed.clipr.Size())}
		ed.scroll(image.ZP) // this doesn't scroll, but fixes ed.clipr if it is out-of-bounds
	}
}
