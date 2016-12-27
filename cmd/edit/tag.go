package main

import (
	"strings"

	"sigint.ca/graphics/editor/address"
)

const tagSep = " |"

func (p *pane) updateTag() {
	old := string(p.tag.ed.Contents())

	// the part before the first " " is the filepath
	i := strings.Index(old, " ")
	if i >= 0 {
		p.currentPath = old[:i]
		old = old[i+1:]
	}

	// everything after tagSep is editable and must be kept
	var keep string
	if i := strings.Index(old, tagSep); i > 0 {
		keep = old[i+len(tagSep):]
		old = old[:i]
	}

	new := p.tagCmds()
	if old == new {
		return
	}

	// tag contents changed

	// save the selection before loading wiping it out
	dot := p.tag.ed.GetDot()

	// load the new text
	var fixed string
	if p.dir {
		fixed = p.cwd
	} else {
		fixed = p.currentPath
	}
	fixed += " " + new + tagSep
	p.tag.ed.Load([]byte(fixed + keep))

	// and fix the selection
	oldSepAddr := address.Simple{Row: 0, Col: i}
	if oldSepAddr.LessThan(dot.From) {
		dot.From.Col += len(fixed) - i
		if dot.To.Row == dot.From.Row {
			dot.To.Col += len(fixed) - i
		}
	}
	p.tag.ed.SetDot(dot)
}

func (p *pane) tagCmds() string {
	var parts []string

	if !p.dir {
		if p.main.ed.CanUndo() {
			parts = append(parts, "Undo")
		}
		if p.main.ed.CanRedo() {
			parts = append(parts, "Redo")
		}

		if p.currentPath != "" && (!p.main.ed.Saved() || p.currentPath != p.savedPath) {
			parts = append(parts, "Put")
		}
	}

	parts = append(parts, "Exit")

	return strings.Join(parts, " ")
}
