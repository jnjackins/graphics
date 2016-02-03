package address

import "regexp"

func ParseAddress(addr string) (Address, bool) {
	ch := lex("ParseAddress", addr)
	for tok := range ch {
		switch tok.typ {
		case tokenRegexpDelim:
			return parseRegexp(ch)
		}
	}
	return nil, false
}

func parseRegexp(ch chan token) (*Regexp, bool) {
	tok := <-ch
	if tok.typ != tokenRegexp {
		return nil, false
	}
	pattern := tok.val
	tok = <-ch
	if tok.typ != tokenRegexpDelim {
		return nil, false
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, false
	}
	return &Regexp{re: re}, true
}
