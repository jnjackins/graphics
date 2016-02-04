// Package editor provides a graphical, editable text area widget.
package editor // import "sigint.ca/graphics/editor"

import (
	"image"

	"sigint.ca/clip"
	"sigint.ca/graphics/editor/internal/address"
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
	B2Action func(string) // define an action for the middle mouse button
	B3Action func(string) // define an action for the right mouse button

	opts *OptionSet

	// textual state
	buf *text.Buffer
	dot address.Selection // the current selection

	// drawing data
	r        image.Rectangle
	font     fontface
	scrollPt image.Point
	dirty    bool

	m mouseState

	// history
	history     *hist.History        // represents the Editor's history
	savePoint   *hist.Transformation // records the last time the Editor was saved, for use by Saved and SetSaved
	uncommitted *hist.Transformation // recent input which hasn't yet been committed to history

	clipboard *clip.Clipboard // used for copy or paste events
}

// NewEditor returns a new Editor with a clipping rectangle defined by size, a font face,
// and an OptionSet opts. If opts is nil, editor.SimpleTheme will be used.
func NewEditor(face font.Face, opts *OptionSet) *Editor {
	if opts == nil {
		opts = SimpleTheme
	}
	ed := &Editor{
		buf:       text.NewBuffer(),
		font:      mkFont(face),
		dirty:     true,
		opts:      opts,
		history:   new(hist.History),
		clipboard: new(clip.Clipboard),
	}
	return ed
}

// SetFont sets the Editor's font face to face.
func (ed *Editor) SetFont(face font.Face) {
	ed.font = mkFont(face)
	ed.dirty = true
}

// SetOpts reconfigures the Editor according to opts.
func (ed *Editor) SetOpts(opts *OptionSet) {
	ed.opts = opts
	ed.dirty = true
}

// Redraw draws the editor onto dst.
func (ed *Editor) Draw(dst *image.RGBA) {
	ed.draw(dst)
	ed.dirty = false
}

// Dirty reports whether the next call to Draw will result in a different image than the previous call.
func (ed *Editor) Dirty() bool {
	return ed.dirty
}

// Load replaces the contents of the Editor's text buffer with s, and resets the Editor's history.
func (ed *Editor) Load(s []byte) {
	last := len(ed.buf.Lines) - 1
	all := address.Selection{To: address.Simple{last, ed.buf.Lines[last].RuneCount()}}
	ed.dot = ed.buf.ClearSel(all)
	ed.buf.InsertString(address.Simple{0, 0}, string(s))
	ed.history = new(hist.History)
	ed.uncommitted = nil
	ed.dirty = true
}

// Contents returns the contents of the Editor's text buffer.
func (ed *Editor) Contents() []byte {
	return ed.buf.Contents()
}

// FindNext searches for s in the Editor's text buffer, and selects the first match
// starting from the current selection, possibly wrapping around to the beginning
// of the buffer. If there are no matches, the selection is unchanged.
// TODO: go back to dedicated implementation
func (ed *Editor) FindNext(s string) {
	if sel, ok := ed.buf.Find(ed.dot.To, s); ok {
		ed.dot = sel
		ed.autoscroll()
		ed.dirty = true
	}
}

// JumpTo sets the selection to the specified address, as define in sam(1).
func (ed *Editor) JumpTo(addr string) {
	if sel, ok := ed.buf.JumpTo(ed.dot.To, addr); ok {
		ed.dot = sel
		ed.autoscroll()
		ed.dirty = true
	}
}

// SetSaved instructs the Editor that the current contents should be
// considered saved. After calling SetSaved, the client can call
// Saved to see if the Editor has unsaved content.
func (ed *Editor) SetSaved() {
	if ed.uncommitted != nil {
		ed.commitTransformation()
	}
	ed.savePoint = ed.history.Current()
}

// Saved reports whether the Editor has been modified since the last
// time SetSaved was called.
func (ed *Editor) Saved() bool {
	return ed.history.Current() == ed.savePoint &&
		(ed.uncommitted == nil || ed.uncommitted.Post.Text == "")
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
		pt.X = int(e.Dx * float32(ed.font.height))
		pt.Y = int(e.Dy * float32(ed.font.height))
	}
	oldPt := ed.scrollPt
	ed.scroll(pt)
	if ed.scrollPt != oldPt {
		ed.dirty = true
	}
}

// SendUndo attempts to apply the Editor's previous history state, if it exists.
func (ed *Editor) SendUndo() {
	ed.undo()
}

// SendRedo attempts to apply the Editor's next history state, if it exists.
func (ed *Editor) SendRedo() {
	ed.redo()
}

// CanUndo reports whether the Editor has a previous history state which can be applied.
func (ed *Editor) CanUndo() bool {
	return ed.history.CanUndo() ||
		ed.uncommitted != nil && len(ed.uncommitted.Post.Text) > 0
}

// CanRedo reports whether the Editor has a following history state which can be applied.
func (ed *Editor) CanRedo() bool {
	return ed.history.CanRedo()
}
