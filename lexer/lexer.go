package lexer

import (
	"github.com/basis-ex/token"
	"strings"
	"unicode"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	tok.Line = l.line

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch), Line: l.line}
		} else {
			tok = newToken(token.ASSIGN, l.ch, l.line)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch, l.line)
	case '-':
		tok = newToken(token.MINUS, l.ch, l.line)
	case '*':
		tok = newToken(token.MULT, l.ch, l.line)
	case '/':
		tok = newToken(token.DIV, l.ch, l.line)
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LE, Literal: string(ch) + string(l.ch), Line: l.line}
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NE, Literal: string(ch) + string(l.ch), Line: l.line}
		} else {
			tok = newToken(token.LT, l.ch, l.line)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GE, Literal: string(ch) + string(l.ch), Line: l.line}
		} else {
			tok = newToken(token.GT, l.ch, l.line)
		}
	case '(':
		tok = newToken(token.LPAREN, l.ch, l.line)
	case ')':
		tok = newToken(token.RPAREN, l.ch, l.line)
	case ',':
		tok = newToken(token.COMMA, l.ch, l.line)
	case ':':
		tok = newToken(token.COLON, l.ch, l.line)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch, l.line)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = l.line
	case '\n':
		tok = newToken(token.NEWLINE, l.ch, l.line)
		l.line++
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.Line = l.line
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(strings.ToUpper(tok.Literal))
			tok.Line = l.line
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.NUMBER
			tok.Literal = l.readNumber()
			tok.Line = l.line
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch, l.line)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	l.readChar()
	position := l.position
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\n' {
			l.line++
		}
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.TokenType, ch byte, line int) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Line: line}
}
