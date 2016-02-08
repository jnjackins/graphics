package editor

import (
	"image"
	"image/draw"
	"unicode"

	"sigint.ca/graphics/editor/internal/text"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// SetFont sets the Editor's font face to face.
func (ed *Editor) SetFont(face font.Face) {
	ed.font = mkFont(face)
	ed.dirty = true
}

const tabwidth = 4

type fontface struct {
	face   font.Face
	height int
}

func mkFont(face font.Face) fontface {
	bounds, _, _ := face.GlyphBounds('|')
	height := int(1.33*float64(bounds.Max.Y>>6-bounds.Min.Y>>6)) + 1
	return fontface{
		face:   face,
		height: height,
	}
}

// draw draws s onto dst starting at pt. It returns the cumulative advance
// in pixels of each glyph.
func (f fontface) draw(dst draw.Image, pt image.Point, l *text.Line, src *image.Uniform) {
	if l.Adv == nil {
		l.Adv = make([]int16, 1, l.RuneCount()+1)
	} else {
		l.Adv = l.Adv[0:1]
	}
	dot := fixed.P(pt.X, pt.Y+f.height-(f.height/4))
	for i, r := range l.Runes() {
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
				dr, mask, maskp, advance, ok = f.face.Glyph(dot, '?')
				if !ok {
					panic("internal error")
				}
			}
		}
		if tab {
			advance *= tabwidth
		} else {
			draw.DrawMask(dst, dr, src, dr.Min, mask, maskp, draw.Over)
		}
		dot.X += advance
		l.Adv = append(l.Adv, l.Adv[i]+int16(advance>>6))
	}
}

// measure returns the cumulative advance in pixels for each glyph in s.
func (f fontface) measure(l *text.Line) {
	if l.Adv == nil {
		l.Adv = make([]int16, 1, l.RuneCount()+1)
	} else {
		l.Adv = l.Adv[0:1]
	}
	for i, r := range l.Runes() {
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
				advance, ok = f.face.GlyphAdvance('?')
				if !ok {
					panic("internal error")
				}
			}
		}
		if tab {
			advance *= tabwidth
		}
		l.Adv = append(l.Adv, l.Adv[i]+int16(advance>>6))
	}
}
