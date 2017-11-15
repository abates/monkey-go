package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/abates/monkey-go/token"
)

const eof = -1

type stateFn func(*goLex) stateFn

type goLex struct {
	state  stateFn
	input  string
	start  int
	pos    int
	width  int
	tokens chan token.Token
}

func NewGoLex(input string) *goLex {
	l := &goLex{
		state:  lex,
		input:  input,
		tokens: make(chan token.Token, 2),
	}

	return l
}

func (l *goLex) NextToken() token.Token {
	for {
		select {
		case t := <-l.tokens:
			return t
		default:
			if l.state == nil {
				return token.Token{token.EOF, ""}
			}

			l.state = l.state(l)
		}
	}
	panic("Not possible")
}

func (l *goLex) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return
}

func (l *goLex) ignoreWhitespace() {
	for isSpace(l.next()) {
		l.ignore()
	}
	l.backup()
}

func (l *goLex) ignore() {
	l.start = l.pos
}

func (l *goLex) backup() {
	l.pos -= l.width
}

func (l *goLex) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *goLex) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *goLex) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func lex(l *goLex) stateFn {
	l.ignoreWhitespace()
	if l.accept("!=") {
		if l.accept("=") {
			l.emit(token.TokenType(l.input[l.start:l.pos]))
		} else {
			l.emit(token.TokenType(l.input[l.start:l.pos]))
		}
		return lex
	} else if l.accept(";(),+-/*<>{}") {
		l.emit(token.TokenType(l.input[l.start:l.pos]))
		return lex
	} else if isAlpha(l.peek()) {
		return lexIdentifier
	} else if isNumber(l.peek()) {
		return lexNumber
	} else if l.peek() == eof {
		return nil
	}
	return l.errorf("Illegal character: %q", l.next())
}

func (l *goLex) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token.Token{
		Type:    token.ILLEGAL,
		Literal: fmt.Sprintf(format, args...),
	}
	return nil
}

func (l *goLex) emit(tokenType token.TokenType) {
	l.tokens <- token.Token{
		Type:    tokenType,
		Literal: l.input[l.start:l.pos],
	}
	l.start = l.pos
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

func isNumber(r rune) bool {
	return '0' <= r && r <= '9'
}

func isAlpha(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || r == '_'
}

func lexIdentifier(l *goLex) stateFn {
	for {
		if !isAlpha(l.next()) {
			l.backup()
			l.emit(token.LookupIdent(l.input[l.start:l.pos]))
			break
		}
	}
	return lex
}

func lexNumber(l *goLex) stateFn {
	digits := "0123456789"
	l.acceptRun(digits)
	if isAlpha(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(token.INT)
	return lex
}
