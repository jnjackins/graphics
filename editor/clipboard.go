package editor

import "log"

type Clipboard interface {
	Get() ([]byte, error)
	Put([]byte) error
}

func (ed *Editor) snarf() {
	if ed.Clipboard != nil {
		err := ed.Clipboard.Put([]byte(ed.buf.GetSel(ed.dot)))
		if err != nil {
			panic(err)
		}
	} else {
		log.Println("snarf: clipboard not setup")
	}
}

func (ed *Editor) paste() {
	if ed.Clipboard != nil {
		s, err := ed.Clipboard.Get()
		if err != nil {
			panic(err)
		}
		ed.putBytes(s)
	} else {
		log.Println("paste: clipboard not setup")
	}
}
