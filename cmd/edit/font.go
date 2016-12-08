package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
)

var (
	ttfFont  *truetype.Font
	fontFace font.Face
)

func loadFont() {
	ttfBytes := gomono.TTF
	if path := os.Getenv("FONT"); path != "" {
		if buf, err := ioutil.ReadFile(path); err == nil {
			ttfBytes = buf
		} else {
			log.Printf("error reading FONT=%s: %v", path, err)
		}
	}

	var err error
	ttfFont, err = truetype.Parse(ttfBytes)
	if err != nil {
		log.Fatalf("error parsing ttf data: %v", err)
	}
}

func updateFont() {
	dpi := (float64(pixelsPerPt) * 72.0) / 2

	opts := truetype.Options{
		Size:    13,
		DPI:     dpi,
		Hinting: font.HintingNone,
	}
	fontFace = truetype.NewFace(ttfFont, &opts)
}
