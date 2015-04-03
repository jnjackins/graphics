// this file was adapted from the freetype package at
// https://github.com/jnjackins/freetype-go.

package text

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"io"
	"io/ioutil"

	"sigint.ca/freetype-go/freetype/raster"
	"sigint.ca/freetype-go/freetype/truetype"
)

// These constants determine the size of the glyph cache. The cache is keyed
// primarily by the glyph index modulo nGlyphs, and secondarily by sub-pixel
// position for the mask image. Sub-pixel positions are quantized to
// nXFractions possible values in both the x and y directions.
const (
	nGlyphs     = 256
	nXFractions = 4
	nYFractions = 1
)

// An entry in the glyph cache is keyed explicitly by the glyph index and
// implicitly by the quantized x and y fractional offset. It maps to a mask
// image and an offset.
type cacheEntry struct {
	valid        bool
	glyph        truetype.Index
	advanceWidth raster.Fix32
	mask         *image.Alpha
	offset       image.Point
}

// pixelsToRaster converts an image.Point (pixels) to a raster.Point
func pixelsToRaster(pt image.Point) raster.Point {
	return raster.Point{
		X: raster.Fix32(pt.X << 8),
		Y: raster.Fix32(pt.Y << 8),
	}
}

// A ttf holds the state for drawing text in a given font and size.
type ttf struct {
	r        *raster.Rasterizer
	font     *truetype.Font
	glyphBuf *truetype.GlyphBuf

	src image.Image // source image for drawing

	// size and dpi are used to calculate scale. scale is the number of
	// 26.6 fixed point units in 1 em.
	size, dpi float64
	scale     int32

	height int // height of the font in pixels

	// cache is the glyph cache.
	cache    [nGlyphs * nXFractions * nYFractions]cacheEntry
	tabwidth raster.Fix32
}

// newTTF creates a new font from the bytes in r. The font is given a default
// size, color, and dpi of 12, black, and 96.
func newTTF(r io.Reader) (*ttf, error) {
	// Read the font data.
	fontBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	f := &ttf{
		font:     font,
		r:        raster.NewRasterizer(0, 0),
		src:      image.NewUniform(image.Black),
		glyphBuf: truetype.NewGlyphBuf(),
		size:     12,
		dpi:      96,
		scale:    12 << 6,
	}
	f.recalc()
	f.tabwidth = f.tabpx(4)
	return f, nil
}

// setDPI sets the screen resolution in dots per inch.
func (f *ttf) setDPI(dpi float64) {
	if f.dpi == dpi {
		return
	}
	f.dpi = dpi
	f.recalc()
}

// setSize sets the font size in points (as in "a 12 point font").
func (f *ttf) setSize(size float64) {
	if f.size == size {
		return
	}
	f.size = size
	f.recalc()
}

// setSrc sets the font color (default is black)
func (f *ttf) setSrc(c color.Color) {
	f.src = image.NewUniform(c)
}

// draw draws s onto dst starting at pt, up to a maximum length of w pixels.
// It returns a slice of x-coords for each rune, as well as a string containing
// the portion of s which exceeded w and was not drawn.
func (f *ttf) draw(dst draw.Image, pt image.Point, s string, w int) ([]int, string) {
	px := make([]int, 1, len(s)+1)
	px[0] = pt.X
	pt.Y += f.height - 3
	p := pixelsToRaster(pt)
	maxwidth := raster.Fix32(w << 8)
	startp := p
	prev, hasPrev := truetype.Index(0), false
	for i, rune := range s {
		// deal with tabstop specially
		if rune == '\t' {
			p.X += f.tabwidth
			px = append(px, int(p.X>>8))
			continue
		}
		index := f.font.Index(rune)
		if hasPrev {
			kern := raster.Fix32(f.font.Kerning(f.scale, prev, index)) << 2
			kern = (kern + 128) &^ 255
			p.X += kern
		}
		advanceWidth, mask, offset, err := f.glyph(index, p)
		if err != nil {
			panic(err) // TODO: put in some placeholder character
		}
		p.X += advanceWidth
		glyphRect := mask.Bounds().Add(offset)
		dr := dst.Bounds().Intersect(glyphRect)
		if !dr.Empty() && (w < 0 || p.X-startp.X <= maxwidth) {
			mp := image.Point{0, dr.Min.Y - glyphRect.Min.Y}
			draw.DrawMask(dst, dr, f.src, image.ZP, mask, mp, draw.Over)
		} else if maxwidth > 0 && p.X-startp.X > maxwidth {
			return px, s[i:]
		}
		prev, hasPrev = index, true
		px = append(px, int(p.X>>8))
	}
	return px, ""
}

