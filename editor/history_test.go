package editor

import (
	"testing"

	"sigint.ca/graphics/editor/address"

	"golang.org/x/image/font/basicfont"
	"golang.org/x/mobile/event/key"
)

var (
	returnEvent = key.Event{
		Code:      key.CodeReturnEnter,
		Direction: key.DirPress,
	}
	backspaceEvent = key.Event{
		Code:      key.CodeDeleteBackspace,
		Direction: key.DirPress,
	}
	undoEvent = key.Event{
		Code:      key.CodeZ,
		Direction: key.DirPress,
		Modifiers: key.ModMeta,
		Rune:      'z',
	}
	redoEvent = key.Event{
		Code:      key.CodeZ,
		Direction: key.DirPress,
		Modifiers: key.ModMeta | key.ModShift,
		Rune:      'z',
	}
)

func TestHistory(t *testing.T) {
	face := basicfont.Face7x13
	ed := NewEditor(face, AcmeYellowTheme)

	// start with one line
	s1 := "The quick brown fox jumps over the lazy dog."
	ed.Load([]byte(s1 + "\n"))

	// move the cursor to the end of the loaded text
	ed.SetDot(address.Selection{
		From: address.Simple{1, 0},
		To:   address.Simple{1, 0},
	})

	// simulate typing of 2 more lines
	s2 := "速い茶色のキツネは、のろまなイヌに飛びかかりました。"
	for _, r := range s2 {
		ed.SendKeyEvent(key.Event{Rune: r})
	}
	ed.SendKeyEvent(returnEvent)

	s3 := "Bonjour tout le monde!"
	for _, r := range s3 {
		ed.SendKeyEvent(key.Event{Rune: r})
	}

	// throw in some backspaces
	ed.SendKeyEvent(backspaceEvent)
	ed.SendKeyEvent(backspaceEvent)

	cases := []struct {
		want  string
		event *key.Event
	}{
		{want: s1 + "\n" + s2 + "\nBonjour tout le mond", event: nil},
		{want: s1 + "\n" + s2 + "\nBonjour tout le monde", event: &undoEvent},
		{want: s1 + "\n" + s2 + "\n", event: &undoEvent},
		{want: s1 + "\n", event: &undoEvent},
		{want: s1 + "\n", event: &undoEvent},
		{want: s1 + "\n" + s2 + "\n", event: &redoEvent},
		{want: s1 + "\n" + s2 + "\nBonjour tout le monde", event: &redoEvent},
		{want: s1 + "\n" + s2 + "\nBonjour tout le mond", event: &redoEvent},
		{want: s1 + "\n" + s2 + "\nBonjour tout le mond", event: &redoEvent},
		{want: s1 + "\n" + s2 + "\nBonjour tout le monde", event: &undoEvent},
	}

	for i, c := range cases {
		if c.event != nil {
			ed.SendKeyEvent(*c.event)
		}
		got := string(ed.Contents())
		if got != c.want {
			t.Errorf("case %d\ngot:    %q\nwanted: %q\n", i, got, c.want)
		}
	}
}

func TestHistory2(t *testing.T) {
	face := basicfont.Face7x13
	ed := NewEditor(face, AcmeYellowTheme)

	last := func() address.Selection { return address.Selection{ed.LastAddress(), ed.LastAddress()} }
	ed.Load([]byte("The quick brown fox jumps over the lazy dog."))
	ed.SetDot(last())
	ed.SendKeyEvent(backspaceEvent)
	ed.Load([]byte("The quick brown fox jumps over the lazy dog."))
	ed.SetDot(last())
	ed.SendKeyEvent(backspaceEvent)
	ed.Load([]byte("The quick brown fox jumps over the lazy dog."))
	ed.SetDot(last())
	ed.SendKeyEvent(backspaceEvent)
	ed.SendKeyEvent(key.Event{Rune: ' '})
	ed.SendKeyEvent(backspaceEvent)
}
