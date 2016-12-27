package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

var (
	scr         screen.Screen
	win         screen.Window
	winSize             = image.Pt(800, 600)
	pixelsPerPt float32 = 1

	tagHeight int
	borderCol = color.RGBA{R: 115, G: 115, B: 190, A: 255}

	panes   []*pane
	widgets []*widget
)

var (
	dflag   = flag.Bool("d", false, "Toggle debug mode.")
	dprintf = func(format string, args ...interface{}) {}
)

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [file ...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *dflag {
		dprintf = log.Printf
	}

	loadFont()

	driver.Main(func(_scr_ screen.Screen) {
		scr = _scr_

		var err error
		win, err = scr.NewWindow(&screen.NewWindowOptions{
			Width:  winSize.X,
			Height: winSize.Y,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer win.Release()

		updateFont(size.Event{ScaleFactor: 1})
		m := fontFace.Metrics()
		tagHeight = (m.Ascent + m.Descent).Round()

		paths := []string{""}
		if flag.NArg() >= 1 {
			paths = flag.Args()
		}
		for _, path := range paths {
			addPane(path, nil)
		}

		selWidget := panes[0].main

		var lastSize image.Point
		for {
			e := win.NextEvent()
			//dprintf("event: %#v", e)
			switch e := e.(type) {
			case key.Event:
				if e.Direction == key.DirPress && e.Modifiers == key.ModMeta {
					switch e.Code {
					case key.CodeS:
						selWidget.pane.save()
					case key.CodeQ:
						return
					}
				}

				if e.Direction == key.DirPress || e.Direction == key.DirNone {
					selWidget.ed.SendKeyEvent(e)
					win.Send(paint.Event{})
				}

			case mouse.Event:
				if e.Modifiers == key.ModAlt {
					e.Button = mouse.ButtonMiddle
				} else if e.Modifiers == key.ModMeta {
					e.Button = mouse.ButtonRight
				}

				if e.Direction == mouse.DirPress || e.Button == mouse.ButtonScroll {
					if w, ok := sel(e.Pos, widgets); ok {
						selWidget = w
					}
				}
				e.Pos = e.Pos.Sub(selWidget.r.Min)

				selWidget.ed.SendMouseEvent(e)
				win.Send(paint.Event{})

			case paint.Event:
				if lastSize != winSize {
					dprintf("resizing panes")
					lastSize = winSize
					for _, p := range panes {
						p.resize()
					}
				}

				dirty := false

				for i, p := range panes {
					// TODO: avoid this when unnecessary
					p.updateTag()

					if p.main.ed.Dirty() || p.main.dirty {
						dprintf("drawing pane %d main", i)
						dirty = true
						p.main.draw()
					}

					if p.tag.ed.Dirty() || p.tag.dirty {
						dprintf("drawing pane %d tag", i)
						dirty = true
						p.tag.draw()
					}
				}

				// redraw screen if any widgets changed
				if dirty || e.External {
					dprintf("publishing to window")
					win.Fill(image.Rectangle{Max: winSize}, borderCol, screen.Src)
					for _, w := range widgets {
						win.Copy(w.r.Min, w.tx, w.tx.Bounds(), screen.Src, nil)
					}
					win.Publish()
				}

			case size.Event:
				if e.PixelsPerPt != pixelsPerPt {
					pixelsPerPt = e.PixelsPerPt
					updateFont(e)

					for _, p := range panes {
						p.tag.ed.SetFont(fontFace)
						m := fontFace.Metrics()
						tagHeight = (m.Ascent + m.Descent).Round()

						p.main.ed.SetFont(fontFace)
					}
				}
				winSize = e.Size()

			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}

			case error:
				log.Print(e)
			}
		}
	})
}
