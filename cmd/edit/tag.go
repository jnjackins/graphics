package main

import "strings"

var tagline string

func updateTag() {
	var newTags []string
	if filename != "" {
		newTags = append(newTags, filename)
	}
	if mainWidget.ed.CanUndo() {
		newTags = append(newTags, "Undo")
	}
	if mainWidget.ed.CanRedo() {
		newTags = append(newTags, "Redo")
	}
	if filename != "" && !mainWidget.ed.Saved() {
		newTags = append(newTags, "Put")
	}

	newTagline := strings.Join(newTags, " ")
	if newTagline == tagline {
		return
	}
	tagline = newTagline

	tagWidget.ed.Load([]byte(tagline))
}

func editorCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	switch cmd {
	case "Put":
		save()
	case "Undo":
		mainWidget.ed.Undo()
	case "Redo":
		mainWidget.ed.Redo()
	}
}
