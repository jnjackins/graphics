package editor

import (
	"bytes"
	"image"
	"testing"

	"golang.org/x/image/font/basicfont"
)

func TestContents(t *testing.T) {
	face := basicfont.Face7x13
	buf := NewEditor(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	input := []byte(`The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。
The quick brown fox jumps over the lazy dog.`)
	buf.putBytes(input)

	output := buf.Contents()
	if !bytes.Equal(input, output) {
		t.Errorf("expected %q, got %q", input, output)
	}
}

func BenchmarkContents(b *testing.B) {
	face := basicfont.Face7x13
	buf := NewEditor(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	input := []byte(`The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。
The quick brown fox jumps over the lazy dog.`)
	buf.putBytes(input)

	for i := 0; i < b.N; i++ {
		_ = buf.Contents()
	}
}
