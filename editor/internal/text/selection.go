package text

type Address struct {
	Row, Col int
}

func (a1 Address) LessThan(a2 Address) bool {
	return a1.Row < a2.Row || (a1.Row == a2.Row && a1.Col < a2.Col)
}

type Selection struct {
	From, To Address
}

func (s Selection) IsEmpty() bool {
	return s.From == s.To
}

func Sel(row1, col1, row2, col2 int) Selection {
	return Selection{From: Address{row1, col1}, To: Address{row2, col2}}
}
