package text

import "testing"

func TestSearch(t *testing.T) {
	const text = `the 早い brown
狐 jumps over the
lazy 犬。
`
	testCases := []struct {
		addr     Address
		pattern  string
		expected Selection
	}{
		{Address{}, "brown", Selection{Address{0, 7}, Address{0, 12}}},
		{Address{0, 7}, "brown", Selection{Address{0, 7}, Address{0, 12}}},
		{Address{0, 12}, "brown", Selection{Address{0, 7}, Address{0, 12}}},
		{Address{0, 8}, "brown", Selection{Address{0, 7}, Address{0, 12}}},
		{Address{1, 0}, "brown", Selection{Address{0, 7}, Address{0, 12}}},
		{Address{}, "brown\n狐", Selection{Address{0, 7}, Address{1, 1}}},
		{Address{}, "。", Selection{Address{2, 6}, Address{2, 7}}},
	}

	b := NewBuffer()
	b.InsertString(Address{}, text)

	for i, c := range testCases {
		results, ok := b.Search(c.addr, c.pattern)
		if !ok {
			t.Errorf("test case %d: search failed, expected success", i)
		}
		if results != c.expected {
			t.Errorf("test case %d: got %v, wanted %v", i, results, c.expected)
		}
	}
	_, ok := b.Search(Address{}, "brwn")
	if ok {
		t.Errorf("search succeeded, expected failure")
	}
}
