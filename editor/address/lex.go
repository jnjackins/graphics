package address

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type token struct {
	typ tokenType // The type of this token.
	val string    // The value of this token.
}

type tokenType int

const (
	tokenRegexpDelim tokenType = iota
	tokenReverseRegexpDelim
	tokenRegexp
	tokenError
)

const (
	regexpDelim        = '/'
	reverseRegexpDelim = '?'
)

const eof = -1

// lexer holds the state of the scanner.
type lexer struct {
	name   string     // used only for error reports.
	input  string     // the string being scanned.
	start  int        // start position of this token.
	pos    int        // current position in the input.
	width  int        // width of last rune read from input.
	tokens chan token // channel of scanned tokens.
}

func lex(name, input string) chan token {
	l := &lexer{
		name:   name,
		input:  input,
		tokens: make(chan token),
	}
	go l.run() // Concurrently run state machine.
	return l.tokens
}

// run lexes the input by executing state functions until
// the state is nil.
func (l *lexer) run() {
	for state := lexAny; state != nil; {
		state = state(l)
	}
	close(l.tokens) // No more tokens will be delivered.
}

// emit passes an token back to the client.
func (l *lexer) emit(t tokenType) {
	l.tokens <- token{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	var r rune
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// stateFn represents the state of the scanner
// as a function that returns the next state.
type stateFn func(*lexer) stateFn

func lexAny(l *lexer) stateFn {
	switch l.peek() {
	case regexpDelim:
		return lexRegexpDelim1(l)
	case eof:
		return nil
	default:
		return l.errorf("couldn't lex %#v", l.peek())
	}
}

func lexRegexpDelim1(l *lexer) stateFn {
	l.pos++
	l.emit(tokenRegexpDelim)
	return lexRegexp
}

func lexRegexpDelim2(l *lexer) stateFn {
	l.pos++
	l.emit(tokenRegexpDelim)
	return lexAny
}

func lexRegexp(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], string(regexpDelim)) {
			l.emit(tokenRegexp)
			return lexRegexpDelim2
		}
		r := l.next()
		if r == eof {
			return l.errorf("unclosed regexp")
		}
	}
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{
		tokenError,
		fmt.Sprintf(format, args...),
	}
	return nil
}
