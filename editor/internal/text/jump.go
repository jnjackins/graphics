// TODO: "address" name collision

package text

import "sigint.ca/graphics/editor/internal/address"

func (b *Buffer) JumpTo(dot address.Simple, addr string) (address.Selection, bool) {
	parsed, ok := address.ParseAddress(addr)
	if !ok {
		return address.Selection{}, false
	}
	return b.jumpTo(dot, parsed)
}

func (b *Buffer) Find(dot address.Simple, s string) (address.Selection, bool) {
	return b.jumpTo(dot, address.Substring(s))
}

// TODO: slow for large files, probably because of GetSel
func (b *Buffer) jumpTo(dot address.Simple, parsed address.Address) (address.Selection, bool) {
	end := address.Simple{Row: len(b.Lines) - 1, Col: b.Lines[len(b.Lines)-1].RuneCount()}
	contents := b.GetSel(address.Selection{dot, end})
	sel, ok := parsed.Execute(contents)
	if ok {
		sel.From = dot.Add(sel.From)
		sel.To = dot.Add(sel.To)
		return sel, true
	}

	// try again from the beginning
	contents = b.GetSel(address.Selection{To: end})
	sel, ok = parsed.Execute(contents)
	if ok {
		return sel, true
	}
	return address.Selection{}, false
}
