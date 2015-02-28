package text

import "log"

type Clipboard interface {
	Get() string
	Put(string)
}

func (b *Buffer) snarf() {
	b.dirty = true
	if b.Clipboard != nil {
		b.Clipboard.Put(b.getSel())
	} else {
		log.Println("snarf: clipboard not setup")
	}
}

func (b *Buffer) paste() {
	b.dirty = true
	if b.Clipboard != nil {
		b.load(b.Clipboard.Get())
	} else {
		log.Println("paste: clipboard not setup")
	}
}