// getPx is like draw, but it only calculates the x positions of the characters
// without actually drawing them. startX is the initial x position to calculate
// from.
func (f *ttf) getPx(startX int, s string) []int {
	px := make([]int, 1, len(s)+1)
	px[0] = startX
	p := pixelsToRaster(image.Pt(startX, 0))
	prev, hasPrev := truetype.Index(0), false
	for _, rune := range s {
		// deal with tabstop specially
		if rune == '\t' {
			p.X += f.tabwidth
			px = append(px, int(p.X>>8))
			continue
		}
		index := f.font.Index(rune)
		if hasPrev {
			kern := raster.Fix32(f.font.Kerning(f.scale, prev, index)) << 2
			kern = (kern + 128) &^ 255
			p.X += kern
		}
		advanceWidth, _, _, err := f.glyph(index, p)
		if err != nil {
			panic(err) // TODO: put in (width of) some placeholder character
		}
		p.X += advanceWidth
		prev, hasPrev = index, true
		px = append(px, int(p.X>>8))
	}
	return px
}

// tabpx returns the width of a tabstop as a raster.Fix32
func (f *ttf) tabpx(width int) raster.Fix32 {
	space := f.font.Index(' ')
	spacewidth := raster.Fix32(f.font.HMetric(f.scale, space).AdvanceWidth) << 2
	return raster.Fix32(width * int(spacewidth))
}

// recalc recalculates scale and bounds values from the font size, screen
// resolution and font metrics, and invalidates the glyph cache.
func (f *ttf) recalc() {
	f.scale = int32(f.size * f.dpi * (64.0 / 72.0))
	if f.font == nil {
		f.r.SetBounds(0, 0)
	} else {
		// Set the rasterizer's bounds to be big enough to handle the largest glyph.
		b := f.font.Bounds(f.scale)
		xmin := +int(b.XMin) >> 6
		ymin := -int(b.YMax) >> 6
		xmax := +int(b.XMax+63) >> 6
		ymax := -int(b.YMin-63) >> 6
		f.r.SetBounds(xmax-xmin, ymax-ymin)
	}
	for i := range f.cache {
		f.cache[i] = cacheEntry{}
	}
	f.height = int(f.size * (f.dpi / 72.0))
}

// drawContour draws the given closed contour with the given offset.
func (f *ttf) drawContour(ps []truetype.Point, dx, dy raster.Fix32) {
	if len(ps) == 0 {
		return
	}

	// The low bit of each point's Flags value is whether the point is on the
	// curve. Truetype fonts only have quadratic Bézier curves, not cubics.
	// Thus, two consecutive off-curve points imply an on-curve point in the
	// middle of those two.
	//
	// See http://chanae.walon.org/pub/ttf/ttf_glyphs.htm for more details.

	// ps[0] is a truetype.Point measured in FUnits and positive Y going
	// upwards. start is the same thing measured in fixed point units and
	// positive Y going downwards, and offset by (dx, dy).
	start := raster.Point{
		X: dx + raster.Fix32(ps[0].X<<2),
		Y: dy - raster.Fix32(ps[0].Y<<2),
	}
	others := []truetype.Point(nil)
	if ps[0].Flags&0x01 != 0 {
		others = ps[1:]
	} else {
		last := raster.Point{
			X: dx + raster.Fix32(ps[len(ps)-1].X<<2),
			Y: dy - raster.Fix32(ps[len(ps)-1].Y<<2),
		}
		if ps[len(ps)-1].Flags&0x01 != 0 {
			start = last
			others = ps[:len(ps)-1]
		} else {
			start = raster.Point{
				X: (start.X + last.X) / 2,
				Y: (start.Y + last.Y) / 2,
			}
			others = ps
		}
	}
	f.r.Start(start)
	q0, on0 := start, true
	for _, p := range others {
		q := raster.Point{
			X: dx + raster.Fix32(p.X<<2),
			Y: dy - raster.Fix32(p.Y<<2),
		}
		on := p.Flags&0x01 != 0
		if on {
			if on0 {
				f.r.Add1(q)
			} else {
				f.r.Add2(q0, q)
			}
		} else {
			if on0 {
				// No-op.
			} else {
				mid := raster.Point{
					X: (q0.X + q.X) / 2,
					Y: (q0.Y + q.Y) / 2,
				}
				f.r.Add2(q0, mid)
			}
		}
		q0, on0 = q, on
	}
	// Close the curve.
	if on0 {
		f.r.Add1(start)
	} else {
		f.r.Add2(q0, start)
	}
}

