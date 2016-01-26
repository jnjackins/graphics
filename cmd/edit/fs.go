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
	f.Close()
}

func save() {
	if filename == "" {
		log.Println("saving untitled file not yet supported")
		return
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("error opening %q for writing: %v", filename, err)
	}
	r := bytes.NewBuffer(mainWidget.ed.Contents())
	if _, err := io.Copy(f, r); err != nil {
		log.Printf("error writing to %q: %v", filename, err)
	}
	f.Close()
}

func getfont() font.Face {
	if font := os.Getenv("PLAN9FONT"); font != "" {
		readFile := func(path string) ([]byte, error) {
			return ioutil.ReadFile(filepath.Join(filepath.Dir(font), path))
		}
		fontData, err := ioutil.ReadFile(font)
		if err != nil {
			log.Fatalf("error loading font: %v", err)
		}
		face, err := plan9font.ParseFont(fontData, readFile)
		if err != nil {
			log.Fatalf("error parsing font: %v", err)
		}
		return face
	}
	return basicfont.Face7x13
}