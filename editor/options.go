package editor

import (
	"image"
	"image/color"
	"image/draw"
)

// SetOpts reconfigures the Editor according to opts.
func (ed *Editor) SetOpts(opts *OptionSet) {
	ed.opts = opts
	ed.dirty = true
}

type OptionSet struct {
	Text       *image.Uniform
	BG1        *image.Uniform
	BG2        *image.Uniform
	Sel        *image.Uniform
	Cursor     func(height int) image.Image
	AutoIndent bool
	ScrollBar  bool
}

func acmeCursor(bg image.Image) func(height int) image.Image {
	fn := func(height int) image.Image {
		// the squares at the top and bottom of the cursor should
		// be 3x3 points, and the center bar should be one point.
		// we don't have any DPI info (TODO), so try to come up with
		// something reasonable based on height.
		scale := height/(6*3) + 1

		cursor := image.NewRGBA(image.Rect(0, 0, 3*scale, height))
		draw.Draw(cursor, cursor.Rect, image.Black, image.ZP, draw.Src)
		draw.Draw(cursor, image.Rect(0, 3*scale, 1*scale, height-(3*scale)), bg, image.ZP, draw.Src)
		draw.Draw(cursor, image.Rect(2*scale, 3*scale, 3*scale, height-(3*scale)), bg, image.ZP, draw.Src)
		return cursor
	}

	return fn
}

var AcmeYellowTheme = &OptionSet{
	Text:       image.Black,
	BG1:        image.NewUniform(color.RGBA{R: 0xFF, G: 0xFF, B: 0xEA, A: 0xFF}),
	BG2:        image.NewUniform(color.RGBA{R: 0xA0, G: 0xA0, B: 0x4B, A: 0xFF}),
	Sel:        image.NewUniform(color.RGBA{R: 0xEE, G: 0xEE, B: 0x9E, A: 0xFF}),
	Cursor:     acmeCursor(image.NewUniform(color.RGBA{R: 0xFF, G: 0xFF, B: 0xEA, A: 0xFF})),
	AutoIndent: true,
	ScrollBar:  true,
}

var AcmeBlueTheme = &OptionSet{
	Text:       image.Black,
	BG1:        image.NewUniform(color.RGBA{R: 0xEA, G: 0xFF, B: 0xFF, A: 0xFF}),
	BG2:        image.NewUniform(color.RGBA{R: 0x88, G: 0x88, B: 0xCC, A: 0xFF}),
	Sel:        image.NewUniform(color.RGBA{R: 0x9F, G: 0xEB, B: 0xEA, A: 0xFF}),
	Cursor:     acmeCursor(image.NewUniform(color.RGBA{R: 0xEA, G: 0xFF, B: 0xFF, A: 0xFF})),
	AutoIndent: true,
	ScrollBar:  true,
}

var simpleCursor = func(height int) image.Image {
	cursor := image.NewRGBA(image.Rect(0, 0, 3, height))
	draw.Draw(cursor, image.Rect(1, 1, 2, height-1), image.Black, image.ZP, draw.Src)
	return cursor
}

var SimpleTheme = &OptionSet{
	Text:   image.Black,
	BG1:    image.White,
	BG2:    image.NewUniform(color.Gray{Y: 0xA0}),
	Sel:    image.NewUniform(color.RGBA{R: 0x90, G: 0xB0, B: 0xD0, A: 0xFF}),
	Cursor: simpleCursor,
}
