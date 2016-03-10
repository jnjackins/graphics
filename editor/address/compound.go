package address

type Compound struct {
}

func (a Compound) Execute(s string) (Selection, bool) {
	return Selection{}, false
}
