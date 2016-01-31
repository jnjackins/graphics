package main

import (
	"os"
	"strings"
)

func updateTag() {
	old := string(tagWidget.ed.Contents())
	var user string
	if i := strings.Index(old, "|"); i > 0 {
		user = old[i+1:]
		old = old[:i+1]
	}

	var newParts []string
	if filename != "" {
		newParts = append(newParts, filename)
	}
	if mainWidget.ed.CanUndo() {
		newParts = append(newParts, "Undo")
	}
	if mainWidget.ed.CanRedo() {
		newParts = append(newParts, "Redo")
	}
	if filename != "" && !mainWidget.ed.Saved() {
		newParts = append(newParts, "Put")
	}
	newParts = append(newParts, "Exit")
	newParts = append(newParts, "|")

	new := strings.Join(newParts, " ")
	if old == new {
		return
	}

	if user == "" {
		user = " "
	}

	tagWidget.ed.Load([]byte(new + user))
}

func doEditorCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	switch cmd {
	case "Put":
		save()
	case "Undo":
		mainWidget.ed.SendUndo()
	case "Redo":
		mainWidget.ed.SendRedo()
	case "Exit":
		if mainWidget.ed.Saved() {
			os.Exit(0)
		} else {
			tagWidget.ed.Search("Put")
		}
	}
}
