package text

import "strings"

func (b *Buffer) Search(dot Address, pattern string) (Selection, bool) {
	patternSize := measureString(pattern)

	// first, look for pattern starting from dot
	end := Address{len(b.Lines) - 1, b.Lines[len(b.Lines)-1].RuneCount()}
	contents := b.GetSel(Selection{dot, end})
	matchIndex := strings.Index(contents, pattern)

	if matchIndex < 0 {
		// try from the beginning
		searchUntil := dot.add(patternSize)
		contents = b.GetSel(b.fixSel(Selection{To: searchUntil}))
		matchIndex = strings.Index(contents, pattern)
		if matchIndex < 0 {
			// not found
			return Selection{}, false
		}
		dot = Address{}
	}

	// found, calculate selection
	from := dot.add(measureString(contents[:matchIndex]))
	to := from.add(patternSize)
	return Selection{from, to}, true
}
