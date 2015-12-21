// this file was adapted from the freetype package at
// https://github.com/jnjackins/freetype-go.

package text

import (
	"image"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const tabwidth = 4

type fontface struct {
	face   font.Face
	height int
}

// draw draws s onto dst starting at pt, up to a maximum length of maxw pixels.
// It returns a slice of x-coords for each rune, as well as a string containing
// the portion of s which exceeded maxw and was not drawn.
func (f fontface) draw(dst draw.Image, pt image.Point, s []rune, maxw int) []int {
	px := make([]int, 1, len(s)+1)
	px[0] = pt.X
	dot := fixed.P(pt.X, pt.Y+f.height)
	var tab bool
	for _, r := range s {
		if r == '\t' {
			tab = true
			r = ' '
		}
		dr, mask, maskp, advance, ok := f.face.Glyph(dot, r)
		if !ok {
			panic("NOT OK")
		}
		if tab {
			advance *= tabwidth
		}
		dot.X += advance
		if maxw > 0 && int(dot.X>>6) > maxw {
			return px
		}
		draw.DrawMask(dst, dr, image.Black, dr.Min, mask, maskp, draw.Over)
		px = append(px, int(dot.X>>6))
	}
	return px
}

// measure returns a slice of character positions in pixels for s,
// beginning from the pixel value start.
func (f fontface) measure(start int, s []rune) []int {
	px := make([]int, 1, len(s)+1)
	px[0] = start
	dot := fixed.I(start)
	for _, r := range s {
		advance, ok := f.face.GlyphAdvance(r)
		if !ok {
			panic("NOT OK")
		}
		dot += advance
		px = append(px, int(dot>>6))
	}
	return px
}
