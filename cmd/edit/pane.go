package main

import (
	"fmt"
	"image"
	"log"
	"path/filepath"
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
	dir         bool
	cwd         string

	// used for confirmation before closing unsaved pane.
	// the destructive action must be requested twice within
	// confirmDuration.
	confirmTime time.Time
}

func (p *pane) String() string {
	return fmt.Sprintf("pane(pos=%v, savedPath=%q, currentPath=%q, dir=%v, cwd=%q)",
		p.pos, p.savedPath, p.currentPath, p.dir, p.cwd)
}

// newPane creates a new pane. If data is nil, the file named by
// name will be loaded into the main editor widget, otherwise
// the contents of data will be loaded. In either case, the tag
// widget will display name at the left side.
func newPane(name string, data []byte) (*pane, error) {
	p := &pane{currentPath: name}

	p.pos = npanes
	npanes++

	// set up the main editor widget
	sz, pt := p.mainDimensions()
	p.main = p.newWidget(sz, pt, editor.AcmeYellowTheme, fontFace)

	// load text into main editor widget
	if data != nil {
		p.load(data)
		p.cwd = getAbs(filepath.Dir(name))
	} else {
		if err := p.loadFile(name); err != nil {
			return nil, err
		}
	}

	// set up the tag widget
	sz, pt = p.tagDimensions()
	p.tag = p.newWidget(sz, pt, editor.AcmeBlueTheme, fontFace)

	// populate the tag
	p.tag.ed.Load([]byte(p.currentPath + " "))
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

func (p *pane) tagDimensions() (size, pos image.Point) {
	h := winSize.Y / npanes
	y := h * p.pos
	return image.Pt(winSize.X, tagHeight), image.Pt(0, y)
}

func (p *pane) mainDimensions() (size, pos image.Point) {
	h := winSize.Y / npanes
	y := h * p.pos
	return image.Pt(winSize.X, h-tagHeight), image.Pt(0, y+tagHeight+1)
}

func (p *pane) resize() {
	p.tag.resize(p.tagDimensions())
	p.main.resize(p.mainDimensions())
}

func (p *pane) draw() {
	p.tag.draw()
	p.main.draw()
}

func addPane(name string, data []byte) {
	p, err := newPane(name, data)
	if err != nil {
		log.Print(err)
		return
	}
	panes = append(panes, p)
	for _, p := range panes {
		p.resize()
	}
	dprintf("added pane: %v", p)
}

func deletePane(i int) {
	p := panes[i]

	// remove widgets from global slice
	for i, w := range widgets {
		if w == p.tag {
			widgets = append(widgets[:i], widgets[i+1:]...)
		}
	}
	for i, w := range widgets {
		if w == p.main {
			widgets = append(widgets[:i], widgets[i+1:]...)
		}
	}
	// release widgets
	p.tag.release()
	p.main.release()

	// remove pane from global slice
	panes = append(panes[:i], panes[i+1:]...)
	for ; i < len(panes); i++ {
		panes[i].pos--
	}
	npanes--
}
