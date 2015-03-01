package text

import (
	"image"
)

type state struct {
	dot        Selection
	scrollpt   image.Point
	lines      []*line
	prev, next *state
}

func (b *Buffer) undo() {
	if b.currentState.prev != nil {
		b.currentState = b.currentState.prev
		b.applyState()
	}
}

func (b *Buffer) redo() {
	if b.currentState.next != nil {
		b.currentState = b.currentState.next
		b.applyState()
	}
}

func linecopy(l *line) *line {
	newstr := make([]rune, len(l.s))
	copy(newstr, l.s)
	return &line{
		s:     newstr,
		dirty: true,
	}
}

// pushState adds a new history state to the list, dropping any that follow the current state.
func (b *Buffer) pushState() {
	oldstate := b.currentState

	lines := make([]*line, len(b.lines))
	change := false
	for i, line := range b.lines {
		if oldstate == nil || i >= len(oldstate.lines) || string(line.s) != string(oldstate.lines[i].s) {
			lines[i] = linecopy(line)
			change = true
		} else {
			lines[i] = line
		}
	}

	if !change {
		// no change, don't save state
		return
	}

	s := &state{
		dot:      b.dot,
		scrollpt: b.clipr.Min,
		lines:    lines,
	}
	b.currentState = s
	if oldstate != nil {
		s.prev = oldstate
		oldstate.next = s
	}
}

func (b *Buffer) applyState() {
	s := b.currentState
	b.dot = s.dot
	b.clipr = image.Rectangle{s.scrollpt, s.scrollpt.Add(b.clipr.Size())}
	b.lines = s.lines
	for _, line := range b.lines {
		line.dirty = true
	}
	b.dirty = true
}
