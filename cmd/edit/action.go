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

func executeCmd(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}

	switch cmd {
	case "Put":
		if !dir {
			save()
		}
	case "Undo":
		mainWidget.ed.SendUndo()
	case "Redo":
		mainWidget.ed.SendRedo()
	case "Exit":
		if confirmUnsaved() {
			win.Send(lifecycle.Event{To: lifecycle.StageDead})
		} else {
			return
		}
	case "Get":
		if confirmUnsaved() {
			load(savedPath)
		} else {
			return
		}
	default:
		switch cmd[0] {
		case '|':
			pipe(cmd[1:])
		default:
			run(cmd)
		}
	}

	end := tagWidget.ed.LastAddress()
	tagWidget.ed.SetDot(address.Selection{From: end, To: end})
}

var confirmTime time.Time

func confirmUnsaved() bool {
	if _, ok := tagWidget.ed.FindNext("Put"); !ok {
		return true // already saved
	}

	confirmed := time.Since(confirmTime) < 3*time.Second
	confirmTime = time.Now()
	return confirmed
}

func pipe(cmd string) {
	ed := mainWidget.ed

	in := bytes.NewBufferString(ed.GetDotContents())
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

	if err := execCmd.Start(); err != nil {
		return
	}
	go func() { execCmd.Wait(); w.Close() }()
	go editorCmd.Run()
}
