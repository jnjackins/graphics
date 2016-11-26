package main

import "strings"

const tagSep = " |"

func updateTag() {
	old := string(tagWidget.ed.Buffer.Contents())

	// the part before the first " " is the filepath
	i := strings.Index(old, " ")
	if i >= 0 {
		currentPath = old[:i]
		old = old[i+1:]
	}

	// only the first line is uneditable
	var keep string
	if i := strings.Index(old, tagSep); i > 0 {
		keep = old[i+len(tagSep):]
		old = old[:i]
	}

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

	new := strings.Join(parts, " ")
	if old == new {
		return
	}

	dot := tagWidget.ed.Dot
	tagWidget.ed.Load([]byte(currentPath + " " + new + tagSep + keep))
	tagWidget.ed.Dot = dot
}
