// this file was adapted from the freetype package at
// https://githued.com/jnjackins/freetype-go.

package editor

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

// draw draws s onto dst starting at pt. It returns the cumulative advance
// in pixels of each glyph.
func (f fontface) draw(dst draw.Image, pt image.Point, s string) []int16 {
	px := make([]int16, 1, len(s)+1)
	px[0] = int16(pt.X) // TODO: check for overflow
	dot := fixed.P(pt.X, pt.Y+f.height)
	for _, r := range s {
		tab := r == '\t'
		if tab {
			r = ' '
		}
		dr, mask, maskp, advance, ok := f.face.Glyph(dot, r)
		if !ok {
			panic("internal error: draw: bad glyph")
		}
		if tab {
			advance *= tabwidth
		}
		dot.X += advance
		draw.DrawMask(dst, dr, image.Black, dr.Min, mask, maskp, draw.Over)
		px = append(px, int16(dot.X>>6))
	}
	return px
}

// measure returns the cumulative advance in pixel for each glyph in s,
// beginning from the pixel value start.
func (f fontface) measure(start int, s string) []int16 {
	px := make([]int16, 1, len(s)+1)
	px[0] = int16(start) // TODO: check for overflow
	dot := fixed.I(start)
	for _, r := range s {
		tab := r == '\t'
		if tab {
			r = ' '
		}
		advance, ok := f.face.GlyphAdvance(r)
		if !ok {
			panic("internal error: measure: GlyphAdvance not ok")
		}
		if tab {
			advance *= tabwidth
		}
		dot += advance
		px = append(px, int16(dot>>6))
	}
	return px
}
