package main

import (
	"bytes"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/mobile/event/lifecycle"
	"sigint.ca/graphics/editor/address"
)

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

var reallyQuit time.Time

func executeCmd(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		return
	}

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
		_, ok := tagWidget.ed.FindNext("Put")
		if !ok || time.Since(reallyQuit) < 3*time.Second {
			win.Send(lifecycle.Event{To: lifecycle.StageDead})
		}
		reallyQuit = time.Now()

	default:
		switch cmd[0] {
		case '|':
			pipe(cmd[1:])
		default:
			return
		}
	}

	// fix tag selection
	end := tagWidget.ed.Buffer.LastAddress()
	tagWidget.ed.Dot = address.Selection{From: end, To: end}
}

func pipe(cmd string) {
	ed := mainWidget.ed

	in := bytes.NewBufferString(ed.Buffer.GetSel(ed.Dot))
	out := new(bytes.Buffer)
	args := strings.Fields(cmd)
	c := exec.Command(args[0], args[1:]...)
	c.Stdin = in
	c.Stdout = out

	c.Run()

	ed.Replace(out.String())
}
