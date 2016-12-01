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

	"sigint.ca/graphics/editor"
	"sigint.ca/graphics/editor/address"
)

var (
	savedPath   string
	currentPath string

	tagWidget  *widget
	tagHeight  int
	mainWidget *widget

	win         screen.Window
	winSize             = image.Pt(1024, 768)
	pixelsPerPt float32 = 1

	borderCol = color.RGBA{R: 115, G: 115, B: 190, A: 255}
)

func init() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [file]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 1 {
		savedPath = flag.Arg(0)
		currentPath = flag.Arg(0)
	} else if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}

	loadFont()
}

func main() {
	driver.Main(func(scr screen.Screen) {
		var err error
		win, err = scr.NewWindow(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer win.Release()

		updateFont()
		m := fontFace.Metrics()
		tagHeight = (m.Ascent + m.Descent).Round()

		// set up the main editor widget
		sz, pt := image.Pt(winSize.X, winSize.Y-tagHeight), image.Pt(0, tagHeight+1)
		mainWidget = newWidget(scr, sz, pt, editor.AcmeYellowTheme, fontFace)
		defer mainWidget.release()

		// load file into main editor widget
		loadMain(savedPath)

		// set up the tag widget
		sz, pt = image.Pt(winSize.X, tagHeight), image.ZP
		tagWidget = newWidget(scr, sz, pt, editor.AcmeBlueTheme, fontFace)
		defer tagWidget.release()

		// populate the tag
		tagWidget.ed.Load([]byte(currentPath + " "))
		updateTag()
		end := tagWidget.ed.Buffer.LastAddress()
		tagWidget.ed.Dot = address.Selection{From: end, To: end}

		// set up B2 and B3 actions
		tagWidget.ed.B2Action = executeCmd
		mainWidget.ed.B2Action = executeCmd
		tagWidget.ed.B3Action = findInEditor
		mainWidget.ed.B3Action = findInEditor

		widgets := []*widget{
			tagWidget,
			mainWidget,
		}
		selected := mainWidget

		var lastSize image.Point

		for {
			e := win.NextEvent()
			//log.Printf("event: %#v", e)
			switch e := e.(type) {
			case key.Event:
				if e.Direction == key.DirPress && e.Modifiers == key.ModMeta {
					switch e.Code {
					case key.CodeS:
						save()
					case key.CodeQ:
						return
					}
				}

				if e.Direction == key.DirPress || e.Direction == key.DirNone {
					selected.ed.SendKeyEvent(e)
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
						selected = w
					}
				}
				e.Pos = e.Pos.Sub(selected.r.Min)

				selected.ed.SendMouseEvent(e)
				win.Send(paint.Event{})

			case paint.Event:
				if lastSize != winSize {
					lastSize = winSize
					resize(scr)
				}

				dirty := false
				updateTag()

				if mainWidget.ed.Dirty() || mainWidget.dirty {
					dirty = true
					mainWidget.redraw()
				}

				if tagWidget.ed.Dirty() || tagWidget.dirty {
					dirty = true
					tagWidget.redraw()
				}

				// redraw screen if any widgets changed
				if dirty {
					win.Fill(image.Rectangle{Max: winSize}, borderCol, screen.Src)
					for _, w := range widgets {
						win.Copy(w.r.Min, w.tx, w.tx.Bounds(), screen.Src, nil)
					}
					win.Publish()
				}

			case size.Event:
				if e.PixelsPerPt != pixelsPerPt {
					pixelsPerPt = e.PixelsPerPt
					updateFont()

					tagWidget.ed.SetFont(fontFace)
					m := fontFace.Metrics()
					tagHeight = (m.Ascent + m.Descent).Round()

					mainWidget.ed.SetFont(fontFace)
				}
				winSize = e.Size()

			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			}
		}
	})
}

func resize(s screen.Screen) {
	tagWidget.resize(s, image.Pt(winSize.X, tagHeight), image.ZP)
	mainWidget.resize(s, image.Pt(winSize.X, winSize.Y-tagHeight), image.Pt(0, tagHeight+1))
}
