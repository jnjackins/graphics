package text

import (
	"image"
	"image/draw"
)

func (b *Buffer) redraw() {
	// clear an area if requested
	draw.Draw(b.img, b.clear, b.bgCol, image.ZP, draw.Src)
	b.clear = image.ZR

	selection := !(b.dot.Head == b.dot.Tail)

	// redraw dirty lines
	count := 0
	for row, line := range b.lines {
		if line.dirty {
			count++
			line.dirty = false

			pt := image.Pt(0, row*b.font.height).Add(b.margin) // the top left pixel of the line

			// make sure b.img is big enough to show this line and one full clipr past it
			if pt.Y+b.clipr.Dy() >= b.img.Bounds().Max.Y {
				b.growImg()
			}

			// clear the line, unless it is completely selected
			if !selection || row <= b.dot.Head.Row || row >= b.dot.Tail.Row {
				// clear all the way to the left side of the image; the margin may have bits of cursor in it
				r := image.Rect(b.img.Bounds().Min.X, pt.Y, pt.X+b.img.Bounds().Dx(), pt.Y+b.font.height)
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
		draw.Draw(b.img, b.cursor.Bounds().Add(pt), b.cursor, image.ZP, draw.Src)
	}

	//log.Println("redraw:", count, "lines")
}

func (b *Buffer) drawSel(row int) {
	x1 := b.margin.X
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

func (b *Buffer) scroll(pt image.Point) {
	b.clipr = b.clipr.Add(pt)

	// check boundaries
	min := b.img.Bounds().Min
	max := b.img.Bounds().Max
	max.Y = (len(b.lines)-1)*b.font.height + b.clipr.Dy()
	if b.clipr.Min.X < min.X {
		b.clipr = image.Rect(min.X, b.clipr.Min.Y, min.X+b.clipr.Dx(), b.clipr.Max.Y)
	}
	if b.clipr.Min.Y < min.Y {
		b.clipr = image.Rect(b.clipr.Min.X, min.Y, b.clipr.Max.X, min.Y+b.clipr.Dy())
	}
	if b.clipr.Max.X > max.X {
		b.clipr = image.Rect(max.X-b.clipr.Dx(), b.clipr.Min.Y, max.X, b.clipr.Max.Y)
	}
	if b.clipr.Max.Y > max.Y {
		b.clipr = image.Rect(b.clipr.Min.X, max.Y-b.clipr.Dy(), b.clipr.Max.X, max.Y)
	}
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

func (b *Buffer) growImg() {
	// new image is double the old
	r := b.img.Bounds()
	r.Max.Y += b.img.Bounds().Dy()
	newImg := image.NewRGBA(r)
	draw.Draw(newImg, newImg.Bounds(), b.bgCol, image.ZP, draw.Src)

	// draw the old image onto the new image
	draw.Draw(newImg, newImg.Bounds(), b.img, image.ZP, draw.Src)

	b.img = newImg

	//log.Println("growImg:", b.img.Bounds())
}
