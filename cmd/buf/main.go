package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/pprof"

	"code.google.com/p/jnj.plan9/draw"
	"code.google.com/p/jnj/die"
	"github.com/jnjackins/graphics/text"
)

var (
	disp   *draw.Display
	screen *draw.Image
	buf    *text.Buffer
)

var cpuprofile = flag.String("cpuprofile", "", "provide a path for cpu profile")

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
		die.On(err, "error creating file for cpu profile")
		defer profileWriter.Close()
		pprof.StartCPUProfile(profileWriter)
		defer pprof.StopCPUProfile()
	}

	gopath := os.Getenv("GOPATH")
	fontpath := gopath + "/src/github.com/jnjackins/graphics/cmd/buf/proggyfont.ttf"
	var err error
	buf, err = text.NewBuffer(image.Rect(0, 0, 800, 600), fontpath, text.AcmeTheme)
	die.On(err)

	var path string
	if inputfile {
		path = flag.Arg(0)
		_, err := os.Stat(path)
		if err != nil {
			// if there's no file, no worries. otherwise, bail.
			if !os.IsNotExist(err) {
				die.On(err)
			}
		} else {
			// no issues, open file for reading
			f, err := os.Open(path)
			die.On(err)
			s, err := ioutil.ReadAll(f)
			buf.LoadString(string(s))
			buf.Select(text.Selection{})
			f.Close()
		}
	}

	disp, err = draw.Init(nil, "buf", "800x622")
	die.On(err)
	defer disp.Close()
	screen = disp.ScreenImage
	mouse := disp.InitMouse()
	kbd := disp.InitKeyboard()
	redraw()

loop:
	for {
		select {
		case <-mouse.Resize:
			resize()
		case me := <-mouse.C:
			buf.SendMouseEvent(me.Point, me.Buttons)
		case ke := <-kbd.C:
			// esc
			if ke == 27 {
				if path != "" {
					ioutil.WriteFile(path, []byte(buf.Contents()), 0666)
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
	_, err := screen.Load(screen.Bounds(), img.SubImage(clipr).(*image.RGBA).Pix)
	die.On(err)
	disp.Flush()
}

func resize() {
	err := disp.Attach(draw.Refmesg)
	die.On(err, "error reattaching display after resize")
	buf.Resize(screen.Bounds())
}
