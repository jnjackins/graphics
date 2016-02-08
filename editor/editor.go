// Package editor provides a graphical, editable text area widget.
package editor // import "sigint.ca/graphics/editor"

import (
	"image"

	"sigint.ca/clip"
	"sigint.ca/graphics/editor/internal/address"
	"sigint.ca/graphics/editor/internal/hist"
	"sigint.ca/graphics/editor/internal/text"

	"golang.org/x/image/font"
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
