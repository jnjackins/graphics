// this file was adapted from the freetype package at
// https://githued.com/jnjackins/freetype-go.

package editor

import (
	"image"
	"image/draw"

	"sigint.ca/graphics/editor/internal/text"

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
func (f fontface) draw(dst draw.Image, pt image.Point, l *text.Line) {
	if l.Adv == nil {
		l.Adv = make([]int16, 1, l.RuneCount()+1)
	} else {
		l.Adv = l.Adv[0:1]
	}
	l.Adv[0] = int16(0)
	dot := fixed.P(pt.X, pt.Y+f.height)
	for i, r := range l.String() {
		tab := r == '\t'
		if tab {
			r = ' '
		}
		dr, mask, maskp, advance, ok := f.face.Glyph(dot, r)
		if !ok {
			panic("internal error: draw font: bad glyph")
		}
		if tab {
			advance *= tabwidth
		}
		dot.X += advance
		draw.DrawMask(dst, dr, image.Black, dr.Min, mask, maskp, draw.Over)
		l.Adv = append(l.Adv, l.Adv[i]+int16(advance>>6))
	}
}

// measure returns the cumulative advance in pixel for each glyph in s,
// beginning from the pixel value start.
func (f fontface) measure(l *text.Line) {
	if l.Adv == nil {
		l.Adv = make([]int16, 1, l.RuneCount()+1)
	} else {
		l.Adv = l.Adv[0:1]
	}
	l.Adv[0] = 0
	for i, r := range l.String() {
		tab := r == '\t'
		if tab {
			r = ' '
		}
		advance, ok := f.face.GlyphAdvance(r)
		if !ok {
			panic("internal error: measure font: bad glyph")
		}
		if tab {
			advance *= tabwidth
		}
		l.Adv = append(l.Adv, l.Adv[i]+int16(advance>>6))
	}
}
