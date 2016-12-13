package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"

	"sigint.ca/graphics/editor/address"
)

func loadMain(path string) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Printf("error opening %q for reading: %v", path, err)
		return
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	mainWidget.ed.Load(contents)
	mainWidget.ed.SetDot(address.Selection{})
	mainWidget.ed.SetSaved()

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

	r := bytes.NewBuffer(mainWidget.ed.Contents())
	if _, err := io.Copy(f, r); err != nil {
		log.Printf("error writing to %q: %v", currentPath, err)
		return
	}

	mainWidget.ed.SetSaved()
	savedPath = currentPath
}
