package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/mobile/event/size"
)

const fontSize = 13

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

func updateFont(e size.Event) {
	dprintf("updateFont: %#v", e)
	dpi := int(e.PixelsPerPt) * 72 / e.ScaleFactor
	dprintf("updateFont: dpi: %v", dpi)

	opts := truetype.Options{
		Size:    fontSize,
		DPI:     float64(dpi),
		Hinting: font.HintingNone,
	}
	fontFace = truetype.NewFace(ttfFont, &opts)
}
