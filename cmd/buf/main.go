package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
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
	disp     *draw.Display
	bufImg   *draw.Image // the full (unclipped) image
	buf      *text.Buffer
	oldClipr image.Rectangle
)

var cpuprofile = flag.String("cpuprofile", "", "provide a path for cpu profile")

type snarfer struct {
	d *draw.Display
}

func (sn snarfer) Get() string {
	b, err := sn.d.ReadSnarf()
	if err != nil {
		fmt.Fprintln(os.Stderr, "buf: error reading from snarf buffer: "+err.Error())
	}
	return string(b)
}

func (sn snarfer) Put(s string) {
	err := sn.d.WriteSnarf([]byte(s))
	if err != nil {
		fmt.Fprintln(os.Stderr, "buf: error writing to snarf buffer: "+err.Error())
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

	kbd := disp.InitKeyboard()
	mouse := disp.InitMouse()
	buf.Clipboard = snarfer{disp}
	redraw()

loop:
	for {
		select {
		case <-mouse.Resize:
			resize()
		case me := <-mouse.C:
			buf.SendMouseEvent(me.Point, me.Buttons)
		case ke := <-kbd.C:
			// save and quit on escape key
			if ke == 27 {
				if path != "" {
					err := ioutil.WriteFile(path, []byte(buf.Contents()), 0666)
					die.On(err, "buf: error writing to file")
				}
				break loop
			}
			buf.SendKey(ke)
		case <-disp.ExitC:
			break loop
		}
		redraw()
	}
}

func redraw() {
	dirty := buf.Dirty()
	img, clipr := buf.Img()
	if dirty {
		if bufImg == nil || bufImg.Bounds() != img.Bounds() {
			var err error
			bufImg, err = disp.AllocImage(img.Bounds(), draw.ABGR32, false, draw.White)
			die.On(err, "buf: error allocating image")
		}
		_, err := bufImg.Load(bufImg.Bounds(), img.Pix)
		die.On(err, "buf: error loading to image")
	}
	if dirty || clipr != oldClipr {
		disp.ScreenImage.Draw(disp.ScreenImage.Bounds(), bufImg, nil, image.ZP.Add(clipr.Min))
		disp.Flush()
	}
	oldClipr = clipr
}

func resize() {
	err := disp.Attach(draw.Refmesg)
	die.On(err, "buf: error reattaching display after resize")
	buf.Resize(disp.ScreenImage.Bounds())
}
