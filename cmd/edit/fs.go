package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/freetype/truetype"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/gomono"
)

func loadMain(path string) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Printf("error opening %q for reading: %v", path, err)
		return
	}
	buf, err := ioutil.ReadFile(path)
	mainWidget.ed.Load(buf)
	mainWidget.ed.Dot.To = mainWidget.ed.Dot.From
	mainWidget.ed.SetSaved()
	f.Close()

	savedPath = path
}

func save() {
	if mainWidget.ed.Saved() && currentPath == savedPath {
		return
	}

	if currentPath == "" {
		return
	}
	f, err := os.Create(currentPath)
	if err != nil {
		log.Printf("error opening %q for writing: %v", currentPath, err)
		return
	}
	defer f.Close()

	r := bytes.NewBuffer(mainWidget.ed.Buffer.Contents())
	if _, err := io.Copy(f, r); err != nil {
		log.Printf("error writing to %q: %v", currentPath, err)
		return
	}

	mainWidget.ed.SetSaved()
	savedPath = currentPath
}

func getfont() font.Face {
	ttf := gomono.TTF
	if font := os.Getenv("FONT"); font != "" {
		if buf, err := ioutil.ReadFile(font); err == nil {
			ttf = buf
		} else {
			log.Printf("error reading FONT=%s: %v", font, err)
		}
	}

	var face font.Face
	font, err := truetype.Parse(ttf)
	if err == nil {
		face = truetype.NewFace(font, nil)
	} else {
		log.Printf("error parsing ttf: %v", err)
		face = basicfont.Face7x13
	}

	return face
}
