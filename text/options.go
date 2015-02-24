package text

import (
	"image"
	"image/color"
	"image/draw"
)

type OptionSet struct {
	BGColor  color.Color
	SelColor color.Color
	Margin   image.Point
	Cursor   func(height int) image.Image
}

var AcmeTheme = OptionSet{
	BGColor:  color.RGBA{R: 0xFF, G: 0xFF, B: 0xEA, A: 0xFF},
	SelColor: color.RGBA{R: 0xEE, G: 0xEE, B: 0x9E, A: 0xFF},
	Margin:   image.Pt(4, 0),
	Cursor: func(height int) image.Image {
		bgCol := image.NewUniform(color.RGBA{R: 0xFF, G: 0xFF, B: 0xEA, A: 0xFF})
		cursor := image.NewRGBA(image.Rect(0, 0, 3, height))
		draw.Draw(cursor, cursor.Rect, image.Black, image.ZP, draw.Src)
		h := height - 3 // TODO magic 3
		draw.Draw(cursor, image.Rect(0, 3, 1, h), bgCol, image.ZP, draw.Src)
		draw.Draw(cursor, image.Rect(2, 3, 3, h), bgCol, image.ZP, draw.Src)
		return cursor
	},
}
