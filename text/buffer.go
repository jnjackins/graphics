package text

import (
	"image"
	"os"
	"time"
)

// Buffer represents a text buffer which, if you send it keystrokes and mouse events,
// will maintain a graphical representation of itself accessible by the Img method.
type Buffer struct {
	// images and drawing data
	img   *image.RGBA
	dirty bool
	clipr image.Rectangle
	clear image.Rectangle // to be cleared next redraw

	// configurable
	bgCol  *image.Uniform
	selCol *image.Uniform
	margin image.Point
	cursor image.Image // the cursor to draw when nothing is selected
	font   *ttf

	// state
	lines []*line   // the text data
	dot   Selection // the current selection

	// mouse related state
	dClicking    bool        // the user is potentially double clicking
	dClickTimer  *time.Timer // times out when it is too late to complete a double click
	mButtons     int         // the buttons of the most recent mouse event
	mPos         image.Point // the position of the most recent mouse event
	mSweepOrigin Address     // keeps track of the origin of a sweep

	// public variables
	Clipboard Clipboard // the Clipboard to be used for copy or paste events
}

// NewBuffer returns a new buffer of size r, using the TTF font at fontpath.
func NewBuffer(r image.Rectangle, fontpath string, options OptionSet) (*Buffer, error) {
	f, err := os.Open(fontpath)
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
		dirty: true,
		clipr: r,
		clear: imgR,

		bgCol:  image.NewUniform(options.BGColor),
		selCol: image.NewUniform(options.SelColor),
		margin: options.Margin,
		cursor: options.Cursor(font.height),
		font:   font,

		lines: []*line{&line{s: []rune{}, px: []int{options.Margin.X}}},
	}

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
	for _, line := range b.lines {
		line.dirty = true
	}
	b.dirty = true
}

// Dirty reports whether the Buffer's image has changed since the last call to Img.
func (b *Buffer) Dirty() bool {
	return b.dirty
}

// Img returns an image representing the current state of the Buffer, and the clipping
// rectangle representing the portion currently in view.
func (b *Buffer) Img() (*image.RGBA, image.Rectangle) {
	if b.dirty {
		b.redraw()
		b.dirty = false
	}
	return b.img, b.clipr
}

// Select sets the current selection of the Buffer.
func (b *Buffer) Select(s Selection) {
	b.dot = s
}

// LoadString replaces the currently selected text with s, and returns the new Selection.
func (b *Buffer) LoadString(s string) Selection {
	b.load(s)
	return b.dot
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
