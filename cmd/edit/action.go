package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/mobile/event/lifecycle"
	"sigint.ca/graphics/editor/address"
)

func findInEditor(s string) {
	if s == "" {
		return
	}
	switch s[1] {
	case ':':
		mainWidget.ed.JumpTo(s[1:])
	default:
		mainWidget.ed.FindNext(s)
	}
}

var reallyQuit time.Time

func executeCmd(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}

	switch cmd {
	case "Put":
		save()
	case "Undo":
		mainWidget.ed.SendUndo()
		tagWidget.ed.Load(nil) // force tag regeneration
	case "Redo":
		mainWidget.ed.SendRedo()
		tagWidget.ed.Load(nil)
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
			run(cmd)
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

func run(cmd string) {
	args := strings.Fields(cmd)
	execCmd := exec.Command(args[0], args[1:]...)
	editorCmd := exec.Command(os.Args[0], "/dev/stdin")

	r, w := io.Pipe()

	execCmd.Stdout = w
	execCmd.Stderr = w
	editorCmd.Stdin = r

	go func() { execCmd.Run(); w.Close() }()
	go editorCmd.Run()
}
