package text

import (
	"image"
	"os"
	"testing"

	"golang.org/x/image/font/basicfont"
)

func TestMain(m *testing.M) {
	//f, _ := os.Create("load.prof")
	//pprof.StartCPUProfile(f)
	status := m.Run()
	//pprof.StopCPUProfile()
	//f.Close()
	os.Exit(status)
}

func BenchmarkLoadBytes(b *testing.B) {
	face := basicfont.Face7x13
	buf := NewBuffer(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	input := []byte(`The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。
The quick brown fox jumps over the lazy dog.`)
	for i := 0; i < b.N; i++ {
		buf.loadBytes(input, false)
	}
}

func BenchmarkLoadRuneASCII(b *testing.B) {
	face := basicfont.Face7x13
	buf := NewBuffer(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	for i := 0; i < b.N; i++ {
		buf.loadRune('a', false)
	}
}

func BenchmarkLoadRuneUnicode(b *testing.B) {
	face := basicfont.Face7x13
	buf := NewBuffer(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	for i := 0; i < b.N; i++ {
		buf.loadRune('せ', false)
	}
}
