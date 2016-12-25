package main

import (
	"image"
	"time"

	"sigint.ca/graphics/editor"
	"sigint.ca/graphics/editor/address"
)

var npanes int

// A pane is an acme-style managed window (not an OS window)
type pane struct {
	tag, main *widget
	pos       int

	savedPath   string
	currentPath string

	// used for confirmation before closing unsaved pane.
	// the destructive action must be requested twice within
	// confirmDuration.
	confirmTime time.Time
}

func newPane(path string) (*pane, error) {
	p := &pane{currentPath: path}

	npanes++
	height := winSize.Y / npanes

	// set up the main editor widget
	sz, pt := image.Pt(winSize.X, height-tagHeight), image.Pt(0, tagHeight+1)
	p.main = newWidget(sz, pt, editor.AcmeYellowTheme, fontFace)

	// load file into main editor widget
	if err := p.load(path); err != nil {
		return nil, err
	}

	// set up the tag widget
	sz, pt = image.Pt(winSize.X, tagHeight), image.ZP
	p.tag = newWidget(sz, pt, editor.AcmeBlueTheme, fontFace)

	// populate the tag
	p.tag.ed.Load([]byte(path + " "))
	p.updateTag()
	end := p.tag.ed.LastAddress()
	p.tag.ed.SetDot(address.Selection{From: end, To: end})

	// set up B2 and B3 actions
	p.tag.ed.B2Action = p.executeCmd
	p.main.ed.B2Action = p.executeCmd
	p.tag.ed.B3Action = p.findInEditor
	p.main.ed.B3Action = p.findInEditor

	widgets = append(widgets, p.tag)
	widgets = append(widgets, p.main)

	return p, nil
}

func (p *pane) release() {
	p.tag.release()
	p.main.release()
}

func (p *pane) draw() {
	p.tag.draw()
	p.main.draw()
}
