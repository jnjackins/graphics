package editor

import (
	"testing"

	"golang.org/x/image/font/basicfont"
	"golang.org/x/mobile/event/key"
)

func TestContents(t *testing.T) {
	face := basicfont.Face7x13
	ed := NewEditor(face, AcmeYellowTheme)

	input := `The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。
The quick brown fox jumps over the lazy dog.`
	ed.putString(input)

	output := ed.Buffer.Contents()
	if input != string(output) {
		t.Errorf("expected %q, got %q", input, output)
	}
}

var bytesink []byte

func BenchmarkContents(b *testing.B) {
	face := basicfont.Face7x13
	ed := NewEditor(face, AcmeYellowTheme)

	input := `The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。
The quick brown fox jumps over the lazy dog.`
	ed.putString(input)

	for i := 0; i < b.N; i++ {
		bytesink = ed.Buffer.Contents()
	}
}

func TestSetSaved(t *testing.T) {
	face := basicfont.Face7x13
	ed := NewEditor(face, AcmeYellowTheme)

	if ed.Saved() {
		t.Error("expected Saved=false, got Saved=true")
	}

	ed.SendKeyEvent(key.Event{Rune: 'a'})
	if ed.Saved() {
		t.Error("expected Saved=false, got Saved=true")
	}

	ed.SetSaved()
	if !ed.Saved() {
		t.Error("expected Saved=true, got Saved=false")
	}

	ed.SendKeyEvent(key.Event{Rune: 'a'})
	ed.SendKeyEvent(key.Event{Rune: 'a'})
	if ed.Saved() {
		t.Error("expected Saved=false, got Saved=true")
	}

	ed.SendKeyEvent(key.Event{Code: key.CodeDeleteBackspace})
	if ed.Saved() {
		t.Error("expected Saved=false, got Saved=true")
	}

	// still false because backspace triggers history event
	ed.SendKeyEvent(key.Event{Code: key.CodeDeleteBackspace})
	if ed.Saved() {
		t.Error("expected Saved=false, got Saved=true")
	}

	ed.SetSaved()
	if !ed.Saved() {
		t.Error("expected Saved=true, got Saved=false")
	}
}
