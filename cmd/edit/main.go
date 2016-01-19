package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"sigint.ca/graphics/editor"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/plan9font"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

var (
	filename   string
	mainEditor *editor.Editor
	tagEditor  *editor.Editor
	win        screen.Window
	winr       image.Rectangle
	bgColor    = color.White
)

func init() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s file ...\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func main() {
	size := image.Pt(1024, 768)
	font := getfont()
	mainEditor = editor.NewEditor(size, font, editor.AcmeYellowTheme)

	if flag.NArg() == 1 {
		load(flag.Arg(0))
	} else if flag.NArg() > 1 {
		log.Println("multiple files not yet supported")
		flag.Usage()
		os.Exit(1)
	}

	driver.Main(func(s screen.Screen) {
		opts := screen.NewWindowOptions{Width: size.X, Height: size.Y}
		if w, err := s.NewWindow(&opts); err != nil {
			log.Fatal(err)
		} else {
			win = w
			defer win.Release()
			if err := eventLoop(); err != nil {
				log.Print(err)
			}
		}
	})
}

func eventLoop() error {
	for {
		switch e := win.NextEvent().(type) {
		case key.Event:
			if e.Code == key.CodeEscape {
				return nil
			}
			if e.Direction == key.DirPress &&
				e.Modifiers == key.ModMeta &&
				e.Code == key.CodeS {
				// meta-s
				save()
			}
			if e.Direction == key.DirPress || e.Direction == key.DirNone {
				mainEditor.SendKeyEvent(e)
				win.Send(paint.Event{})
			}

		case mouse.Event:
			if e.Direction == mouse.DirPress || e.Direction == mouse.DirNone {
				mainEditor.SendMouseEvent(e)
				win.Send(paint.Event{})
			}

		case mouse.ScrollEvent:
			mainEditor.SendScrollEvent(e)
			win.Send(paint.Event{})

		case paint.Event:
			if !e.External {
				if mainEditor.Dirty() {
					win.Upload(image.ZP, mainEditor, mainEditor.Bounds())
					win.Publish()
				}
			}

		case size.Event:
			winr = e.Bounds()
			mainEditor.Resize(e.Size())
			win.Send(paint.Event{})

		case lifecycle.Event:
			if e.To == lifecycle.StageDead {
				return nil
			}

		default:
			log.Printf("unhandled %T: %[1]v", e)
		}

	}
	return nil
}

func load(s string) {
	filename = s
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Printf("error opening %q for reading: %v", filename, err)
		return
	}
	buf, err := ioutil.ReadFile(filename)
	mainEditor.Load(buf)
	f.Close()
}

func save() {
	if filename == "" {
		log.Println("saving untitled file not yet supported")
		return
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("error opening %q for writing: %v", filename, err)
	}
	r := bytes.NewBuffer(mainEditor.Contents())
	if _, err := io.Copy(f, r); err != nil {
		log.Printf("error writing to %q: %v", filename, err)
	}
	f.Close()
}

func getfont() font.Face {
	if font := os.Getenv("PLAN9FONT"); font != "" {
		readFile := func(path string) ([]byte, error) {
			return ioutil.ReadFile(filepath.Join(filepath.Dir(font), path))
		}
		fontData, err := ioutil.ReadFile(font)
		if err != nil {
			log.Fatalf("error loading font: %v", err)
		}
		face, err := plan9font.ParseFont(fontData, readFile)
		if err != nil {
			log.Fatalf("error parsing font: %v", err)
		}
		return face
	}
	return basicfont.Face7x13
}
