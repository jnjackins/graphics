package text

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"time"
)

// Buffer represents a text buffer which, if you send it keystrokes and mouse events,
// will maintain a graphical representation of itself accessible by the Img method.
type Buffer struct {
	img    *image.RGBA
	bgCol  *image.Uniform
	selCol *image.Uniform
	cursor *image.RGBA // the cursor to draw when nothing is selected
	font   *ttf
	lines  []*line         // the text data
	dot    Selection       // the current selection
	clear  image.Rectangle // to be cleared next redraw
	dirty  bool            // indicates the client should redraw

	// mouse related state
	dClicking    bool        // the user is potentially double clicking
	dClickTimer  *time.Timer // times out when it is too late to complete a double click
	mButtons     int         // the buttons of the most recent mouse event
	mPos         image.Point // the position of the most recent mouse event
	mSweepOrigin Address     // keeps track of the origin of a sweep
}

// NewBuffer returns a new buffer of size r, using the TTF font at fontpath.
func NewBuffer(r image.Rectangle, fontpath string) (*Buffer, error) {
	f, err := os.Open(fontpath)
	if err != nil {
		return nil, err
	}
	font, err := newTTF(f)
	if err != nil {
		return nil, err
	}

	bgCol := image.NewUniform(color.RGBA{R: 0xFF, G: 0xFF, B: 0xEA, A: 0xFF})
	selCol := image.NewUniform(color.RGBA{R: 0xEE, G: 0xEE, B: 0x9E, A: 0xFF})

	// draw the default cursor
	cursor := image.NewRGBA(image.Rect(0, 0, 3, font.height))
	draw.Draw(cursor, cursor.Rect, image.Black, image.ZP, draw.Src)
	h := font.height - 3 // TODO magic 3
	draw.Draw(cursor, image.Rect(0, 3, 1, h), bgCol, image.ZP, draw.Src)
	draw.Draw(cursor, image.Rect(2, 3, 3, h), bgCol, image.ZP, draw.Src)

	b := &Buffer{
		img:    image.NewRGBA(r),
		bgCol:  bgCol,
		selCol: selCol,
		cursor: cursor,
		font:   font,
		lines:  []*line{&line{}},
		clear:  r,
	}

	return b, nil
}

// Resize resizes the Buffer. Subsequent calls to Img will return an image of size r.
func (b *Buffer) Resize(r image.Rectangle) {
	b.img = image.NewRGBA(r)
	b.clear = r
	for _, line := range b.lines {
		line.dirty = true
	}
	b.dirty = true
}

// Dirty returns true if the Buffer has changed visibly since the last call to
// Img.
func (b *Buffer) Dirty() bool {
	return b.dirty
}

// Img returns an image representing the current state of the Buffer.
func (b *Buffer) Img() *image.RGBA {
	b.redraw()
	b.dirty = false
	return b.img
}

// Select sets the current selection of the Buffer.
func (b *Buffer) Select(head, tail Address) {
	b.dot = Selection{head, tail}
}

// LoadString replaces the currently selected text with s, and returns the new Selection.
func (b *Buffer) LoadString(s string) Selection {
	b.load(s)
	return b.dot
}

// SendKey sends a keystroke to be interpreted by the Buffer.
func (b *Buffer) SendKey(r rune) {
	b.handleKey(r)
}

// SendMouseEvent sends a mouse event to be interpreted by the Buffer.
func (b *Buffer) SendMouseEvent(pos image.Point, buttons int) {
	b.handleMouseEvent(pos, buttons)
}
