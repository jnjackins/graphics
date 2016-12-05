package editor

import (
	"image"
	"image/draw"
	"unicode"
	"unicode/utf8"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// TODO: make this configurable
const tabstop = 4

// SetFont sets the Editor's font face to face.
func (ed *Editor) SetFont(face font.Face) {
	if face == nil {
		panic("nil font")
	}
	ed.font = face

	ed.fontHeight = (face.Metrics().Ascent + face.Metrics().Descent).Round()
	if ed.fontHeight == 0 {
		panic("bad font")
	}

	// fix tabstop width
	advance, ok := ed.font.GlyphAdvance('0')
	if !ok {
		panic("couldn't get glyph advance")
	}
	ed.tabwidth = advance * tabstop

	ed.dirty = true
}

// drawLine draws s onto dst starting at pt.
func (ed *Editor) drawString(dst draw.Image, pt image.Point, s string, src *image.Uniform) {
	dot := fixed.P(pt.X, pt.Y)
	dot.Y += ed.font.Metrics().Ascent

	// used to calculate tabstop locations (pt.X may not start from 0)
	var width fixed.Int26_6

	for _, r := range s {
		// handle tabstops
		if r == '\t' {
			off := ed.tabwidth - width%ed.tabwidth
			dot.X += off
			width += off
			continue
		}
		// try to draw the glyph
		dr, mask, maskp, advance, ok := ed.font.Glyph(dot, r)
		if !ok {
			// try to draw unicode.ReplacementChar
			dr, mask, maskp, advance, ok = ed.font.Glyph(dot, unicode.ReplacementChar)
			if !ok {
				// last ditch effort to draw something
				dr, mask, maskp, advance, ok = ed.font.Glyph(dot, '?')
				if !ok {
					panic("couldn't draw glyph")
				}
			}
		}
		draw.DrawMask(dst, dr, src, dr.Min, mask, maskp, draw.Over)
		dot.X += advance
		width += advance
	}
}

// measureString returns a slice of monotonically increasing pixel offsets
// for runes in s.
func (ed *Editor) measureString(s string) []fixed.Int26_6 {
	adv := make([]fixed.Int26_6, 1, utf8.RuneCountInString(s)+1)
	for _, r := range s {
		// handle tabstops
		if r == '\t' {
			last := adv[len(adv)-1]
			adv = append(adv, last+ed.tabwidth-last%ed.tabwidth)
			continue
		}
		advance, ok := ed.font.GlyphAdvance(r)
		if !ok {
			advance, ok = ed.font.GlyphAdvance(unicode.ReplacementChar)
			if !ok {
				advance, ok = ed.font.GlyphAdvance('?')
				if !ok {
					panic("couldn't get glyph advance")
				}
			}
		}
		adv = append(adv, adv[len(adv)-1]+advance)
	}
	return adv
}
