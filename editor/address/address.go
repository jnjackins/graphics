package address

import "fmt"

type Address interface {
	Execute(string) (Selection, bool)
}

type Selection struct {
	From, To Simple
}

func (s Selection) String() string {
	return fmt.Sprintf("%v-%v", s.From, s.To)
}

func (s Selection) IsEmpty() bool {
	return s.From == s.To
}

type Simple struct {
	Row, Col int
}

func (a1 Simple) Add(a2 Simple) Simple {
	sum := Simple{Row: a1.Row + a2.Row}
	if a2.Row > 0 {
		sum.Col = a2.Col
	} else {
		sum.Col = a1.Col + a2.Col
	}
	return sum
}

func (a1 Simple) LessThan(a2 Simple) bool {
	return a1.Row < a2.Row || (a1.Row == a2.Row && a1.Col < a2.Col)
}

func (a Simple) In(sel Selection) bool {
	return sel.From.LessThan(a) && a.LessThan(sel.To)
}

func (a Simple) Execute(s string) (Selection, bool) {
	return Selection{From: a, To: a}, true
}

func (a Simple) String() string {
	return fmt.Sprintf("(%d,%d)", a.Row, a.Col)
}
