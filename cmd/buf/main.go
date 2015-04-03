// TODO: right click to search
// TODO: make scrollbar clickable
// TODO: mouse button chords
// TODO: drag border to resize tag
// TODO: output element
// TODO: escape toggles visibility of tag/output elements

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
	widthHint  = 800
	heightHint = 600
	tagHeight  = 40
	sbWidth    = 12
)

type element struct {
	buf      *text.Buffer
	img      *draw.Image
	pos      image.Point
	sb       *scrollbar.Scrollbar
	sbImg    *draw.Image
	oldClipr image.Rectangle
}

var (
	filePath string
	tag      *element
	primary  *element
	disp     *draw.Display
	screen   *draw.Image
)

var cprof = flag.String("cprof", "", "save cpu profile to `path`")

type snarfer struct {
	d *draw.Display
}

func (sn snarfer) Get() string {
	b, err := sn.d.ReadSnarf()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading from snarf buffer: "+err.Error())
	}
	return string(b)
}

func (sn snarfer) Put(s string) {
	err := sn.d.WriteSnarf([]byte(s))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error writing to snarf buffer: "+err.Error())
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
		die.On(err, "error creating file for cpu profile")
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
				die.On(err, "error opening input file")
			}
		} else {
			// no issues, open file for reading
			inputFile, err = os.Open(filePath)
			die.On(err, "error opening input file")
		}
	}

	// setup display device
	winName := filePath
	if winName == "" {
		winName = "<no file>"
	}
	var err error
	disp, err = draw.Init(winName, widthHint, heightHint)
	die.On(err, "error initializing display device")
	defer disp.Close()
	screen = disp.ScreenImage
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	// create tag buffer
	tag = new(element)
	r := image.Rect(sbWidth, 0, width, tagHeight)
	tag.pos = r.Min
	tag.buf, err = text.NewBuffer(r, fontPath, nil, text.AcmeBlue)
	die.On(err, "error creating tag text buffer")

	// create primary buffer
	primary = new(element)
	r.Min.Y = tagHeight + 1
	r.Max.Y = height
	primary.pos = r.Min
	if inputFile == nil {
		// even though inputFile is nil, we must use the value nil. Otherwise,
		// NewBuffer will report inputFile != nil because it receives a non-nil
		// interface.
		primary.buf, err = text.NewBuffer(r, fontPath, nil, text.AcmeYellow)
	} else {
		primary.buf, err = text.NewBuffer(r, fontPath, inputFile, text.AcmeYellow)
		inputFile.Close()
	}
	die.On(err, "error creating main text buffer")

	// scrollbar
	bg := color.RGBA{R: 0x99, G: 0x99, B: 0x4C, A: 0xFF}
	fg := color.RGBA{R: 0xFF, G: 0xFF, B: 0xEA, A: 0xFF}
	primary.sb = scrollbar.New(image.Rect(0, 0, sbWidth, height), bg, fg)

	kbd := disp.InitKeyboard()
	mouse := disp.InitMouse()
	primary.buf.Clipboard = snarfer{disp}

	redraw(tag)
	redraw(primary)

loop:
	for {
		select {
		case <-mouse.Resize:
			resize()
		case me := <-mouse.C:
			primary.buf.SendMouseEvent(me.Point, me.Buttons)
			for len(mouse.C) > 0 {
				me = <-mouse.C
				primary.buf.SendMouseEvent(me.Point, me.Buttons)
			}
		case ke := <-kbd.C:
			switch ke {
			case keys.Return:
				n := primary.buf.Dot.Head.Row
				s := primary.buf.GetLine(n)
				indentation := getIndent(s)
				primary.buf.SendKey('\n')
				primary.buf.Load(indentation)
				primary.buf.Dot = text.Selection{primary.buf.Dot.Tail, primary.buf.Dot.Tail}
			case keys.Escape:
				break loop
			case keys.Save:
				save()
			default:
				primary.buf.SendKey(ke)
			}
		case <-disp.ExitC:
			break loop
		}
		redraw(tag)
		redraw(primary)

		// Rest for a moment. This allows mouse.C to fill up with mouse events,
		// in case we are receiving rapid fire mouse events (in which case we
		// don't need to redraw after each event)
		time.Sleep(1 * time.Millisecond)
	}
}

func redraw(e *element) {
	img, clipr, dirty := e.buf.Img()
	if dirty != image.ZR {
		// it is possible that img.Bounds() has changed, in which case we need to
		// resize buf.img as well.
		if e.img == nil || e.img.Bounds() != img.Bounds() {
			fmt.Println("allocating new plan9 image")
			var err error
			e.img, err = disp.AllocImage(img.Bounds(), draw.ABGR32, false, draw.White)
			die.On(err, "error allocating image")
		}
		_, err := e.img.Load(dirty, img.SubImage(dirty).(*image.RGBA).Pix)
		die.On(err, "error loading buffer image to plan9 image")

		if e.buf.Saved() {
			err = disp.SetLabel(filePath)
		} else {
			err = disp.SetLabel(filePath + " (unsaved)")
		}
		die.On(err, "error setting window label")
	}
	if dirty != image.ZR || clipr != e.oldClipr {
		fmt.Printf("drawing to screen. dirty: %v, clipr: %v\n", dirty, clipr)
		screen.Draw(e.img.Bounds().Add(e.pos), e.img, nil, clipr.Min)
		drawScrollbar(clipr, img.Bounds())
		disp.Flush()
	}
	e.oldClipr = clipr
}

func resize() {
	err := disp.Attach(draw.Refmesg)
	die.On(err, "error reattaching display after resize")
	r := screen.Bounds()
	primary.buf.Resize(image.Rect(sbWidth, 0, r.Dx(), r.Dy()))
	primary.sb.Resize(image.Rect(0, 0, sbWidth, r.Dy()))
}

func save() {
	if filePath != "" {
		err := ioutil.WriteFile(filePath, []byte(primary.buf.Contents()), 0666)
		die.On(err, "error writing to file")
	} else {
		fmt.Fprintln(os.Stderr, "error writing to file: no filename")
	}
	primary.buf.SetSaved()
	err := disp.SetLabel(filePath)
	die.On(err, "error setting window label")
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
	img := primary.sb.Img(visible, actual)
	if primary.sbImg == nil || primary.sbImg.Bounds() != img.Bounds() {
		var err error
		primary.sbImg, err = disp.AllocImage(img.Bounds(), draw.ABGR32, false, draw.White)
		die.On(err, "error allocating image")
	}
	_, err := primary.sbImg.Load(primary.sbImg.Bounds(), img.Pix)
	die.On(err, "error loading scrollbar image to plan9 image")
	screen.Draw(screen.Bounds(), primary.sbImg, nil, image.ZP)
}
