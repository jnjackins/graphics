package main

import (
	"image"
	"log"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/font"

	"sigint.ca/graphics/editor"
)

type widget struct {
	ed    *editor.Editor
	r     image.Rectangle
	buf   screen.Buffer
	tx    screen.Texture
	dirty bool
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
	for _, w := range widgets {
		if pt.In(w.r) {
			return w, true
		}
	}
	return nil, false
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

	w.dirty = true
}

func (w *widget) redraw() {
	w.ed.Draw(w.buf.RGBA(), w.buf.Bounds())
	w.tx.Upload(image.ZP, w.buf, w.buf.Bounds())
	w.dirty = false
}

func (w *widget) release() {
	w.tx.Release()
	w.buf.Release()
}
