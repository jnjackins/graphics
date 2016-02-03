package text

import (
	"testing"

	"sigint.ca/graphics/editor/internal/address"
)

func TestInsertString(t *testing.T) {
	buf := NewBuffer()
	buf.InsertString(address.Simple{0, 0}, "c")
	buf.InsertString(address.Simple{0, 0}, "aa")
	addr := buf.InsertString(address.Simple{0, 2}, "b")
	expected := address.Simple{0, 3}
	if addr != expected {
		t.Errorf("got %v, wanted %v", addr, expected)
	}

	got := string(buf.Contents())
	if got != "aabc" {
		t.Errorf("got %q, wanted %q", got, "aabc")
	}

	buf.InsertString(address.Simple{0, 0}, "こんにち")
	buf.InsertString(address.Simple{0, 8}, "は")

	got = string(buf.Contents())
	if got != "こんにちaabcは" {
		t.Errorf("got %q, wanted %q", got, "こんにちaabcは")
	}
}

func TestInsertStringLines(t *testing.T) {
	buf := NewBuffer()

	addr := buf.InsertString(address.Simple{0, 0}, "the 早い\nbrown 狐\njumps over the lazy 犬")
	expected := address.Simple{2, 21}
	if addr != expected {
		t.Errorf("got %v, wanted %v", addr, expected)
	}

	got := string(buf.Contents())
	if got != "the 早い\nbrown 狐\njumps over the lazy 犬" {
		t.Errorf("got %q, wanted %q", got, "the 早い\nbrown 狐\njumps over the lazy 犬")
	}

	addr = buf.InsertString(addr, "\n")
	expected = address.Simple{3, 0}
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

	buf.InsertString(address.Simple{0, 0}, "the quick brown fox\njumps over\nthe lazy dog\n")
	buf.InsertString(address.Simple{1, 6}, "angrily ")

	got := string(buf.Contents())
	if got != "the quick brown fox\njumps angrily over\nthe lazy dog\n" {
		t.Errorf("got %q, wanted %q", got, "the quick brown fox\njumps angrily over\n the lazy dog\n")
	}
}

func TestGetSel(t *testing.T) {
	buf := NewBuffer()
	buf.InsertString(address.Simple{0, 0}, "the 早い\nbrown 狐\njumps over the lazy 犬")

	got := buf.GetSel(address.Selection{address.Simple{0, 4}, address.Simple{1, 5}})
	if got != "早い\nbrown" {
		t.Errorf("got %q, wanted %q", got, "早い\nbrown")
	}

	last := len(buf.Lines) - 1
	sel := address.Selection{address.Simple{}, address.Simple{last, buf.Lines[last].RuneCount()}}
	got = buf.GetSel(sel)
	if got != "the 早い\nbrown 狐\njumps over the lazy 犬" {
		t.Errorf("got %q, wanted %q", got, "the 早い\nbrown 狐\njumps over the lazy 犬")
	}
}

func TestClearSel(t *testing.T) {
	buf := NewBuffer()
	buf.InsertString(address.Simple{0, 0}, "the 早い\nbrown 狐\njumps over the lazy 犬")
	buf.ClearSel(address.Selection{address.Simple{0, 4}, address.Simple{2, 20}})

	got := string(buf.Contents())
	if got != "the 犬" {
		t.Errorf("got %q, wanted %q", got, "the犬")
	}
}

func TestAutoSelect(t *testing.T) {
	buf := NewBuffer()
	buf.InsertString(address.Simple{0, 0}, "こんにちは (in there)")
	got := buf.GetSel(buf.AutoSelect(address.Simple{0, 1}))
	if got != "こんにちは" {
		t.Errorf("got %q, wanted %q", got, "こんにちは")
	}
	got = buf.GetSel(buf.AutoSelect(address.Simple{0, 7}))
	if got != "in there" {
		t.Errorf("got %q, wanted %q", got, "in there")
	}
	got = buf.GetSel(buf.AutoSelect(address.Simple{0, 11}))
	if got != "there" {
		t.Errorf("got %q, wanted %q", got, "there")
	}
}

func BenchmarkInsertString(b *testing.B) {
	buf := NewBuffer()
	input := `The quick brown fox jumps over the lazy dog.
速い茶色のキツネは、のろまなイヌに飛びかかりました。
The quick brown fox jumps over the lazy dog.`
	from := address.Simple{}
	for i := 0; i < b.N; i++ {
		to := buf.InsertString(from, input)
		buf.ClearSel(address.Selection{from, to})
	}
}

func BenchmarkInsertStringOne(b *testing.B) {
	buf := NewBuffer()
	from := address.Simple{0, 0}
	for i := 0; i < b.N; i++ {
		to := buf.InsertString(from, "世")
		buf.ClearSel(address.Selection{from, to})
	}
}

func BenchmarkInsertStringMany(b *testing.B) {
	buf := NewBuffer()
	for i := 0; i < b.N; i++ {
		from := address.Simple{0, 0}
		for j := 0; j < 200; j++ {
			from = buf.InsertString(from, "世")
		}
		buf.InsertString(from, "\n")
		buf.ClearSel(address.Selection{address.Simple{0, 0}, address.Simple{1, 0}})
	}
}

func BenchmarkGetSelLine(b *testing.B) {
	buf := NewBuffer()

	buf.InsertString(address.Simple{}, "the 早い brown 狐 jumps over the lazy 犬\n")

	sel := address.Selection{address.Simple{}, address.Simple{1, 0}}
	for i := 0; i < b.N; i++ {
		buf.GetSel(sel)
	}
}

func BenchmarkGetSelLarge(b *testing.B) {
	buf := NewBuffer()

	for i := 0; i < 1000; i++ {
		buf.InsertString(address.Simple{}, "the 早い brown 狐 jumps over the lazy 犬\n")
	}

	last := len(buf.Lines) - 1
	sel := address.Selection{address.Simple{}, address.Simple{last, 0}}
	for i := 0; i < b.N; i++ {
		buf.GetSel(sel)
	}
}

func BenchmarkGetSelHuge(b *testing.B) {
	buf := NewBuffer()

	for i := 0; i < 10000; i++ {
		buf.InsertString(address.Simple{}, "the 早い brown 狐 jumps over the lazy 犬\n")
	}

	last := len(buf.Lines) - 1
	sel := address.Selection{address.Simple{}, address.Simple{last, 0}}
	for i := 0; i < b.N; i++ {
		buf.GetSel(sel)
	}
}

func BenchmarkAutoSelectLarge(b *testing.B) {
	buf := NewBuffer()

	for i := 0; i < 1000; i++ {
		buf.InsertString(address.Simple{}, "the 早い brown 狐 jumps over the lazy 犬\n")
	}
	buf.InsertString(address.Simple{}, "{")
	buf.InsertString(address.Simple{Row: len(buf.Lines) - 1}, "}")

	got := buf.AutoSelect(address.Simple{0, 1})
	want := address.Selection{
		address.Simple{0, 1},
		address.Simple{len(buf.Lines) - 1, 0},
	}

	if got != want {
		b.Error("got %v, wanted %v", got, want)
	}

	for i := 0; i < b.N; i++ {
		buf.AutoSelect(address.Simple{0, 1})
	}
}
