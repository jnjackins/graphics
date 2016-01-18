package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"

	"sigint.ca/graphics/editor"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
)

const textLeft = `(Widget #1)

Thanks for trying this example! Please note that sigint.ca/graphics/editor is
still a work-in-progress; the API may change drastically.

Features:
- typing
- scrolling
- sweeping
- copy/paste
- undo/redo

To do:
- scrollbar
`

const textRight = `(Widget #2)`

func main() {
	width, height := 2001, 1000
	rLeft := image.Rect(0, 0, width/2, height)
	rRight := image.Rect(width/2+1, 0, width, height)
	rSelected := rLeft

	left := editor.NewEditor(rLeft.Size(), basicfont.Face7x13, editor.SimpleTheme)
	left.Load([]byte(textLeft))
	right := editor.NewEditor(rRight.Size(), basicfont.Face7x13, editor.SimpleTheme)
	right.Load([]byte(textRight))

	selected := left

	driver.Main(func(s screen.Screen) {
		opts := screen.NewWindowOptions{Width: width, Height: height}
		w, err := s.NewWindow(&opts)
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		for {
			switch e := w.NextEvent().(type) {
			case key.Event:
				if e.Code == key.CodeEscape {
					return
				}

				if e.Direction == key.DirPress || e.Direction == key.DirNone {
					selected.SendKeyEvent(e)
					w.Send(paint.Event{})
				}

			case mouse.Event:
				if e.Direction == mouse.DirPress {
					if e2Pt(e).In(rLeft) {
						selected = left
						rSelected = rLeft
					} else if e2Pt(e).In(rRight) {
						selected = right
						rSelected = rRight
					}
				}
				e.X -= float32(rSelected.Min.X)
				e.Y -= float32(rSelected.Min.Y)
				if e.Direction == mouse.DirPress || e.Direction == mouse.DirNone {
					selected.SendMouseEvent(e)
					w.Send(paint.Event{})
				}

			case mouse.ScrollEvent:
				if e.Direction == mouse.DirPress {
					if e2Pt(e.Event).In(rLeft) {
						selected = left
						rSelected = rLeft
					} else if e2Pt(e.Event).In(rRight) {
						selected = right
						rSelected = rRight
					}
				}
				selected.SendScrollEvent(e)
				w.Send(paint.Event{})

			case paint.Event:
				if left.Dirty() || right.Dirty() {
					w.Fill(image.Rect(0, 0, width, height), color.Black, draw.Src)
					w.Upload(rLeft.Min, left, left.Bounds())
					w.Upload(rRight.Min, right, right.Bounds())
					w.Publish()
				}

			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			}
		}
	})
}

func e2Pt(e mouse.Event) image.Point {
	return image.Pt(int(e.X), int(e.Y))
}
