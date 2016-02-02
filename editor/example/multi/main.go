// +build ignore

package main

import (
	"image"
	"image/color"
	"log"
	"sync"

	"golang.org/x/exp/shiny/driver"
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
	- search (right-click a selection)
	- autoindent

Planned:
	- scrollbar
`
const text2 = "(Widget #2)\n"
const text3 = "(Widget #3)\n"
const text4 = "(Widget #4)\n"

var width, height = 1001, 1001

func main() {
	driver.Main(func(scr screen.Screen) {

		opts := screen.NewWindowOptions{Width: width, Height: height}
		win, err := scr.NewWindow(&opts)
		if err != nil {
			log.Fatal(err)
		}
		defer win.Release()

		sz := image.Pt(width/2, height/2)
		widgets := []*widget{
			newWidget(scr, sz, image.ZP, text1),
			newWidget(scr, sz, image.Pt((width/2)+1, 0), text2),
			newWidget(scr, sz, image.Pt(0, (height/2)+1), text3),
			newWidget(scr, sz, image.Pt((width/2)+1, (height/2)+1), text4),
		}

		selected, _ := sel(image.ZP, widgets) // select the top left widget to start

		for {
			switch e := win.NextEvent().(type) {
			case key.Event:
				if e.Direction == key.DirPress || e.Direction == key.DirNone {
					selected.ed.SendKeyEvent(e)
					win.Send(paint.Event{})
				}

			case mouse.Event:
				if e.Direction == mouse.DirPress {
					if w, ok := sel(e2Pt(e), widgets); ok {
						selected = w
					}
				}
				e.X -= float32(selected.r.Min.X)
				e.Y -= float32(selected.r.Min.Y)

				selected.ed.SendMouseEvent(e)
				win.Send(paint.Event{})

			case mouse.ScrollEvent:
				if w, ok := sel(e2Pt(e.Event), widgets); ok {
					selected = w
				}
				selected.ed.SendScrollEvent(e)
				win.Send(paint.Event{})

			case paint.Event:
				dirty := false
				var wg sync.WaitGroup

				// redraw any widgets that changed
				for _, w := range widgets {
					if w.ed.Dirty() {
						dirty = true
						wg.Add(1)
						go func(w *widget) {
							*w.buf.RGBA() = *w.ed.RGBA()
							w.tx.Upload(w.r.Min, w.buf, w.buf.Bounds())
							wg.Done()
						}(w)
					}
				}
				wg.Wait()

				// redraw screen if any widgets changed
				if dirty {
					r := image.Rect(0, 0, width, height)
					win.Fill(r, color.Black, screen.Src)
					for _, w := range widgets {
						wg.Add(1)
						go func(w *widget) {
							win.Copy(w.r.Min, w.tx, w.tx.Bounds(), screen.Src, nil)
							wg.Done()
						}(w)
					}
					wg.Wait()
					win.Publish()
				}

			case size.Event:
				resize(scr, e.Size(), widgets)
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

func resize(s screen.Screen, size image.Point, widgets []*widget) {
	width, height = size.X, size.Y
	wSize := image.Pt(width/2, height/2)
	widgets[0].resize(s, wSize, image.ZP)
	widgets[1].resize(s, wSize, image.Pt(width/2+1, 0))
	widgets[2].resize(s, wSize, image.Pt(0, height/2+1))
	widgets[3].resize(s, wSize, image.Pt(width/2+1, height/2+1))
}
