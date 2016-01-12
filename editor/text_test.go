package editor

import (
	"image"
	"testing"

	"golang.org/x/image/font/basicfont"
)

func BenchmarkPutString(b *testing.B) {
	face := basicfont.Face7x13
	ed := NewEditor(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	input := `The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。
The quick brown fox jumps over the lazy dog.`
	for i := 0; i < b.N; i++ {
		ed.putString(input)
	}
}

func BenchmarkLoadRuneOne(b *testing.B) {
	face := basicfont.Face7x13
	ed := NewEditor(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	for i := 0; i < b.N; i++ {
		ed.putString("a")
	}
}

func BenchmarkLoadRuneMany(b *testing.B) {
	face := basicfont.Face7x13
	ed := NewEditor(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			ed.putString("a")
			ed.dot.From = ed.dot.To
		}
		ed.selAll()
	}
}
