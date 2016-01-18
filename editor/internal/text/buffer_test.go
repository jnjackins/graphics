package text

import "testing"

func TestInsertString(t *testing.T) {
	buf := NewBuffer()
	buf.InsertString(Address{0, 0}, "c")
	buf.InsertString(Address{0, 0}, "aa")
	addr := buf.InsertString(Address{0, 2}, "b")
	expected := Address{0, 3}
	if addr != expected {
		t.Errorf("got %v, wanted %v", addr, expected)
	}

	got := string(buf.Contents())
	if got != "aabc" {
		t.Errorf("got %q, wanted %q", got, "aabc")
	}

	buf.InsertString(Address{0, 0}, "こんにち")
	buf.InsertString(Address{0, 8}, "は")

	got = string(buf.Contents())
	if got != "こんにちaabcは" {
		t.Errorf("got %q, wanted %q", got, "こんにちaabcは")
	}
}

func TestInsertStringLines(t *testing.T) {
	buf := NewBuffer()

	addr := buf.InsertString(Address{0, 0}, "the 早い\nbrown 狐\njumps over the lazy 犬")
	expected := Address{2, 21}
	if addr != expected {
		t.Errorf("got %v, wanted %v", addr, expected)
	}

	got := string(buf.Contents())
	if got != "the 早い\nbrown 狐\njumps over the lazy 犬" {
		t.Errorf("got %q, wanted %q", got, "the 早い\nbrown 狐\njumps over the lazy 犬")
	}

	addr = buf.InsertString(addr, "\n")
	expected = Address{3, 0}
	if addr != expected {
		t.Errorf("got %v, wanted %v", addr, expected)
	}

	got = string(buf.Contents())
	if got != "the 早い\nbrown 狐\njumps over the lazy 犬\n" {
		t.Errorf("got %q, wanted %q", got, "the 早い\nbrown 狐\njumps over the lazy 犬\n")
	}
}

func TestInsertStringMiddle(t *testing.T) {
	buf := NewBuffer()

	buf.InsertString(Address{0, 0}, "the quick brown fox\njumps over\nthe lazy dog\n")
	buf.InsertString(Address{1, 6}, "angrily ")

	got := string(buf.Contents())
	if got != "the quick brown fox\njumps angrily over\nthe lazy dog\n" {
		t.Errorf("got %q, wanted %q", got, "the quick brown fox\njumps angrily over\n the lazy dog\n")
	}
}

func TestGetSel(t *testing.T) {
	buf := NewBuffer()
	buf.InsertString(Address{0, 0}, "the 早い\nbrown 狐\njumps over the lazy 犬")

	got := buf.GetSel(Selection{Address{0, 4}, Address{1, 5}})
	if got != "早い\nbrown" {
		t.Errorf("got %q, wanted %q", got, "早い\nbrown")
	}

	last := len(buf.Lines) - 1
	sel := Selection{Address{}, Address{last, buf.Lines[last].RuneCount()}}
	got = buf.GetSel(sel)
	if got != "the 早い\nbrown 狐\njumps over the lazy 犬" {
		t.Errorf("got %q, wanted %q", got, "the 早い\nbrown 狐\njumps over the lazy 犬")
	}
}

func TestClearSel(t *testing.T) {
	buf := NewBuffer()
	buf.InsertString(Address{0, 0}, "the 早い\nbrown 狐\njumps over the lazy 犬")
	buf.ClearSel(Selection{Address{0, 4}, Address{2, 20}})

	got := string(buf.Contents())
	if got != "the 犬" {
		t.Errorf("got %q, wanted %q", got, "the犬")
	}
}

func TestAutoSelect(t *testing.T) {
	buf := NewBuffer()
	buf.InsertString(Address{0, 0}, "こんにちは (in there)")
	got := buf.GetSel(buf.AutoSelect(Address{0, 1}))
	if got != "こんにちは" {
		t.Errorf("got %q, wanted %q", got, "こんにちは")
	}
	got = buf.GetSel(buf.AutoSelect(Address{0, 7}))
	if got != "in there" {
		t.Errorf("got %q, wanted %q", got, "in there")
	}
	got = buf.GetSel(buf.AutoSelect(Address{0, 11}))
	if got != "there" {
		t.Errorf("got %q, wanted %q", got, "there")
	}
}

func BenchmarkInsertString(b *testing.B) {
	buf := NewBuffer()
	input := `The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。
The quick brown fox jumps over the lazy dog.`
	from := Address{}
	for i := 0; i < b.N; i++ {
		to := buf.InsertString(from, input)
		buf.ClearSel(Selection{from, to})
	}
}

func BenchmarkInsertStringOne(b *testing.B) {
	buf := NewBuffer()
	from := Address{0, 0}
	for i := 0; i < b.N; i++ {
		to := buf.InsertString(from, "世")
		buf.ClearSel(Selection{from, to})
	}
}

func BenchmarkInsertStringMany(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		from := Address{0, 0}
		for j := 0; j < 200; j++ {
			from = buf.InsertString(from, "世")
		}
		buf.InsertString(from, "\n")
		buf.ClearSel(Selection{Address{0, 0}, Address{1, 0}})
	}
}
