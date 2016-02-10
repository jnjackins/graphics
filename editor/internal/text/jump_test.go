package text

import (
	"testing"

	"sigint.ca/graphics/editor/internal/address"
)

func TestJumpTo(t *testing.T) {
	const text = `the 早い brown
狐 jumps over the
lazy 犬。
`
	testCases := []struct {
		addr     address.Simple
		pattern  string
		expected address.Selection
	}{
		{address.Simple{}, "/brown/", address.Selection{address.Simple{0, 7}, address.Simple{0, 12}}},
		{address.Simple{0, 7}, "/brown/", address.Selection{address.Simple{0, 7}, address.Simple{0, 12}}},
		{address.Simple{0, 12}, "/brown/", address.Selection{address.Simple{0, 7}, address.Simple{0, 12}}},
		{address.Simple{0, 8}, "/brown/", address.Selection{address.Simple{0, 7}, address.Simple{0, 12}}},
		{address.Simple{1, 0}, "/brown/", address.Selection{address.Simple{0, 7}, address.Simple{0, 12}}},
		{address.Simple{}, "/brown\n狐/", address.Selection{address.Simple{0, 7}, address.Simple{1, 1}}},
		{address.Simple{}, "/。/", address.Selection{address.Simple{2, 6}, address.Simple{2, 7}}},
	}

	b := NewBuffer()
	b.InsertString(address.Simple{}, text)

	for i, c := range testCases {
		results, ok := b.JumpTo(c.addr, c.pattern)
		if !ok {
			t.Errorf("test case %d: search failed, expected success", i)
		}
		if results != c.expected {
			t.Errorf("test case %d: got %v, wanted %v", i, results, c.expected)
		}
	}
	_, ok := b.JumpTo(address.Simple{}, "/brwn/")
	if ok {
		t.Errorf("search succeeded, expected failure")
	}
}

func BenchmarkFind5000(b *testing.B) {
	const line = `The quick brown fox jumps over the lazy dog.`
	buf := NewBuffer()

	for i := 0; i < 5000; i++ {
		buf.InsertString(address.Simple{}, line)
	}

	var dot address.Selection
	var ok bool
	for i := 0; i < b.N; i++ {
		dot, ok = buf.Find(dot.To, "fox")
		if !ok {
			b.Error("exptected ok = true, got false")
		}
	}
}
