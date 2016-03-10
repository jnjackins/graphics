package address

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type Regexp struct {
	re *regexp.Regexp
}

func (a Regexp) Execute(s string) (Selection, bool) {
	return getMatch(a.re.FindStringIndex(s), s)
}

type Substring string

func (a Substring) Execute(s string) (Selection, bool) {
	i := strings.Index(s, string(a))
	if i < 0 {
		return Selection{}, false
	}
	return getMatch([]int{i, i + len(string(a))}, s)
}

func getMatch(loc []int, s string) (Selection, bool) {
	if loc == nil {
		return Selection{}, false
	}
	row1 := strings.Count(s[:loc[0]], "\n")
	row2 := row1 + strings.Count(s[loc[0]:loc[1]], "\n")
	var i int
	if row1 > 0 {
		i = strings.LastIndexByte(s[:loc[0]], '\n') + 1
	}
	col1 := utf8.RuneCountInString(s[i:loc[0]])
	if row2 > row1 {
		i = loc[0] + strings.LastIndexByte(s[loc[0]:loc[1]], '\n') + 1
	}
	col2 := utf8.RuneCountInString(s[i:loc[1]])
	return Selection{From: Simple{row1, col1}, To: Simple{row2, col2}}, true
}
