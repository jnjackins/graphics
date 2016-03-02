package main

import (
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/mobile/event/lifecycle"
)

func updateTag() {
	old := string(tagWidget.ed.Contents())

	// the part before the first " " is the filepath
	i := strings.Index(old, " ")
	if i >= 0 {
		pathCurrent = old[:i]
		old = old[i+1:]
	}

	// only the first line is uneditable
	var keep string
	if i := strings.Index(old, "\n"); i > 0 {
		keep = old[i+1:]
		old = old[:i]
	}

	var parts []string

	if mainWidget.ed.CanUndo() {
		parts = append(parts, "Undo")
	}
	if mainWidget.ed.CanRedo() {
		parts = append(parts, "Redo")
	}

	if pathCurrent != "" && (!mainWidget.ed.Saved() || pathCurrent != pathSaved) {
		parts = append(parts, "Put")
	}

	parts = append(parts, "Exit")

	new := strings.Join(parts, " ")
	if old == new {
		return
	}

	tagWidget.ed.Load([]byte(pathCurrent + " " + new + "\n" + keep))
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
		ok := tagWidget.ed.FindNext("Put")
		if !ok || time.Since(reallyQuit) < 3*time.Second {
			win.Send(lifecycle.Event{To: lifecycle.StageDead})
		}
		reallyQuit = time.Now()
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
