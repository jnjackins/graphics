package main

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
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
	args := strings.Fields(cmd)
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "Put":
		if !p.dir {
			p.save()
		}
	case "Undo":
		p.main.ed.SendUndo()
	case "Redo":
		p.main.ed.SendRedo()
	case "New":
		paths := []string{""}
		if len(args) > 1 {
			paths = args[1:]
		}
		for _, path := range paths {
			addPane(path, nil)
		}
	case "Exit":
		if p.confirmUnsaved() {
			dprintf("deleting pane %d", p.pos)
			deletePane(p.pos)
			if len(panes) == 0 {
				win.Send(lifecycle.Event{To: lifecycle.StageDead})
			}
			for _, p := range panes {
				p.resize()
			}
		} else {
			return
		}
	case "Get":
		if p.confirmUnsaved() {
			p.loadFile(p.savedPath)
		} else {
			return
		}

	default:
		switch args[0][0] {
		case '|':
			p.pipe(cmd[1:])
		default:
			p.run(cmd)
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
	if len(args) == 0 {
		return
	}
	c := exec.Command(args[0], args[1:]...)
	c.Stdin = in
	c.Stdout = out

	c.Run()

	ed.Replace(out.String())
}

func (p *pane) run(cmd string) {
	args := strings.Fields(cmd)
	go func() {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil || len(out) == 0 {
			return
		}
		addPane(p.cwd+"+Errors", out)
		win.Send(paint.Event{})
	}()
}
