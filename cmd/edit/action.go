package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/mobile/event/lifecycle"
	"sigint.ca/graphics/editor/address"
)

func (p *pane) findInEditor(s string) {
	if s == "" {
		return
	}
	switch s[1] {
	case ':':
		p.main.ed.JumpTo(s[1:])
	default:
		p.main.ed.FindNext(s)
	}
}

func (p *pane) executeCmd(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}

	switch cmd {
	case "Put":
		if !dir {
			p.save()
		}
	case "Undo":
		p.main.ed.SendUndo()
	case "Redo":
		p.main.ed.SendRedo()
	case "Exit":
		if p.confirmUnsaved() {
			win.Send(lifecycle.Event{To: lifecycle.StageDead})
		} else {
			return
		}
	case "Get":
		if p.confirmUnsaved() {
			p.load(p.savedPath)
		} else {
			return
		}
	default:
		switch cmd[0] {
		case '|':
			p.pipe(cmd[1:])
		default:
			run(cmd)
		}
	}

	end := p.tag.ed.LastAddress()
	p.tag.ed.SetDot(address.Selection{From: end, To: end})
}

const confirmDuration = 3 * time.Second

func (p *pane) confirmUnsaved() bool {
	if _, ok := p.tag.ed.FindNext("Put"); !ok {
		return true // already saved
	}

	confirmed := time.Since(p.confirmTime) < confirmDuration
	p.confirmTime = time.Now()
	return confirmed
}

func (p *pane) pipe(cmd string) {
	ed := p.main.ed

	in := bytes.NewBufferString(ed.GetDotContents())
	out := new(bytes.Buffer)
	args := strings.Fields(cmd)
	c := exec.Command(args[0], args[1:]...)
	c.Stdin = in
	c.Stdout = out

	c.Run()

	ed.Replace(out.String())
}

// BUG: only works if the command exits.
func run(cmd string) {
	args := strings.Fields(cmd)
	go func() {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil || len(out) == 0 {
			return
		}
		editor := exec.Command(os.Args[0], "/dev/stdin")
		editor.Stdin = bytes.NewBuffer(out)
		editor.Run()
	}()
}
