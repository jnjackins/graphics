package main

import (
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/mobile/event/lifecycle"
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

var reallyQuit time.Time

func executeCmd(cmd string) {
	cmd = strings.TrimSpace(cmd)
	switch cmd {
	case "Put":
		save()
	case "Undo":
		mainWidget.ed.SendUndo()
		tagWidget.ed.Load([]byte{}) // force tag regeneration
	case "Redo":
		mainWidget.ed.SendRedo()
		tagWidget.ed.Load([]byte{})
	case "Exit":
		if mainWidget.ed.Saved() || time.Since(reallyQuit) < 3*time.Second {
			win.Send(lifecycle.Event{To: lifecycle.StageDead})
		} else {
			tagWidget.ed.FindNext("Put")
			reallyQuit = time.Now()
		}
	}
}

func findInEditor(s string) {
	if s == "" {
		return
	}
	first, sz := utf8.DecodeRuneInString(s)
	switch first {
	case ':':
		mainWidget.ed.JumpTo(s[sz:])
	default:
		mainWidget.ed.FindNext(s)
	}
}
