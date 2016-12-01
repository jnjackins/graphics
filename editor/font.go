package editor

import (
	"image"
	"image/draw"
	"unicode"

	"sigint.ca/graphics/editor/text"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// SetFont sets the Editor's font face to face.
func (ed *Editor) SetFont(face font.Face) {
	ed.font = face

	ed.fontHeight = (face.Metrics().Ascent + face.Metrics().Descent).Round()
	ed.dirty = true
}

const tabwidth = 4

// drawLine draws l onto dst starting at pt and updates the values in l.Adv.
func (ed *Editor) drawLine(dst draw.Image, pt image.Point, l *text.Line, src *image.Uniform) {
	f := ed.font

	if l.Adv == nil {
		l.Adv = make([]fixed.Int26_6, 1, l.RuneCount()+1)
	} else {
		l.Adv = l.Adv[0:1]
	}
	dot := fixed.P(pt.X, pt.Y)
	dot.Y += f.Metrics().Ascent
	var i int
	for _, r := range l.String() {
		tab := r == '\t'
		if tab {
			// in case of variable-width fonts, try to pick a font with an
			// average width
			r = '_'
		}
		// try to draw the glyph
		dr, mask, maskp, advance, ok := f.Glyph(dot, r)
		if !ok {
			// try to draw unicode.ReplacementChar
			dr, mask, maskp, advance, ok = f.Glyph(dot, unicode.ReplacementChar)
			if !ok {
				// last ditch effort to draw something
				dr, mask, maskp, advance, ok = f.Glyph(dot, '?')
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
		l.Adv = append(l.Adv, l.Adv[i]+advance)
		i++
	}
}

// measureLine updates the values in l.Adv.
func (ed *Editor) measureLine(l *text.Line) {
	f := ed.font

	if l.Adv == nil {
		l.Adv = make([]fixed.Int26_6, 1, l.RuneCount()+1)
	} else {
		l.Adv = l.Adv[0:1]
	}
	var i int
	for _, r := range l.String() {
		tab := r == '\t'
		if tab {
			// for the sake of variable-width fonts, try to pick a font with an
			// average width
			r = '_'
		}
		advance, ok := f.GlyphAdvance(r)
		if !ok {
			advance, ok = f.GlyphAdvance(unicode.ReplacementChar)
			if !ok {
				advance, ok = f.GlyphAdvance('?')
				if !ok {
					panic("internal error")
				}
			}
		}
		if tab {
			advance *= tabwidth
		}
		l.Adv = append(l.Adv, l.Adv[i]+advance)
		i++
	}
}
