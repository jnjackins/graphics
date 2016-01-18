package editor

import (
	"image"
	"testing"

	"golang.org/x/image/font/basicfont"
)

func TestContents(t *testing.T) {
	face := basicfont.Face7x13
	ed := NewEditor(image.Pt(100, 100), face, AcmeYellowTheme)

	input := `The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。
The quick brown fox jumps over the lazy dog.`
	ed.putString(input)

	output := ed.Contents()
	if input != string(output) {
		t.Errorf("expected %q, got %q", input, output)
	}
}

func BenchmarkContents(b *testing.B) {
	face := basicfont.Face7x13
	ed := NewEditor(image.Pt(100, 100), face, AcmeYellowTheme)

	input := `The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。
The quick brown fox jumps over the lazy dog.`
	ed.putString(input)

	for i := 0; i < b.N; i++ {
		_ = ed.Contents()
	}
}
