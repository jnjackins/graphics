package text

import "log"

type Clipboard interface {
	Get() string
	Put(string)
}

func (b *Buffer) snarf() {
	if b.Clipboard != nil {
		b.Clipboard.Put(b.contents(b.dot))
	} else {
		log.Println("snarf: clipboard not setup")
	}
}

func (b *Buffer) paste() {
	if b.Clipboard != nil {
		b.load(b.Clipboard.Get(), true)
	} else {
		log.Println("paste: clipboard not setup")
	}
}
