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
// window package capable of drawing a widget via an image.RGBA.
// See sigint.ca/cmd/edit for an example program using this type.
type Editor struct {
	// textual state
	buf *text.Buffer
	dot text.Selection // the current selection

	// images and drawing data
	img      *image.RGBA
	dirty    bool
	scrollPt image.Point // the part of the image in view

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
func NewEditor(size image.Point, face font.Face, opt OptionSet) *Editor {
	fontface := mkFontface(face)
	ed := &Editor{
		buf: text.NewBuffer(),

		img:   image.NewRGBA(image.Rectangle{Max: size}),
		dirty: true,

		bgcol:      image.NewUniform(opt.BGColor),
		selcol:     image.NewUniform(opt.SelColor),
		cursor:     opt.Cursor(fontface.height),
		font:       fontface,
		lineHeight: fontface.height,
		margin:     opt.Margin,

		history:   new(hist.History),
		clipboard: new(clip.Clipboard),
	}
	return ed
}

func (ed *Editor) Release() {
}

func (ed *Editor) Bounds() image.Rectangle {
	return ed.img.Bounds()
}

func (ed *Editor) Size() image.Point {
	return ed.Bounds().Size()
}

// Resize resizes the Editor. Subsequent calls to RGBA will return an image of
// the given size.
func (ed *Editor) Resize(size image.Point) {
	r := image.Rectangle{Max: size}
	ed.img = image.NewRGBA(r)
	ed.dirty = true
}

// RGBA returns an image representing the current state of the Editor.
func (ed *Editor) RGBA() (img *image.RGBA) {
	if ed.dirty {
		ed.redraw()
		ed.dirty = false
	}
	return ed.img
}

// Dirty reports whether the next call to RGBA will result in a different
// image than the previous call.
func (ed *Editor) Dirty() bool {
	return ed.dirty
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
	ed.dirty = true
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

// SendScrollEvent sends a scroll event to be interpreted by the Editor.
func (ed *Editor) SendScrollEvent(e mouse.ScrollEvent) {
	var pt image.Point
	if e.Precise {
		pt.X = int(e.Dx)
		pt.Y = int(e.Dy)
	} else {
		pt.X = int(e.Dx * float32(ed.lineHeight))
		pt.Y = int(e.Dy * float32(ed.lineHeight))
	}
	oldPt := ed.scrollPt
	ed.scroll(pt)
	if ed.scrollPt != oldPt {
		ed.dirty = true
	}
}
