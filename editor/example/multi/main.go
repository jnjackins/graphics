// +build ignore

package main

import (
	"image"
	"image/color"
	"image/draw"
	"log"

	"sigint.ca/graphics/editor"

	"golang.org/x/exp/shiny/driver/gldriver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

const text1 = `(Widget #1)

Thanks for trying this example!
Please note that sigint.ca/graphics/editor a work-in-progress.
The API may change.

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
- autoindent
`
const text2 = "(Widget #2)\n"
const text3 = "(Widget #3)\n"
const text4 = "(Widget #4)\n"

var width, height = 801, 801

func main() {
	gldriver.Main(func(s screen.Screen) {

		opts := screen.NewWindowOptions{Width: width, Height: height}
		win, err := s.NewWindow(&opts)
		if err != nil {
			log.Fatal(err)
		}
		defer win.Release()

		sz := image.Pt(width/2, height/2)
		widgets := []*widget{
			newWidget(s, sz, image.ZP, text1),
			newWidget(s, sz, image.Pt((width/2)+1, 0), text2),
			newWidget(s, sz, image.Pt(0, (height/2)+1), text3),
			newWidget(s, sz, image.Pt((width/2)+1, (height/2)+1), text4),
		}

		selected := sel(image.ZP, widgets) // select the top left widget to start

		win.Send(paint.Event{})

		for {
			switch e := win.NextEvent().(type) {
			case key.Event:
				if e.Code == key.CodeEscape {
					return
				}
				if e.Direction == key.DirPress || e.Direction == key.DirNone {
					selected.ed.SendKeyEvent(e)
					win.Send(paint.Event{})
				}

			case mouse.Event:
				if e.Direction == mouse.DirPress {
					selected = sel(e2Pt(e), widgets)
				}
				e.X -= float32(selected.r.Min.X)
				e.Y -= float32(selected.r.Min.Y)
				if e.Direction == mouse.DirPress || e.Direction == mouse.DirNone {
					selected.ed.SendMouseEvent(e)
					win.Send(paint.Event{})
				}

			case mouse.ScrollEvent:
				selected = sel(e2Pt(e.Event), widgets)
				selected.ed.SendScrollEvent(e)
				win.Send(paint.Event{})

			case paint.Event:
				dirty := false
				for _, w := range widgets {
					if w.ed.Dirty() {
						dirty = true
						*w.buf.RGBA() = *w.ed.RGBA()
						w.tx.Upload(w.r.Min, w.buf, w.buf.Bounds())
					}
				}
				if dirty {
					r := image.Rect(0, 0, width, height)
					win.Fill(r, color.Black, draw.Src)
					for _, w := range widgets {
						screen.Copy(win, w.r.Min, w.tx, w.tx.Bounds(), draw.Src, nil)
					}
					win.Publish()
				}

			case size.Event:
				resize(s, e.Size(), widgets)
				win.Send(paint.Event{})

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

func sel(pt image.Point, widgets []*widget) *widget {
	var selected *widget
	for _, w := range widgets {
		if pt.In(w.r) {
			selected = w
			w.ed.SetOpts(editor.AcmeBlueTheme)
		} else {
			w.ed.SetOpts(editor.AcmeYellowTheme)
		}
	}
	return selected
}

func resize(s screen.Screen, size image.Point, widgets []*widget) {
	width, height = size.X, size.Y
	wSize := image.Pt(width/2, height/2)
	widgets[0].resize(s, wSize, image.ZP)
	widgets[1].resize(s, wSize, image.Pt(width/2+1, 0))
	widgets[2].resize(s, wSize, image.Pt(0, height/2+1))
	widgets[3].resize(s, wSize, image.Pt(width/2+1, height/2+1))
}
