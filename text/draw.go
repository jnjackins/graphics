package text

import (
	"image"
	"image/draw"
)

func (b *Buffer) redraw() {
	// clear an area if requested
	draw.Draw(b.img, b.clear, b.bgCol, image.ZP, draw.Src)
	b.dirty = b.dirty.Union(b.clear)
	b.clear = image.ZR

	selection := !(b.Dot.Head == b.Dot.Tail)

	// redraw dirty lines
	var grown bool
	for row, line := range b.lines {
		if line.dirty {
			line.dirty = false

			// the top left pixel of the line, relative to the image it's being drawn onto
			pt := image.Pt(0, row*b.lineHeight).Add(b.margin)

			// make sure b.img is big enough to show this line and one full clipr past it
			if pt.Y+b.clipr.Dy() >= b.img.Bounds().Dy() {
				b.growImg()
				grown = true
			}

			// clear the line, unless it is completely selected
			if !selection || row <= b.Dot.Head.Row || row >= b.Dot.Tail.Row {
				// clear all the way to the left side of the image; the margin may have bits of cursor in it
				r := image.Rect(b.img.Bounds().Min.X, pt.Y, pt.X+b.img.Bounds().Dx(), pt.Y+b.lineHeight)
				draw.Draw(b.img, r, b.bgCol, image.ZP, draw.Src)
			}

			// draw selection rectangles
			if selection && (row >= b.Dot.Head.Row && row <= b.Dot.Tail.Row) {
				b.drawSel(row)
			}

			// draw font overtop
			line.px = b.font.draw(b.img, pt, line.s)
		}
	}
	if grown {
		b.shrinkImg() // shrink back down to the correct size
	}

	// draw cursor
	if !selection {
		// subtract a pixel from x coordinate to match acme
		pt := image.Pt(b.getxpx(b.Dot.Head)-1, b.getypx(b.Dot.Head.Row))
		draw.Draw(b.img, b.cursor.Bounds().Add(pt), b.cursor, image.ZP, draw.Src)
	}
}

func (b *Buffer) drawSel(row int) {
	x1 := b.margin.X
	if row == b.Dot.Head.Row {
		x1 = b.getxpx(b.Dot.Head)
	}
	x2 := b.img.Bounds().Dx()
	if row == b.Dot.Tail.Row {
		x2 = b.getxpx(b.Dot.Tail)
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
	max.Y = (len(b.lines)-1)*b.lineHeight + b.clipr.Dy()
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
		return l.px[len(l.px)-1]
	}
	return l.px[a.Col]
}

// returns y (pixels) for a given row
func (b *Buffer) getypx(row int) int {
	return row * b.lineHeight
}

func (b *Buffer) growImg() {
	r := b.img.Bounds()
	r.Max.Y += b.img.Bounds().Dy() // new image is double the old
	newImg := image.NewRGBA(r)
	draw.Draw(newImg, newImg.Bounds(), b.bgCol, image.ZP, draw.Src)
	draw.Draw(newImg, newImg.Bounds(), b.img, image.ZP, draw.Src)
	b.img = newImg
	b.dirty = b.img.Bounds()
}

func (b *Buffer) shrinkImg() {
	r := b.img.Bounds()
	height := (len(b.lines)-1)*b.lineHeight + b.clipr.Dy()
	if r.Max.Y != height {
		r.Max.Y = height
		b.img = b.img.SubImage(r).(*image.RGBA)
	}
	b.dirty = b.dirty.Intersect(b.img.Bounds())
}

func (b *Buffer) dirtyLine(row int) {
	b.lines[row].dirty = true
	r := b.img.Bounds()
	r.Min.Y = row * b.lineHeight
	r.Max.Y = r.Min.Y + b.lineHeight
	b.dirty = b.dirty.Union(r)
}

// TODO: this is kinda dumb, callers actually need to give row2+1
func (b *Buffer) dirtyLines(row1, row2 int) {
	for _, line := range b.lines[row1:row2] {
		line.dirty = true
	}
	r := b.img.Bounds()
	r.Min.Y = row1 * b.lineHeight
	r.Max.Y = row2*b.lineHeight + b.lineHeight
	b.dirty = b.dirty.Union(r)
}

// autoScroll does nothing if b.Dot.Head is currently in view, or
// scrolls so that it is 20% down from the top of the screen if it is not.
func (b *Buffer) autoScroll() {
	headpx := b.Dot.Head.Row * b.lineHeight
	if headpx < b.clipr.Min.Y || headpx > b.clipr.Max.Y-b.lineHeight {
		padding := int(0.20 * float64(b.clipr.Dy()))
		padding -= padding % b.lineHeight
		scrollpt := image.Pt(0, b.Dot.Head.Row*b.lineHeight-padding)
		b.clipr = image.Rectangle{scrollpt, scrollpt.Add(b.clipr.Size())}
		b.scroll(image.ZP) // this doesn't scroll, but fixes b.clipr if it is out-of-bounds
	}
}
