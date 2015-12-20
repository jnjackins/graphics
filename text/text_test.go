package text

import (
	"image"
	"testing"

	"golang.org/x/image/font/basicfont"
)

func BenchmarkLoad(b *testing.B) {
	face := basicfont.Face7x13
	buf := NewBuffer(image.Pt(100, 100), face, face.Height, AcmeYellowTheme)

	/*
		f, err := os.Create("load.prof")
		if err != nil {
			b.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	*/

	const input = `The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。`
	for i := 0; i < b.N; i++ {
		buf.load(input, false)
	}
}
