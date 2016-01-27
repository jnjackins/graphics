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
)

var (
	filename   string
	tagWidget  *widget
	mainWidget *widget
	winSize    = image.Pt(512, 512)
	borderCol  = color.RGBA{R: 115, G: 115, B: 190, A: 255}
	tagHeight  int
)

func init() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [file]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	filename = ""
	if flag.NArg() == 1 {
		filename = flag.Arg(0)
	} else if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	font, h := getfont()
	tagHeight = h

	driver.Main(func(scr screen.Screen) {
		win, err := scr.NewWindow(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer win.Release()

		pt, sz := image.Pt(winSize.X, tagHeight), image.ZP
		tagWidget = newWidget(scr, sz, pt, editor.AcmeBlueTheme, font)
		tagWidget.action = editorCommand

		sz, pt = image.Pt(winSize.X, winSize.Y-tagHeight), image.Pt(0, tagHeight+1)
		mainWidget = newWidget(scr, sz, pt, editor.AcmeYellowTheme, font)
		loadMain(filename)

		widgets := []*widget{
			tagWidget,
			mainWidget,
		}
		selected := mainWidget

		for {
			switch e := win.NextEvent().(type) {
			case key.Event:
				if e.Direction == key.DirPress && e.Modifiers == key.ModMeta && e.Code == key.CodeS {
					save()
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

				if e.Direction == mouse.DirPress {
					if w, ok := sel(e2Pt(e), widgets); ok {
						selected = w
					}
				}
				e.X -= float32(selected.r.Min.X)
				e.Y -= float32(selected.r.Min.Y)

				selected.ed.SendMouseEvent(e)

				switch e.Button {
				case mouse.ButtonMiddle:
					if e.Direction == mouse.DirRelease {
						if selected.action != nil {
							selected.action(selected.ed.GetSel())
						}
					}
				case mouse.ButtonRight:
					if e.Direction == mouse.DirRelease {
						mainWidget.ed.Search(selected.ed.GetSel())
					}
				}

				win.Send(paint.Event{})

			case mouse.ScrollEvent:
				if w, ok := sel(e2Pt(e.Event), widgets); ok {
					selected = w
				}
				selected.ed.SendScrollEvent(e)
				win.Send(paint.Event{})

			case paint.Event:
				dirty := false
				updateTag()

				if mainWidget.ed.Dirty() {
					dirty = true
					mainWidget.redraw()
				}

				if tagWidget.ed.Dirty() {
					dirty = true
					tagWidget.redraw()
				}

				// redraw screen if any widgets changed
				if dirty {
					win.Fill(image.Rectangle{Max: winSize}, borderCol, screen.Src)
					for _, w := range widgets {
						screen.Copy(win, w.r.Min, w.tx, w.tx.Bounds(), screen.Src, nil)
					}
					win.Publish()
				}

			case size.Event:
				resize(scr, e.Size())
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

func resize(s screen.Screen, size image.Point) {
	winSize = size
	tagWidget.resize(s, image.Pt(winSize.X, tagHeight), image.ZP)
	mainWidget.resize(s, image.Pt(winSize.X, winSize.Y-tagHeight), image.Pt(0, tagHeight+1))
}
