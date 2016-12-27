package main

import (
	"image"
	"log"

	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/font"

	"sigint.ca/graphics/editor"
)

type widget struct {
	pane  *pane
	ed    *editor.Editor
	r     image.Rectangle
	buf   screen.Buffer
	tx    screen.Texture
	dirty bool
}

func (p *pane) newWidget(size, loc image.Point, opts *editor.OptionSet, face font.Face) *widget {
	buf, err := scr.NewBuffer(size)
	if err != nil {
		log.Fatalf("error creating buffer: %v", err)
	}
	tx, err := scr.NewTexture(size)
	if err != nil {
		log.Fatalf("error creating texture: %v", err)
	}
	return &widget{
		pane:  p,
		ed:    editor.NewEditor(face, opts),
		r:     image.Rectangle{loc, loc.Add(size)},
		buf:   buf,
		tx:    tx,
		dirty: true,
	}
}

func sel(pt image.Point, widgets []*widget) (*widget, bool) {
	for _, w := range widgets {
		if pt.In(w.r) {
			return w, true
		}
	}
	return nil, false
}

func (w *widget) resize(size, pos image.Point) {
	w.r = image.Rectangle{pos, pos.Add(size)}

	w.tx.Release()
	tx, err := scr.NewTexture(size)
	if err != nil {
		log.Fatalf("error resizing texture: %v", err)
	}
	w.tx = tx

	w.buf.Release()
	buf, err := scr.NewBuffer(size)
	if err != nil {
		log.Fatalf("error resizing buffer: %v", err)
	}
	w.buf = buf

	w.dirty = true
}

func (w *widget) draw() {
	w.ed.Draw(w.buf.RGBA(), w.buf.Bounds())
	w.tx.Upload(image.ZP, w.buf, w.buf.Bounds())
	w.dirty = false
}

func (w *widget) release() {
	w.tx.Release()
	w.buf.Release()
}
