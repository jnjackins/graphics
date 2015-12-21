package text

import (
	"image"
	"os"
	"testing"

	"golang.org/x/image/font/basicfont"
)

func TestMain(m *testing.M) {
	/*
		f, err := os.Create("load.prof")
		if err != nil {
			b.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	*/

	status := m.Run()

	/*
		pprof.StopCPUProfile()
		f.Close()
	*/

	os.Exit(status)
}

func BenchmarkLoadBytes(b *testing.B) {
	face := basicfont.Face7x13
	buf := NewBuffer(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	var input = []byte(`The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。`)
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
		buf.loadRune('a', false)
	}
}
