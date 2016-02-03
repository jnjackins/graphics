package address

import "testing"

func TestParseAddress(t *testing.T) {
	cases := []struct {
		addr string
		text string
		want Selection
	}{
		{
			addr: "/test/",
			text: "123test123",
			want: Selection{From: Simple{0, 3}, To: Simple{0, 7}},
		},
		{
			addr: "/test/",
			text: "123\n123test123\n123",
			want: Selection{From: Simple{1, 3}, To: Simple{1, 7}},
		},
	}

	for i, c := range cases {
		addr, ok := parseAddress(c.addr)
		if !ok {
			t.Errorf("test case #%d: parseAddress: got ok=false, wanted ok=true", i)
			continue
		}
		got, ok := addr.Execute([]byte(c.text))
		if !ok {
			t.Errorf("test case #%d: Execute: got ok=false, wanted ok=true", i)
			continue
		}
		if got != c.want {
			t.Errorf("test case #%d: got %v, wanted %v", i, got, c.want)
		}
	}
}
