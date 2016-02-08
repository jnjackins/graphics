package main

import (
	"image"
	"log"

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
		log.Fatalf("error creating buffer: %v", err)
	}
	tx, err := s.NewTexture(size)
	if err != nil {
		log.Fatalf("error creating texture: %v", err)
	}
	w := &widget{
		ed:  editor.NewEditor(face, opts),
		r:   image.Rectangle{loc, loc.Add(size)},
		buf: buf,
		tx:  tx,
	}
	return w
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

func (w *widget) resize(s screen.Screen, size, loc image.Point) {
	w.r = image.Rectangle{loc, loc.Add(size)}

	w.tx.Release()
	tx, err := s.NewTexture(size)
	if err != nil {
		log.Fatalf("error resizing texture: %v", err)
	}
	w.tx = tx

	w.buf.Release()
	buf, err := s.NewBuffer(size)
	if err != nil {
		log.Fatalf("error resizing buffer: %v", err)
	}
	w.buf = buf

	w.ed.SetDirty()
}

func (w *widget) redraw() {
	w.ed.Draw(w.buf.RGBA(), image.ZP)
	w.tx.Upload(image.ZP, w.buf, w.buf.Bounds())
}

func (w *widget) release() {
	w.tx.Release()
	w.buf.Release()
}
