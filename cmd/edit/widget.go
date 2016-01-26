package main

import (
	"image"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/font"
	"sigint.ca/graphics/editor"
)

type widget struct {
	ed  *editor.Editor
	r   image.Rectangle
	buf screen.Buffer
	tx  screen.Texture
}

func newWidget(s screen.Screen, size, loc image.Point, opts *editor.OptionSet, face font.Face) *widget {
	buf, err := s.NewBuffer(size)
	if err != nil {
		panic("error creating buffer: " + err.Error())
	}
	tx, err := s.NewTexture(size)
	if err != nil {
		panic("error creating texture: " + err.Error())
	}
	w := &widget{
		ed:  editor.NewEditor(size, face, opts),
		r:   image.Rectangle{loc, loc.Add(size)},
		buf: buf,
		tx:  tx,
	}
	return w
}

func (w *widget) resize(s screen.Screen, size, loc image.Point) {
	w.ed.Resize(size)
	w.r = image.Rectangle{loc, loc.Add(size)}

	w.buf.Release()
	buf, err := s.NewBuffer(size)
	if err != nil {
		panic("error resizing buffer: " + err.Error())
	}
	w.buf = buf

	w.tx.Release()
	tx, err := s.NewTexture(size)
	if err != nil {
		panic("error resizing texture: " + err.Error())
	}
	w.tx = tx
}

func sel(pt image.Point, widgets []*widget) (*widget, bool) {
	var selected *widget
	for _, w := range widgets {
		if pt.In(w.r) {
			selected = w
		}
	}
	if selected == nil {
		return nil, false
	}

	return selected, true
}