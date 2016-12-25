package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"sigint.ca/graphics/editor/address"
)

var dir bool

func load(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	var contents []byte
	if fi.IsDir() {
		dir = true
		names, err := f.Readdirnames(0)
		if err != nil {
			return err
		}
		contents = []byte(strings.Join(names, "\n"))
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	} else {
		contents, err = ioutil.ReadAll(f)
		if err != nil {
			return err
		}
	}

	mainWidget.ed.Load(contents)
	mainWidget.ed.SetDot(address.Selection{})
	mainWidget.ed.SetSaved()
	
	currentPath = path
	savedPath = path

	return nil
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
