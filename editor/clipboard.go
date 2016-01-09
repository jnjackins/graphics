package editor

import "log"

type Clipboard interface {
	Get() ([]byte, error)
	Put([]byte) error
}

func (b *Buffer) snarf() {
	if b.Clipboard != nil {
		err := b.Clipboard.Put([]byte(b.contents(b.dot)))
		if err != nil {
			panic(err)
		}
	} else {
		log.Println("snarf: clipboard not setup")
	}
}

func (b *Buffer) paste() {
	if b.Clipboard != nil {
		buf, err := b.Clipboard.Get()
		if err != nil {
			panic(err)
		}
		b.loadBytes(buf)
	} else {
		log.Println("paste: clipboard not setup")
	}
}
