// +build ignore

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
	"golang.org/x/mobile/event/size"
)

const textLeft = `(Widget #1)

Thanks for trying this example! Please note that sigint.ca/graphics/editor is
still a work-in-progress; the API may change.

Features:
- typing
- scrolling
- sweeping
- cut/copy/paste
- undo/redo
- acme style double-click selection
- resizing

Planned:
- scrollbar
- search
- configurable middle/right click actions
`

const textRight = "(Widget #2)\n"

var (
	left, right, selected    *editor.Editor
	rLeft, rRight, rSelected image.Rectangle
)

func main() {
	width, height := 2001, 1000

	rLeft = image.Rect(0, 0, width/2, height)
	rRight = image.Rect(width/2+1, 0, width, height)

	left = editor.NewEditor(rLeft.Size(), basicfont.Face7x13, editor.AcmeBlueTheme)
	left.Load([]byte(textLeft))
	right = editor.NewEditor(rRight.Size(), basicfont.Face7x13, editor.AcmeYellowTheme)
	right.Load([]byte(textRight))

	selected = left
	rSelected = rLeft

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
					sel(e2Pt(e))
				}
				e.X -= float32(rSelected.Min.X)
				e.Y -= float32(rSelected.Min.Y)
				if e.Direction == mouse.DirPress || e.Direction == mouse.DirNone {
					selected.SendMouseEvent(e)
					w.Send(paint.Event{})
				}

			case mouse.ScrollEvent:
				sel(e2Pt(e.Event))
				selected.SendScrollEvent(e)
				w.Send(paint.Event{})

			case paint.Event:
				if left.Dirty() || right.Dirty() {
					w.Fill(image.Rect(0, 0, width, height), color.Black, draw.Src)
					w.Upload(rLeft.Min, left, left.Bounds())
					w.Upload(rRight.Min, right, right.Bounds())
					w.Publish()
				}

			case size.Event:
				resize(e.Size())
				w.Send(paint.Event{})

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

func sel(pt image.Point) {
	if pt.In(rLeft) {
		selected = left
		rSelected = rLeft
		left.SetOpts(editor.AcmeBlueTheme)
		right.SetOpts(editor.AcmeYellowTheme)
	} else if pt.In(rRight) {
		selected = right
		rSelected = rRight
		right.SetOpts(editor.AcmeBlueTheme)
		left.SetOpts(editor.AcmeYellowTheme)
	}
}

func resize(size image.Point) {
	rLeft = image.Rect(0, 0, size.X/2, size.Y)
	rRight = image.Rect(size.X/2+1, 0, size.X, size.Y)
	left.Resize(rLeft.Size())
	right.Resize(rRight.Size())
}
