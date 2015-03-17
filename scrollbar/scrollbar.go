package scrollbar // import "sigint.ca/graphics/scrollbar"

import (
	"image"
	"image/color"
	"image/draw"
)

type Scrollbar struct {
	img             *image.RGBA
	bg, fg          *image.Uniform
	visible, actual image.Rectangle
}

func New(width, height int, bgCol, fgCol color.Color) *Scrollbar {
	r := image.Rect(0, 0, width, height)
	return &Scrollbar{
		img: image.NewRGBA(r),
		bg:  image.NewUniform(bgCol),
		fg:  image.NewUniform(fgCol),
	}
}

func (sb *Scrollbar) Resize(width, height int) {
	r := image.Rect(0, 0, width, height)
	*sb = Scrollbar{
		img: image.NewRGBA(r),
		bg:  sb.bg,
		fg:  sb.fg,
	}
}

func (sb *Scrollbar) sliderRect(visible, actual image.Rectangle) image.Rectangle {
	barHeight := float64(sb.img.Bounds().Dy())
	sliderHeight := int(barHeight * float64(visible.Dy()) / float64(actual.Dy()))
	sliderPos := int(barHeight * float64(visible.Min.Y) / float64(actual.Max.Y))
	sliderPos -= 3 // show a wee bit of slider when we're scrolled to the bottom, like acme

	return image.Rect(0, sliderPos, sb.img.Bounds().Dx()-1, sliderPos+sliderHeight)
}

func (sb *Scrollbar) Img(visible, actual image.Rectangle) *image.RGBA {
	if visible == sb.visible && actual == sb.actual {
		return sb.img
	}
	sb.visible = visible
	sb.actual = actual

	draw.Draw(sb.img, sb.img.Rect, sb.bg, image.ZP, draw.Src)
	draw.Draw(sb.img, sb.sliderRect(visible, actual), sb.fg, image.ZP, draw.Src)
	return sb.img
}
