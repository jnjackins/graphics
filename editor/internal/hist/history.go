package hist

import (
	"fmt"

	"sigint.ca/graphics/editor/internal/text"
)

type History struct {
	current *Transformation
}

type Transformation struct {
	Pre, Post  Chunk
	prev, next *Transformation
}

func (t *Transformation) String() string {
	return fmt.Sprintf("%v -> %v", t.Pre, t.Post)
}

type Chunk struct {
	Text string
	Sel  text.Selection
}

func (c Chunk) String() string {
	return fmt.Sprintf("%q%v", c.Text, c.Sel)
}

func (h *History) Current() *Transformation {
	if h.current == nil {
		h.current = new(Transformation)
	}
	return h.current
}

func (h *History) Commit() {
	if h.current == nil {
		h.current = new(Transformation)
	}
	if h.current.Pre.Sel.From != h.current.Post.Sel.From {
		panic(fmt.Sprintf("internal error: mismatched Sel.From values in history transformation: %v != %v",
			h.current.Pre.Sel.From, h.current.Post.Sel.From))
	}
	h.current.next = &Transformation{prev: h.current}
	h.current = h.current.next
}

func (h *History) Undo() (Chunk, bool) {
	if h.current == nil || h.current.prev == nil {
		return Chunk{}, false
	}
	h.current = h.current.prev
	return Chunk{
		Text: h.current.Pre.Text,
		Sel:  h.current.Post.Sel,
	}, true
}

func (h *History) Redo() (Chunk, bool) {
	if h.current == nil || h.current.next == nil {
		return Chunk{}, false
	}
	tr := h.current
	h.current = h.current.next
	return Chunk{
		Text: tr.Post.Text,
		Sel:  tr.Pre.Sel,
	}, true
}

func (h *History) CanUndo() bool {
	return h.current != nil && h.current.prev != nil
}

func (h *History) CanRedo() bool {
	return h.current != nil && h.current.next != nil
}
