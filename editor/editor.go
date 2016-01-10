package editor // import "sigint.ca/graphics/editor"

import (
	"image"
	"time"

	"sigint.ca/graphics/editor/internal/hist"
	"sigint.ca/graphics/editor/internal/text"

	"golang.org/x/image/font"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/mouse"
)

// Editor represents a text buffer which, if you send it keystrokes and mouse events,
// will maintain a graphical representation of itself accessible by the Img method.
type Editor struct {
	// images and drawing data
	img   *image.RGBA
	clipr image.Rectangle // the part of the image in view

	// configurable
	bgcol      *image.Uniform
	selcol     *image.Uniform
	margin     image.Point
	lineHeight int
	cursor     image.Image // the cursor to draw when nothing is selected
	font       fontface

	// textual state
	buf *text.Buffer
	dot text.Selection         // the current selection
	adv map[*text.Line][]int16 // pixel advances of the runes in a line

	// history
	history     *hist.History        // represents the Editor's history
	savePoint   *hist.Transformation // records the last time the buffer was saved, for use by Saved and SetSaved
	uncommitted []rune               // recent input which hasn't yet been committed to history

	// mouse related state
	lastClickTime time.Time    // used to detect a double-click
	mButton       mouse.Button // the button of the most recent mouse event
	mPos          image.Point  // the position of the most recent mouse event
	mSweepOrigin  text.Address // keeps track of the origin of a sweep

	Clipboard Clipboard // the Clipboard to be used for copy or paste events
}

// NewEditor returns a new buffer with a clipping rectangle of size r. If init
// is not nil, the buffer uses ioutil.ReadAll(init) to initialize the text
// in the buffer. The caller should do any necessary cleanup on init after NewEditor returns.
func NewEditor(size image.Point, face font.Face, height int, opt OptionSet) *Editor {
	r := image.Rectangle{Max: size}
	ed := &Editor{
		img:   image.NewRGBA(r), // grows as needed
		clipr: r,

		bgcol:      image.NewUniform(opt.BGColor),
		selcol:     image.NewUniform(opt.SelColor),
		margin:     opt.Margin,
		lineHeight: height,
		cursor:     opt.Cursor(height),
		font:       fontface{face: face, height: height - 3},

		buf: text.NewBuffer(),
		adv: make(map[*text.Line][]int16),

		history: new(hist.History),
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

// Resize resizes the Editor. Subsequent calls to Img will return an image of
// at least size r, and a clipping rectangle of size r.
func (ed *Editor) Resize(size image.Point) {
	r := image.Rectangle{Max: size}
	ed.img = image.NewRGBA(r)
	ed.clipr = r
}

// Img returns an image representing the current state of the Editor, a rectangle
// representing the portion of the image in view based on the current scrolling position,
// and a rectangle representing the portion of the image that has changed and needs
// to be redrawn onto the display by the caller.
func (ed *Editor) RGBA() (img *image.RGBA) {
	ed.redraw()
	return ed.img
}

// Contents returns the contents of the buffer.
func (ed *Editor) Contents() []byte {
	return ed.buf.Contents()
}

// Load replaces the contents of the buffer with s, and
// resets the Editor's history.
func (ed *Editor) Load(s []byte) {
	ed.buf.ClearSel(ed.dot)
	ed.buf.InsertString(text.Address{0, 0}, string(s))
	ed.history = new(hist.History)
	ed.dot = text.Selection{}
}

// SetSaved instructs the buffer that the current contents should be
// considered saved. After calling SetSaved, the client can call
// Saved to see if the Editor has unsaved content.
func (ed *Editor) SetSaved() {
	// TODO: ensure ed.uncommitted is empty?
	if len(ed.uncommitted) > 0 {
		panic("TODO")
	}
	ed.savePoint = ed.history.Current()
}

// Saved reports whether the Editor has been modified since the last
// time SetSaved was called.
func (ed *Editor) Saved() bool {
	return ed.history.Current() == ed.savePoint && len(ed.uncommitted) == 0
}

// SendKeyEvent sends a key event to be interpreted by the Editor.
func (ed *Editor) SendKeyEvent(e key.Event) {
	ed.handleKeyEvent(e)
}

// SendMouseEvent sends a mouse event to be interpreted by the Editor.
func (ed *Editor) SendMouseEvent(e mouse.Event) {
	ed.handleMouseEvent(e)
}
