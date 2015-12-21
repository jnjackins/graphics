package text

import "log"

type Clipboard interface {
	Get() []byte
	Put([]byte)
}

func (b *Buffer) snarf() {
	if b.Clipboard != nil {
		b.Clipboard.Put(b.contents(b.Dot))
	} else {
		log.Println("snarf: clipboard not setup")
	}
}

func (b *Buffer) paste() {
	if b.Clipboard != nil {
		b.loadBytes(b.Clipboard.Get(), true)
	} else {
		log.Println("paste: clipboard not setup")
	}
}
