package main

import (
	"strings"

	"sigint.ca/graphics/editor/address"
)

const tagSep = " |"

func updateTag() {
	old := string(tagWidget.ed.Buffer.Contents())

	// the part before the first " " is the filepath
	i := strings.Index(old, " ")
	if i >= 0 {
		currentPath = old[:i]
		old = old[i+1:]
	}

	// everything after tagSep is editable and must be kept
	var keep string
	if i := strings.Index(old, tagSep); i > 0 {
		keep = old[i+len(tagSep):]
		old = old[:i]
	}

	new := tagCmds()
	if old == new {
		return
	}

	// tag contents changed

	// save the selection before loading wiping it out
	dot := tagWidget.ed.Dot

	// load the new text
	fixed := currentPath + " " + new + tagSep
	tagWidget.ed.Load([]byte(fixed + keep))

	// and fix the selection
	oldSepAddr := address.Simple{Row: 0, Col: i}
	if oldSepAddr.LessThan(dot.From) {
		dot.From.Col += len(fixed) - i
		if dot.To.Row == dot.From.Row {
			dot.To.Col += len(fixed) - i
		}
	}
	tagWidget.ed.Dot = dot
}

func tagCmds() string {
	var parts []string

	if mainWidget.ed.CanUndo() {
		parts = append(parts, "Undo")
	}
	if mainWidget.ed.CanRedo() {
		parts = append(parts, "Redo")
	}

	if currentPath != "" && (!mainWidget.ed.Saved() || currentPath != savedPath) {
		parts = append(parts, "Put")
	}

	parts = append(parts, "Exit")

	return strings.Join(parts, " ")
}
