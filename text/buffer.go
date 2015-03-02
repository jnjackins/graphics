package text

import (
	"image"
	"io"
	"io/ioutil"
	"os"
	"time"
)

// Buffer represents a text buffer which, if you send it keystrokes and mouse events,
// will maintain a graphical representation of itself accessible by the Img method.
type Buffer struct {
	// images and drawing data
	img   *image.RGBA
	clipr image.Rectangle
	clear image.Rectangle // to be cleared next redraw
	dirty image.Rectangle // the portion that is dirty (the user needs to redraw)

	// configurable
	bgCol  *image.Uniform
	selCol *image.Uniform
	margin image.Point
	cursor image.Image // the cursor to draw when nothing is selected
	font   *ttf

	// state
	lines         []*line   // the text data
	dot           Selection // the current selection
	currentState  *state    // current history state
	changingState bool      // flag used to avoid considering a state change as a new state

	// mouse related state
	dClicking    bool        // the user is potentially double clicking
	dClickTimer  *time.Timer // times out when it is too late to complete a double click
	mButtons     int         // the buttons of the most recent mouse event
	mPos         image.Point // the position of the most recent mouse event
	mSweepOrigin Address     // keeps track of the origin of a sweep

	// public variables
	Clipboard Clipboard // the Clipboard to be used for copy or paste events
}

// NewBuffer returns a new buffer with a clipping rectangle of size r. If initialText
// is not nil, the buffer uses ioutil.ReadAll(initialText) to initialize the text
// in the buffer. The caller should do any necessary cleanup on initialText after NewBuffer returns.
func NewBuffer(r image.Rectangle, fontPath string, initialText io.Reader, options OptionSet) (*Buffer, error) {
	f, err := os.Open(fontPath)
	if err != nil {
		return nil, err
	}
	font, err := newTTF(f)
	if err != nil {
		return nil, err
	}

	imgR := r
	imgR.Max.Y *= 2

	b := &Buffer{
		img:   image.NewRGBA(imgR),
		clipr: r,
		clear: imgR,
		dirty: imgR,

		bgCol:  image.NewUniform(options.BGColor),
		selCol: image.NewUniform(options.SelColor),
		margin: options.Margin,
		cursor: options.Cursor(font.height),
		font:   font,

		lines: []*line{&line{s: []rune{}, px: []int{options.Margin.X}}},
	}
	if initialText != nil {
		s, err := ioutil.ReadAll(initialText)
		if err != nil {
			return nil, err
		}
		b.load(string(s))
		b.dot = Selection{} // move dot to the beginning of the file
	}
	b.pushState()
	return b, nil
}

// Resize resizes the Buffer. Subsequent calls to Img will return an image of
// at least size r, and a clipping rectangle of size r.
func (b *Buffer) Resize(r image.Rectangle) {
	imgR := r
	imgR.Max.Y *= 2
	b.img = image.NewRGBA(imgR)
	b.clear = imgR
	b.clipr = r
	b.dirtyLines(0, len(b.lines))
	b.clear = b.img.Bounds()
}

// Img returns an image representing the current state of the Buffer, a rectangle
// representing the portion of the image in view based on the current scrolling position,
// and a rectangle representing the portion of the image that has changed and needs
// to be redrawn onto the display by the caller.
func (b *Buffer) Img() (img *image.RGBA, clipr, dirty image.Rectangle) {
	if b.dirty != image.ZR {
		b.redraw()
		dirty = b.dirty
		b.dirty = image.ZR
	}
	return b.img, b.clipr, dirty
}

func (b *Buffer) Contents() string {
	var s string
	for i, line := range b.lines {
		s += string(line.s)
		if i < len(b.lines)-1 {
			s += "\n"
		}
	}
	return s
}

// SendKey sends a keystroke to be interpreted by the Buffer.
func (b *Buffer) SendKey(r rune) {
	b.handleKey(r)
}

// SendMouseEvent sends a mouse event to be interpreted by the Buffer.
func (b *Buffer) SendMouseEvent(pos image.Point, buttons int) {
	b.handleMouseEvent(pos, buttons)
}
