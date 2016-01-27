package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/plan9font"
)

func loadMain(s string) {
	filename = s
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Printf("error opening %q for reading: %v", filename, err)
		return
	}
	buf, err := ioutil.ReadFile(filename)
	mainWidget.ed.Load(buf)
	mainWidget.ed.SetSaved()
	f.Close()
}

func save() {
	if mainWidget.ed.Saved() {
		return
	}

	if filename == "" {
		log.Println("saving untitled file not yet supported")
		return
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("error opening %q for writing: %v", filename, err)
	}
	defer f.Close()

	r := bytes.NewBuffer(mainWidget.ed.Contents())
	if _, err := io.Copy(f, r); err != nil {
		log.Printf("error writing to %q: %v", filename, err)
	}

	mainWidget.ed.SetSaved()
}

func getfont() (font.Face, int) {
	var face font.Face
	if font := os.Getenv("PLAN9FONT"); font != "" {
		readFile := func(path string) ([]byte, error) {
			return ioutil.ReadFile(filepath.Join(filepath.Dir(font), path))
		}
		fontData, err := ioutil.ReadFile(font)
		if err != nil {
			log.Fatalf("error loading font: %v", err)
		}
		face, err = plan9font.ParseFont(fontData, readFile)
		if err != nil {
			log.Fatalf("error parsing font: %v", err)
		}

	} else {
		face = basicfont.Face7x13
	}
	bounds, _, _ := face.GlyphBounds('|')
	height := int(1.33*float64(bounds.Max.Y>>6-bounds.Min.Y>>6)) + 1
	return face, height
}
