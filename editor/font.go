// this file was adapted from the freetype package at
// https://githued.com/jnjackins/freetype-go.

package editor

import (
	"image"
	"image/draw"
	"unicode"

	"sigint.ca/graphics/editor/internal/text"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const tabwidth = 4

type fontface struct {
	face   font.Face
	height int
}

func mkFont(face font.Face) fontface {
	bounds, _, _ := face.GlyphBounds('X')
	height := int(1.5 * float64(bounds.Max.Y>>6-bounds.Min.Y>>6))
	return fontface{
		face:   face,
		height: height,
	}
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
	dot := fixed.P(pt.X, pt.Y+f.height-6)
	var i int
	for _, r := range l.String() {
		tab := r == '\t'
		if tab {
			// for the sake of variable-width fonts, try to pick a font with an
			// average width
			r = '_'
		}
		dr, mask, maskp, advance, ok := f.face.Glyph(dot, r)
		if !ok {
			dr, mask, maskp, advance, ok = f.face.Glyph(dot, unicode.ReplacementChar)
			if !ok {
				panic("internal error")
			}
		}
		if tab {
			advance *= tabwidth
		} else {
			draw.DrawMask(dst, dr, image.Black, dr.Min, mask, maskp, draw.Over)
		}
		dot.X += advance
		l.Adv = append(l.Adv, l.Adv[i]+int16(advance>>6))
		i++
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
	var i int
	for _, r := range l.String() {
		tab := r == '\t'
		if tab {
			// for the sake of variable-width fonts, try to pick a font with an
			// average width
			r = '_'
		}
		advance, ok := f.face.GlyphAdvance(r)
		if !ok {
			advance, ok = f.face.GlyphAdvance(unicode.ReplacementChar)
			if !ok {
				panic("internal error")
			}
		}
		if tab {
			advance *= tabwidth
		}
		l.Adv = append(l.Adv, l.Adv[i]+int16(advance>>6))
		i++
	}
}
