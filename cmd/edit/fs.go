package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
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
