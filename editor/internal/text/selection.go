package text

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Selection struct {
	From, To Address
}

func (s Selection) String() string {
	return fmt.Sprintf("%v-%v", s.From, s.To)
}

func (s Selection) IsEmpty() bool {
	return s.From == s.To
}

type Address struct {
	Row, Col int
}

func measureString(s string) Address {
	row := strings.Count(s, "\n")
	rowIndex := 0
	if row > 0 {
		rowIndex = strings.LastIndex(s, "\n") + 1
	}
	col := utf8.RuneCountInString(s[rowIndex:])
	return Address{row, col}
}

func (a Address) String() string {
	return fmt.Sprintf("(%d,%d)", a.Row, a.Col)
}

func (a1 Address) add(a2 Address) Address {
	sum := Address{Row: a1.Row + a2.Row}
	if a2.Row > 0 {
		sum.Col = a2.Col
	} else {
		sum.Col = a1.Col + a2.Col
	}
	return sum
}

func (a1 Address) LessThan(a2 Address) bool {
	return a1.Row < a2.Row || (a1.Row == a2.Row && a1.Col < a2.Col)
}
