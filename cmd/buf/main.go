// TODO: right click to search
// TODO: make scrollbar clickable
// TODO: mouse button chords

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
	"unicode"

	"sigint.ca/die"
	"sigint.ca/graphics/keys"
	"sigint.ca/graphics/scrollbar"
	"sigint.ca/graphics/text"
	"sigint.ca/plan9/draw"
)

const (
	width   = 800
	height  = 600
	sbWidth = 12 // scrollbar
)

var (
	filePath string
	buf      *text.Buffer
	bufImg   *draw.Image
	bufPos   = image.Pt(sbWidth, 0)
	sb       *scrollbar.Scrollbar
	sbImg    *draw.Image
	disp     *draw.Display
	screen   *draw.Image
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
		fontPath = os.Getenv("GOPATH") + "/src/sigint.ca/graphics/cmd/buf/proggyfont.ttf"
	}

	// possibly load input file
	var inputFile *os.File
	if len(flag.Args()) == 1 {
		filePath = flag.Arg(0)
		_, err := os.Stat(filePath)
		if err != nil {
			// if there's no file, no worries. otherwise, bail.
			if !os.IsNotExist(err) {
				die.On(err, "buf: error opening input file")
			}
		} else {
			// no issues, open file for reading
			inputFile, err = os.Open(filePath)
			die.On(err, "buf: error opening input file")
		}
	}

	var err error
	if inputFile == nil {
		// even though inputFile is nil, we must use the value nil. Otherwise, NewBuffer will
		// report inputFile != nil because it receives a non-nil interface.
		buf, err = text.NewBuffer(width-sbWidth, height, fontPath, nil, text.AcmeTheme)
	} else {
		buf, err = text.NewBuffer(width-sbWidth, height, fontPath, inputFile, text.AcmeTheme)
		inputFile.Close()
	}
	die.On(err, "buf: error creating new text buffer")

	// scrollbar
	bg := color.RGBA{R: 0x99, G: 0x99, B: 0x4C, A: 0xFF}
	fg := color.RGBA{R: 0xFF, G: 0xFF, B: 0xEA, A: 0xFF}
	sb = scrollbar.New(sbWidth, height, bg, fg)

	// setup display device
	winName := filePath
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
			switch ke {
			case keys.Return:
				n := buf.Dot.Head.Row
				s := buf.GetLine(n)
				indentation := getIndent(s)
				buf.SendKey('\n')
				buf.Load(indentation)
				buf.Dot = text.Selection{buf.Dot.Tail, buf.Dot.Tail}
			case keys.Escape:
				break loop
			case keys.Save:
				save()
			default:
				buf.SendKey(ke)
			}
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
		die.On(err, "buf: error loading buffer image to plan9 image")

		if buf.Saved() {
			err = disp.SetLabel(filePath)
		} else {
			err = disp.SetLabel(filePath + " (unsaved)")
		}
		die.On(err, "buf: error setting window label")
	}
	if dirty != image.ZR || clipr != oldClipr {
		screen.Draw(bufImg.Bounds().Add(bufPos), bufImg, nil, clipr.Min)
		drawScrollbar(clipr, img.Bounds())
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

func save() {
	if filePath != "" {
		err := ioutil.WriteFile(filePath, []byte(buf.Contents()), 0666)
		die.On(err, "buf: error writing to file")
	} else {
		fmt.Fprintln(os.Stderr, "buf: error writing to file: no filename")
	}
	buf.SetSaved()
	err := disp.SetLabel(filePath)
	die.On(err, "buf: error setting window label")
}

func getIndent(line string) string {
	if len(line) == 0 {
		return ""
	}
	var indent string
	for _, r := range line {
		if !unicode.IsSpace(r) {
			break
		}
		indent += string(r)
	}
	return indent
}

func drawScrollbar(visible, actual image.Rectangle) {
	actual.Max.Y -= visible.Dy()
	img := sb.Img(visible, actual)
	if sbImg == nil || sbImg.Bounds() != img.Bounds() {
		var err error
		sbImg, err = disp.AllocImage(img.Bounds(), draw.ABGR32, false, draw.White)
		die.On(err, "buf: error allocating image")
	}
	_, err := sbImg.Load(sbImg.Bounds(), img.Pix)
	die.On(err, "buf: error loading scrollbar image to plan9 image")
	screen.Draw(screen.Bounds(), sbImg, nil, image.ZP)
}
