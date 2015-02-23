package text

import (
	"image"
	"image/draw"
	"log"
)

func (b *Buffer) redraw() {
	log.Println("redraw")

	// clear an area if requested
	draw.Draw(b.img, b.clear, b.bgCol, image.ZP, draw.Src)
	b.clear = image.ZR

	selection := !(b.dot.Head == b.dot.Tail)

	// redraw dirty lines
	for row, line := range b.lines {
		if line.dirty {
			line.dirty = false

			pt := image.Pt(0, row*b.font.height)

			// clear the line, unless it is completely selected
			if !selection || row <= b.dot.Head.Row || row >= b.dot.Tail.Row {
				r := image.Rect(pt.X, pt.Y, pt.X+b.img.Bounds().Dx(), pt.Y+b.font.height)
				draw.Draw(b.img, r, b.bgCol, image.ZP, draw.Src)
			}

			// draw selection rectangles
			if selection && (row >= b.dot.Head.Row && row <= b.dot.Tail.Row) {
				b.drawSel(row)
			}

			// draw font overtop
			line.px, _ = b.font.draw(b.img, pt, string(line.s), -1)
		}
	}

	// draw cursor
	if !selection {
		// subtract a pixel from x coordinate to match acme
		pt := image.Pt(b.getxpx(b.dot.Head)-1, b.getypx(b.dot.Head.Row))
		draw.Draw(b.img, b.cursor.Rect.Add(pt), b.cursor, image.ZP, draw.Src)
	}
}

func (b *Buffer) drawSel(row int) {
	x1 := 0
	if row == b.dot.Head.Row {
		x1 = b.getxpx(b.dot.Head)
	}
	x2 := b.img.Bounds().Dx()
	if row == b.dot.Tail.Row {
		x2 = b.getxpx(b.dot.Tail)
	}
	min := image.Pt(x1, b.getypx(row))
	max := image.Pt(x2, b.getypx(row+1))
	r := image.Rectangle{min, max}
	draw.Draw(b.img, r, b.selCol, image.ZP, draw.Src)
}

// returns x (pixels) for a given Address
func (b *Buffer) getxpx(a Address) int {
	l := b.lines[a.Row]
	if a.Col >= len(l.px) {
		return l.px[len(l.px)-1] + b.img.Bounds().Min.X
	}
	return l.px[a.Col] + b.img.Bounds().Min.X
}

// returns y (pixels) for a given row
func (b *Buffer) getypx(row int) int {
	return b.img.Bounds().Min.Y + row*b.font.height
}
