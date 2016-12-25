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

func (p *pane) load(path string) error {
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

	p.main.ed.Load(contents)
	p.main.ed.SetDot(address.Selection{})
	p.main.ed.SetSaved()

	p.currentPath = path
	p.savedPath = path

	return nil
}

func (p *pane) save() {
	if p.main.ed.Saved() && p.currentPath == p.savedPath {
		return
	}

	if p.currentPath == "" {
		return
	}
	f, err := os.Create(p.currentPath)
	if err != nil {
		log.Printf("error opening %q for writing: %v", p.currentPath, err)
		return
	}
	defer f.Close()

	r := bytes.NewBuffer(p.main.ed.Contents())
	if _, err := io.Copy(f, r); err != nil {
		log.Printf("error writing to %q: %v", p.currentPath, err)
		return
	}

	p.main.ed.SetSaved()
	p.savedPath = p.currentPath
}