// rasterize returns the advance width, glyph mask and integer-pixel offset
// to render the given glyph at the given sub-pixel offsets.
// The 24.8 fixed point arguments fx and fy must be in the range [0, 1).
func (f *ttf) rasterize(glyph truetype.Index, fx, fy raster.Fix32) (
	raster.Fix32, *image.Alpha, image.Point, error) {

	if err := f.glyphBuf.Load(f.font, f.scale, glyph, truetype.Hinting(truetype.FullHinting)); err != nil {
		return 0, nil, image.Point{}, err
	}
	// Calculate the integer-pixel bounds for the glyph.
	xmin := int(fx+raster.Fix32(f.glyphBuf.B.XMin<<2)) >> 8
	ymin := int(fy-raster.Fix32(f.glyphBuf.B.YMax<<2)) >> 8
	xmax := int(fx+raster.Fix32(f.glyphBuf.B.XMax<<2)+0xff) >> 8
	ymax := int(fy-raster.Fix32(f.glyphBuf.B.YMin<<2)+0xff) >> 8
	if xmin > xmax || ymin > ymax {
		return 0, nil, image.Point{}, errors.New("freetype: negative sized glyph")
	}
	// A TrueType's glyph's nodes can have negative co-ordinates, but the
	// rasterizer clips anything left of x=0 or above y=0. xmin and ymin
	// are the pixel offsets, based on the font's FUnit metrics, that let
	// a negative co-ordinate in TrueType space be non-negative in
	// rasterizer space. xmin and ymin are typically <= 0.
	fx += raster.Fix32(-xmin << 8)
	fy += raster.Fix32(-ymin << 8)
	// Rasterize the glyph's vectors.
	f.r.Clear()
	e0 := 0
	for _, e1 := range f.glyphBuf.End {
		f.drawContour(f.glyphBuf.Point[e0:e1], fx, fy)
		e0 = e1
	}
	a := image.NewAlpha(image.Rect(0, 0, xmax-xmin, ymax-ymin))
	f.r.Rasterize(raster.NewAlphaSrcPainter(a))
	return raster.Fix32(f.glyphBuf.AdvanceWidth << 2), a, image.Point{xmin, ymin}, nil
}

// glyph returns the advance width, glyph mask and integer-pixel offset to
// render the given glyph at the given sub-pixel point. It is a cache for the
// rasterize method. Unlike rasterize, p's co-ordinates do not have to be in
// the range [0, 1).
func (f *ttf) glyph(glyph truetype.Index, p raster.Point) (
	raster.Fix32, *image.Alpha, image.Point, error) {

	// Split p.X and p.Y into their integer and fractional parts.
	ix, fx := int(p.X>>8), p.X&0xff
	iy, fy := int(p.Y>>8), p.Y&0xff
	// Calculate the index t into the cache array.
	tg := int(glyph) % nGlyphs
	tx := int(fx) / (256 / nXFractions)
	ty := int(fy) / (256 / nYFractions)
	t := ((tg*nXFractions)+tx)*nYFractions + ty
	// Check for a cache hit.
	if e := f.cache[t]; e.valid && e.glyph == glyph {
		return e.advanceWidth, e.mask, e.offset.Add(image.Point{ix, iy}), nil
	}
	// Rasterize the glyph and put the result into the cache.
	advanceWidth, mask, offset, err := f.rasterize(glyph, fx, fy)
	if err != nil {
		return 0, nil, image.Point{}, err
	}
	f.cache[t] = cacheEntry{true, glyph, advanceWidth, mask, offset}
	return advanceWidth, mask, offset.Add(image.Point{ix, iy}), nil
}
