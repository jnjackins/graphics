package text

import (
	"bytes"
	"image"
	"strings"
	"testing"

	"golang.org/x/image/font/basicfont"
	"golang.org/x/mobile/event/key"
)

func TestHistory(t *testing.T) {
	face := basicfont.Face7x13
	buf := NewBuffer(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	expect := func(expected string) {
		actual := buf.Contents()
		if !bytes.Equal([]byte(expected), actual) {
			t.Errorf("expected %q, got %q", expected, actual)
		}
	}

	returnEvent := key.Event{
		Code:      key.CodeReturnEnter,
		Direction: key.DirPress,
		Rune:      '\n',
	}
	undoEvent := key.Event{
		Code:      key.CodeZ,
		Direction: key.DirPress,
		Modifiers: key.ModMeta,
		Rune:      'z',
	}
	redoEvent := key.Event{
		Code:      key.CodeZ,
		Direction: key.DirPress,
		Modifiers: key.ModMeta | key.ModShift,
		Rune:      'z',
	}

	// simulate typing of 3 lines
	s1 := "The quick brown fox jumps over the lazy dog."
	s2 := "速い茶色のキツネは、のろまなイヌに飛びかかりました。"
	s3 := "Bonjour tout le monde!"
	for _, r := range s1 {
		buf.loadRune(r, true)
		// TODO: shouldn't need to do this
		buf.dot.head = buf.dot.tail
	}
	buf.SendKeyEvent(returnEvent)
	for _, r := range s2 {
		buf.loadRune(r, true)
		buf.dot.head = buf.dot.tail
	}
	buf.SendKeyEvent(returnEvent)
	for _, r := range s3 {
		buf.loadRune(r, true)
		buf.dot.head = buf.dot.tail
	}
	buf.SendKeyEvent(returnEvent)

	expect(strings.Join([]string{s1, s2, s3}, "\n") + "\n")

	buf.SendKeyEvent(undoEvent)
	expect(strings.Join([]string{s1, s2}, "\n") + "\n")

	buf.SendKeyEvent(undoEvent)
	expect(s1 + "\n")

	buf.SendKeyEvent(undoEvent)
	expect("")

	buf.SendKeyEvent(redoEvent)
	expect(s1 + "\n")

	buf.SendKeyEvent(redoEvent)
	expect(strings.Join([]string{s1, s2}, "\n") + "\n")

	buf.SendKeyEvent(redoEvent)
	expect(strings.Join([]string{s1, s2, s3}, "\n") + "\n")
}
