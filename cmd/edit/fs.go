package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"sigint.ca/graphics/editor/address"
	"sigint.ca/text/column"
)

func (p *pane) load(data []byte) {
	p.main.ed.Load(data)
	p.main.ed.SetDot(address.Selection{})
	p.main.ed.SetSaved()
}

func (p *pane) loadFile(path string) error {
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
		p.dir = true
		names, err := f.Readdirnames(0)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		colwriter := column.NewWriter(&buf, 100)
		for _, n := range names {
			colwriter.Write([]byte(n + "\n"))
		}
		colwriter.Flush()
		contents = buf.Bytes()
		p.cwd = path
	} else {
		contents, err = ioutil.ReadAll(f)
		if err != nil {
			return err
		}
	}
	p.cwd = getAbs(p.cwd)

	p.load(contents)

	p.currentPath = path
	p.savedPath = path

	return nil
}

func getAbs(name string) string {
	abs, err := filepath.Abs(name)
	if err != nil {
		dprintf("couldn't get absolute path for %q: %v", name, err)
	}
	if abs != "/" {
		abs += "/"
	}
	return abs

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
