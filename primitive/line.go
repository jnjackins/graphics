package primitive

import (
	"image"
	"image/color"
	"image/draw"
)

func Line(dst draw.Image, c color.Color, p1, p2 image.Point) {
	x0, y0, x1, y1 := p1.X, p1.Y, p2.X, p2.Y
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}
	dx := x1 - x0
	dy := y1 - y0
	if dx == 0 {
		vline(dst, c, p1, p2)
		return
	}
	var err float64
	slope := abs(float64(dy) / float64(dx))
	y := y0
	yDir := sign(y1 - y0)
	for x := x0; x <= x1; x++ {
		dst.Set(x, y, c)
		err = err + slope
		for err >= 0.5 && !exceeded(y, y1, yDir) {
			dst.Set(x, y, c)
			y += yDir
			err -= 1.0
		}
	}
}

func vline(dst draw.Image, c color.Color, p1, p2 image.Point) {
	y0, y1 := p1.Y, p2.Y
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y < y1; y++ {
		dst.Set(p1.X, y, c)
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -1 * v
	}
	return v
}

func sign(v int) int {
	if v < 0 {
		return -1
	}
	return 1
}

func exceeded(from, to, dir int) bool {
	if dir < 0 {
		return to > from
	}
	return from > to
}
