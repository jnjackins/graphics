// Package editor provides a graphical, editable text area widget.
package editor // import "sigint.ca/graphics/editor"

import (
	"image"
	"time"

	"sigint.ca/clip"
	"sigint.ca/graphics/editor/internal/hist"
	"sigint.ca/graphics/editor/internal/text"

	"golang.org/x/image/font"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"
)

// An Editor is a graphical, editable text area widget, intended to be
// compatible with golang.org/x/exp/shiny, or any other graphical
// window package capable of drawing a widget via an image.RGBA and
// clipping rectangle. See sigint.ca/cmd/edit for an example program
// using this type.
type Editor struct {
	// textual state
	buf *text.Buffer
	dot text.Selection // the current selection

	// images and drawing data
	img   *image.RGBA
	clipr image.Rectangle // the part of the image in view

	// configurable
	bgcol      *image.Uniform
	selcol     *image.Uniform
	cursor     image.Image // the cursor to draw when nothing is selected
	font       fontface
	lineHeight int
	margin     image.Point

	// history
	history     *hist.History        // represents the Editor's history
	savePoint   *hist.Transformation // records the last time the Editor was saved, for use by Saved and SetSaved
	uncommitted *hist.Transformation // recent input which hasn't yet been committed to history

	// mouse related state
	lastClickTime time.Time    // used to detect a double-click
	mPos          image.Point  // the position of the most recent mouse event
	mSweepOrigin  text.Address // keeps track of the origin of a sweep

	clipboard *clip.Clipboard // the clipboard to be used for copy or paste events
}

// NewEditor returns a new Editor with a clipping rectangle defined by size, a font face
// defined by face and height, and an OptionSet opt.
func NewEditor(size image.Point, face font.Face, height int, opt OptionSet) *Editor {
	r := image.Rectangle{Max: size}
	ed := &Editor{
		buf: text.NewBuffer(),

		img:   image.NewRGBA(r), // grows as needed
		clipr: r,

		bgcol:      image.NewUniform(opt.BGColor),
		selcol:     image.NewUniform(opt.SelColor),
		cursor:     opt.Cursor(height),
		font:       fontface{face: face, height: height - 3},
		lineHeight: height,
		margin:     opt.Margin,

		history:   new(hist.History),
		clipboard: new(clip.Clipboard),
	}
	return ed
}

func (ed *Editor) Release() {
}

func (ed *Editor) Bounds() image.Rectangle {
	return ed.clipr
}

func (ed *Editor) Size() image.Point {
	return ed.clipr.Size()
}

// Resize resizes the Editor. Subsequent calls to RGBA will return an image of
// at least size, and a clipping rectangle of size.
func (ed *Editor) Resize(size image.Point) {
	r := image.Rectangle{Max: size}
	ed.img = image.NewRGBA(r)
	ed.clipr = r
}

// RGBA returns an image representing the current state of the Editor. The image
// may be larger than the rectangle returned by Bounds, which represents
// the portion of the image currently scrolled into view.
func (ed *Editor) RGBA() (img *image.RGBA) {
	ed.redraw()
	return ed.img
}

// Contents returns the contents of the Editor.
func (ed *Editor) Contents() []byte {
	return ed.buf.Contents()
}

// Load replaces the contents of the Editor with s, and
// resets the Editor's history.
func (ed *Editor) Load(s []byte) {
	ed.buf.ClearSel(ed.dot)
	ed.buf.InsertString(text.Address{0, 0}, string(s))
	ed.history = new(hist.History)
	ed.dot = text.Selection{}
}

// SetSaved instructs the Editor that the current contents should be
// considered saved. After calling SetSaved, the client can call
// Saved to see if the Editor has unsaved content.
func (ed *Editor) SetSaved() {
	// TODO: ensure ed.uncommitted is empty?
	if ed.uncommitted != nil {
		panic("TODO")
	}
	ed.savePoint = ed.history.Current()
}

// Saved reports whether the Editor has been modified since the last
// time SetSaved was called.
func (ed *Editor) Saved() bool {
	return ed.history.Current() == ed.savePoint && ed.uncommitted == nil
}

// SendKeyEvent sends a key event to be interpreted by the Editor.
func (ed *Editor) SendKeyEvent(e key.Event) {
	ed.handleKeyEvent(e)
}

// SendMouseEvent sends a mouse event to be interpreted by the Editor.
func (ed *Editor) SendMouseEvent(e mouse.Event) {
	ed.handleMouseEvent(e)
}
