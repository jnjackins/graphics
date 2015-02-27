package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/jnjackins/die"
	"github.com/jnjackins/graphics/text"
	"github.com/jnjackins/plan9/draw"
)

const (
	width  = 800
	height = 600
)

var (
	disp   *draw.Display
	screen *draw.Image
	buf    *text.Buffer
)

var cpuprofile = flag.String("cpuprofile", "", "provide a path for cpu profile")

type snarfer struct {
	d *draw.Display
}

func (sn snarfer) Get() string {
	b, err := sn.d.ReadSnarf()
	if err != nil {
		log.Println("buf: error reading from snarf buffer: " + err.Error())
	}
	return string(b)
}

func (sn snarfer) Put(s string) {
	err := sn.d.WriteSnarf([]byte(s))
	if err != nil {
		log.Println("buf: error writing to snarf buffer: " + err.Error())
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	inputfile := false
	if len(flag.Args()) == 1 {
		inputfile = true
	} else if len(flag.Args()) > 1 {
		fmt.Fprintln(os.Stderr, "Usage: buf [file]")
		os.Exit(1)
	}

	if *cpuprofile != "" {
		profileWriter, err := os.Create(*cpuprofile)
		die.On(err, "buf: error creating file for cpu profile")
		defer profileWriter.Close()
		pprof.StartCPUProfile(profileWriter)
		defer pprof.StopCPUProfile()
	}

	gopath := os.Getenv("GOPATH")
	fontpath := gopath + "/src/github.com/jnjackins/graphics/cmd/buf/proggyfont.ttf"
	var err error
	buf, err = text.NewBuffer(image.Rect(0, 0, width, height), fontpath, text.AcmeTheme)
	die.On(err, "buf: error creating new text buffer")

	var path string
	if inputfile {
		path = flag.Arg(0)
		_, err := os.Stat(path)
		if err != nil {
			// if there's no file, no worries. otherwise, bail.
			if !os.IsNotExist(err) {
				die.On(err, "buf: error opening input file")
			}
		} else {
			// no issues, open file for reading
			f, err := os.Open(path)
			die.On(err, "buf: error opening input file")
			s, err := ioutil.ReadAll(f)
			buf.LoadString(string(s))
			buf.Select(text.Selection{})
			f.Close()
		}
	}

	disp, err = draw.Init("buf", width, height)
	die.On(err, "buf: error initializing display device")
	defer disp.Close()
	buf.Clipboard = snarfer{disp}
	screen = disp.ScreenImage
	kbd := disp.InitKeyboard()
	mouse := disp.InitMouse()
	redraw()

loop:
	for {
		select {
		case <-mouse.Resize:
			resize()
		case me := <-mouse.C:
			buf.SendMouseEvent(me.Point, me.Buttons)

			// if we are getting rapid-fire mouse events (scrolling or sweeping),
			// batch up a bunch of them before redrawing.
			for len(mouse.C) > 0 {
				me = <-mouse.C
				buf.SendMouseEvent(me.Point, me.Buttons)
			}
		case ke := <-kbd.C:
			// esc
			if ke == 27 {
				if path != "" {
					err := ioutil.WriteFile(path, []byte(buf.Contents()), 0666)
					die.On(err, "buf: error writing to file")
				}
				break loop
			}
			buf.SendKey(ke)
		}
		if buf.Dirty() {
			redraw()
		}
	}
}

func redraw() {
	img, clipr := buf.Img()
	pix := img.SubImage(clipr).(*image.RGBA).Pix
	_, err := screen.Load(screen.Bounds(), pix)
	die.On(err, "buf: error sending image bytes to display")
	disp.Flush()
}

func resize() {
	err := disp.Attach(draw.Refmesg)
	die.On(err, "buf: error reattaching display after resize")
	buf.Resize(screen.Bounds())
}
