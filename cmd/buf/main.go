package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/jnjackins/die"
	"github.com/jnjackins/graphics/scrollbar"
	"github.com/jnjackins/graphics/text"
	"github.com/jnjackins/plan9/draw"
)

const (
	width   = 800
	height  = 600
	sbWidth = 12 // scrollbar
)

var (
	buf      *text.Buffer
	bufPos   = image.Pt(sbWidth, 0)
	sb       *scrollbar.Scrollbar
	disp     *draw.Display
	bufImg   *draw.Image // the full (unclipped) image
	screen   *draw.Image // the final clipped image
	oldClipr image.Rectangle
)

var cprof = flag.String("cprof", "", "save cpu profile to `path`")

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

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [file]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) > 1 {
		usage()
		os.Exit(1)
	}

	// possibly start cpu profiling
	if *cprof != "" {
		profileWriter, err := os.Create(*cprof)
		die.On(err, "buf: error creating file for cpu profile")
		defer profileWriter.Close()
		pprof.StartCPUProfile(profileWriter)
		defer pprof.StopCPUProfile()
	}

	// load font
	fontPath := os.Getenv("font")
	if fontPath == "" {
		fontPath = os.Getenv("GOPATH") + "/src/github.com/jnjackins/graphics/cmd/buf/proggyfont.ttf"
	}

	// possibly load input file
	var path string // used later to save the file
	var inputFile *os.File
	if len(flag.Args()) == 1 {
		path = flag.Arg(0)
		_, err := os.Stat(path)
		if err != nil {
			// if there's no file, no worries. otherwise, bail.
			if !os.IsNotExist(err) {
				die.On(err, "buf: error opening input file")
			}
		} else {
			// no issues, open file for reading
			inputFile, err = os.Open(path)
			die.On(err, "buf: error opening input file")
		}
	}

	var err error
	if inputFile == nil {
		// even though inputFile is nil, we must use the value nil. Otherwise, NewBuffer will
		// report inputFile != nil because it receives a non-nil interface.
		buf, err = text.NewBuffer(width-sbWidth, height, fontPath, nil, text.AcmeTheme)
		die.On(err, "buf: error creating new text buffer")
	} else {
		buf, err = text.NewBuffer(width-sbWidth, height, fontPath, inputFile, text.AcmeTheme)
		inputFile.Close()
	}

	// scrollbar
	bg := color.RGBA{R: 0x99, G: 0x99, B: 0x4C, A: 0xFF}
	fg := color.RGBA{R: 0xFF, G: 0xFF, B: 0xEA, A: 0xFF}
	sb = scrollbar.New(sbWidth, height, bg, fg)

	// setup display device
	winName := path
	if winName == "" {
		winName = "<no file>"
	}
	disp, err = draw.Init(winName, width, height)
	die.On(err, "buf: error initializing display device")
	defer disp.Close()
	screen = disp.ScreenImage

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
			buf.SendMouseEvent(me.Point.Sub(bufPos), me.Buttons)
			for len(mouse.C) > 0 {
				me = <-mouse.C
				buf.SendMouseEvent(me.Point.Sub(bufPos), me.Buttons)
			}
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

		// Rest for a moment. This allows mouse.C to fill up with mouse events,
		// in case we are receiving rapid fire mouse events (in which case we
		// don't need to redraw after each event)
		time.Sleep(1 * time.Millisecond)
	}
}

func redraw() {
	img, clipr, dirty := buf.Img()
	if dirty != image.ZR {
		// it is possible that img.Bounds() has changed, in which case we need to
		// resize bufImg as well.
		if bufImg == nil || bufImg.Bounds() != img.Bounds() {
			var err error
			bufImg, err = disp.AllocImage(img.Bounds(), draw.ABGR32, false, draw.White)
			die.On(err, "buf: error allocating image")
		}
		_, err := bufImg.Load(dirty, img.SubImage(dirty).(*image.RGBA).Pix)
		die.On(err, "buf: error loading buffer to plan9 image")
	}
	if dirty != image.ZR || clipr != oldClipr {
		// draw buffer image to screen
		screen.Draw(bufImg.Bounds().Add(bufPos), bufImg, nil, clipr.Min)

		// load scrollbar image directly onto screen
		r := img.Bounds()
		r.Max.Y -= clipr.Dy()
		sb := sb.Img(clipr, r)
		_, err := screen.Load(sb.Bounds(), sb.Pix)
		die.On(err, "buf: error loading scrollbar image to screen")

		disp.Flush()
	}
	oldClipr = clipr
}

func resize() {
	err := disp.Attach(draw.Refmesg)
	die.On(err, "buf: error reattaching display after resize")
	r := screen.Bounds()
	buf.Resize(r.Dx()-sbWidth, r.Dy())
	sb.Resize(sbWidth, r.Dy())
}
