package editor

import (
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype/truetype"
)

var advSliceSink []fixed.Int26_6

func BenchmarkMeasureString(b *testing.B) {
	ttf, err := truetype.Parse(goregular.TTF)
	if err != nil {
		b.Fatal(err)
	}
	face := truetype.NewFace(ttf, nil)
	ed := NewEditor(face, SimpleTheme)
	s := "the quick brown fox jumps over the lazy dog. the quick brown fox jumps over the lazy dog."

	for i := 0; i < b.N; i++ {
		advSliceSink = ed.measureString(s)
	}
}

var advSink fixed.Int26_6

func BenchmarkMeasureString2(b *testing.B) {
	ttf, err := truetype.Parse(goregular.TTF)
	if err != nil {
		b.Fatal(err)
	}
	face := truetype.NewFace(ttf, nil)
	s := "the quick brown fox jumps over the lazy dog. the quick brown fox jumps over the lazy dog."

	for i := 0; i < b.N; i++ {
		advSink = font.MeasureString(face, s)
	}
}
