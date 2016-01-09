package editor

import (
	"bytes"
	"image"
	"strings"
	"testing"

	"golang.org/x/image/font/basicfont"
	"golang.org/x/mobile/event/key"
)

func TestHistoryLoad(t *testing.T) {
	face := basicfont.Face7x13
	buf := NewEditor(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	caseNum := 1
	expect := func(expected string) {
		actual := buf.Contents()
		if !bytes.Equal([]byte(expected), actual) {
			t.Errorf("case %d:\nexpected:\t%q\ngot:\t\t%q", caseNum, expected, actual)
		}
		caseNum++
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

	// start with one line
	s1 := "The quick brown fox jumps over the lazy dog."
	buf.Load([]byte(s1 + "\n"))

	// move the cursor to the end of the loaded text
	buf.selAll()
	buf.dot.From = buf.dot.To

	// simulate typing of 2 more lines
	s2 := "速い茶色のキツネは、のろまなイヌに飛びかかりました。"
	for _, r := range s2 {
		buf.SendKeyEvent(key.Event{Rune: r})
	}
	buf.SendKeyEvent(returnEvent)

	s3 := "Bonjour tout le monde!"
	for _, r := range s3 {
		buf.SendKeyEvent(key.Event{Rune: r})
	}

	expect(strings.Join([]string{s1, s2, s3}, "\n"))

	buf.SendKeyEvent(undoEvent)
	expect(s1 + "\n" + s2 + "\n")

	buf.SendKeyEvent(undoEvent)
	expect(s1 + "\n")

	buf.SendKeyEvent(undoEvent)
	expect(s1 + "\n")

	buf.SendKeyEvent(redoEvent)
	expect(s1 + "\n" + s2 + "\n")

	buf.SendKeyEvent(redoEvent)
	expect(strings.Join([]string{s1, s2, s3}, "\n"))

	buf.SendKeyEvent(redoEvent)
	expect(strings.Join([]string{s1, s2, s3}, "\n"))

	buf.SendKeyEvent(undoEvent)
	expect(s1 + "\n" + s2 + "\n")
}

func TestHistoryNoLoad(t *testing.T) {
	face := basicfont.Face7x13
	buf := NewEditor(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	caseNum := 1
	expect := func(expected string) {
		actual := buf.Contents()
		if !bytes.Equal([]byte(expected), actual) {
			t.Errorf("case %d:\nexpected:\t%q\ngot:\t\t%q", caseNum, expected, actual)
		}
		caseNum++
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

	// start with one line
	s1 := "The quick brown fox jumps over the lazy dog."
	for _, r := range s1 {
		buf.SendKeyEvent(key.Event{Rune: r})
	}
	buf.SendKeyEvent(returnEvent)

	// simulate typing of 2 more lines
	s2 := "速い茶色のキツネは、のろまなイヌに飛びかかりました。"
	for _, r := range s2 {
		buf.SendKeyEvent(key.Event{Rune: r})
	}
	buf.SendKeyEvent(returnEvent)

	s3 := "Bonjour tout le monde!"
	for _, r := range s3 {
		buf.SendKeyEvent(key.Event{Rune: r})
	}

	expect(strings.Join([]string{s1, s2, s3}, "\n"))

	buf.SendKeyEvent(undoEvent)
	expect(s1 + "\n" + s2 + "\n")

	buf.SendKeyEvent(undoEvent)
	expect(s1 + "\n")

	buf.SendKeyEvent(undoEvent)
	expect("")

	buf.SendKeyEvent(undoEvent)
	expect("")

	buf.SendKeyEvent(redoEvent)
	expect(s1 + "\n")

	buf.SendKeyEvent(redoEvent)
	expect(s1 + "\n" + s2 + "\n")

	buf.SendKeyEvent(redoEvent)
	expect(strings.Join([]string{s1, s2, s3}, "\n"))

	buf.SendKeyEvent(redoEvent)
	expect(strings.Join([]string{s1, s2, s3}, "\n"))

	buf.SendKeyEvent(undoEvent)
	expect(s1 + "\n" + s2 + "\n")
}
