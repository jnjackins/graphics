package text // import "sigint.ca/graphics/text"

import (
	"bytes"
	"image"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"
)

// Buffer represents a text buffer which, if you send it keystrokes and mouse events,
// will maintain a graphical representation of itself accessible by the Img method.
type Buffer struct {
	// images and drawing data
	img   *image.RGBA
	clipr image.Rectangle // the part of the image in view
	clear image.Rectangle // to be cleared next redraw
	dirty image.Rectangle // the portion that is dirty (the user needs to redraw)

	// configurable
	bgcol      *image.Uniform
	selcol     *image.Uniform
	margin     image.Point
	lineHeight int
	cursor     image.Image // the cursor to draw when nothing is selected
	font       fontface

	// internal state
	dot   selection // the current selection
	lines []*line   // the text data

	// history
	lastAction    *action // the most recently performed action
	savedAction   *action // used by Changed to report whether the buffer has changed
	currentAction *action // the action currently in progress

	// mouse related state
	lastClickTime time.Time    // used to detect a double-click
	mButton       mouse.Button // the button of the most recent mouse event
	mPos          image.Point  // the position of the most recent mouse event
	mSweepOrigin  address      // keeps track of the origin of a sweep

	Clipboard Clipboard // the Clipboard to be used for copy or paste events
}

// NewBuffer returns a new buffer with a clipping rectangle of size r. If init
// is not nil, the buffer uses ioutil.ReadAll(init) to initialize the text
// in the buffer. The caller should do any necessary cleanup on init after NewBuffer returns.
func NewBuffer(size image.Point, face font.Face, height int, opt OptionSet) *Buffer {
	r := image.Rectangle{Max: size}
	b := &Buffer{
		img:   image.NewRGBA(r), // grows as needed
		clipr: r,
		clear: r,
		dirty: r,

		bgcol:      image.NewUniform(opt.BGColor),
		selcol:     image.NewUniform(opt.SelColor),
		margin:     opt.Margin,
		lineHeight: height,
		cursor:     opt.Cursor(height),
		font:       fontface{face: face, height: height - 3},

		lines: []*line{&line{s: []rune{}, px: []int{opt.Margin.X}}},

		lastAction:    new(action),
		currentAction: new(action),
	}
	return b
}

func (b *Buffer) Release() {
}

func (b *Buffer) Bounds() image.Rectangle {
	return b.clipr
}

func (b *Buffer) Size() image.Point {
	return b.clipr.Size()
}

// Resize resizes the Buffer. Subsequent calls to Img will return an image of
// at least size r, and a clipping rectangle of size r.
func (b *Buffer) Resize(size image.Point) {
	r := image.Rectangle{Max: size}
	b.img = image.NewRGBA(r)
	b.clipr = r
	b.clear = r
	b.dirtyLines(0, len(b.lines))
}

// Img returns an image representing the current state of the Buffer, a rectangle
// representing the portion of the image in view based on the current scrolling position,
// and a rectangle representing the portion of the image that has changed and needs
// to be redrawn onto the display by the caller.
func (b *Buffer) RGBA() (img *image.RGBA) {
	if b.dirty != image.ZR {
		b.redraw()
		b.dirty = image.ZR
	}
	return b.img
}

// Contents returns the contents of the buffer as a string.
func (b *Buffer) Contents() []byte {
	var buf bytes.Buffer
	for i, line := range b.lines {
		buf.WriteString(string(line.s))
		if i < len(b.lines)-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.Bytes()
}

// GetLine returns a string containing the text of the nth line, where
// the first line of the buffer is line 0.
func (b *Buffer) GetLine(n int) string {
	return string(b.lines[n].s)
}

// Load replaces the current selection with s.
// TODO: clients no longer control selection
func (b *Buffer) Load(s []byte) {
	b.loadBytes(s, true)
	b.sel(address{}, address{})
}

// SetSaved instructs the buffer that the current contents should be
// considered "saved". After calling SetSaved, the client can call
// Saved to see if the Buffer has unsaved content.
func (b *Buffer) SetSaved() {
	b.savedAction = b.lastAction
}

// Saved reports whether the Buffer has been modified since the last
// time SetSaved was called.
func (b *Buffer) Saved() bool {
	committed := b.commitAction()
	if committed || b.lastAction != b.savedAction {
		return false
	}
	return true
}

// SendKey sends a key event to be interpreted by the Buffer.
func (b *Buffer) SendKeyEvent(e key.Event) {
	b.handleKeyEvent(e)
}

// SendMouseEvent sends a mouse event to be interpreted by the Buffer.
func (b *Buffer) SendMouseEvent(e mouse.Event) {
	b.handleMouseEvent(e)
}
