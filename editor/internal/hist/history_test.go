package hist

import (
	"testing"

	"sigint.ca/graphics/editor/internal/text"
)

func TestUndoRedo(t *testing.T) {
	h := new(History)
	h.Current().Pre = Chunk{Text: "foo", Sel: text.Sel(1, 0, 1, 2)}
	h.Current().Post = Chunk{Text: "foobar", Sel: text.Sel(1, 0, 1, 5)}
	h.Commit()

	ch, ok := h.Undo()
	if !ok {
		t.Errorf("got ok=%v, expected %v", ok, true)
	}
	expected := Chunk{Text: "foo", Sel: text.Sel(1, 0, 1, 5)}
	if ch.Sel != expected.Sel {
		t.Errorf("got ch.Sel=%v, expected %v", ch.Sel, expected.Sel)
	}
	if ch.Text != expected.Text {
		t.Errorf("got ch.Text=%v, expected %v", ch.Text, expected.Text)
	}

	ch, ok = h.Redo()
	if !ok {
		t.Errorf("got ok=%v, expected %v", ok, true)
	}
	expected = Chunk{Text: "foobar", Sel: text.Sel(1, 0, 1, 2)}
	if ch.Sel != expected.Sel {
		t.Errorf("got ch.Sel=%v, expected %v", ch.Sel, expected.Sel)
	}
	if ch.Text != expected.Text {
		t.Errorf("got ch.Text=%v, expected %v", ch.Text, expected.Text)
	}
}
