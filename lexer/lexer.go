package lexer

import (
	"fmt"
	"strings"
)

type TokenType int

const (
	EOF       TokenType = iota // End of file
	LineBreak                  // \n
	Indent                     // "    " (4 spaces)
	Component                  // \HeroV2, \Button, \Card
	Property                   // Title, Href, Description
	Text                       // "Hello World", "/demo"
	Number                     // 123, 3.14
	Bool                       // !Visible, !Hidden
	Colon                      // :
	Quote                      // "
)

func (t TokenType) String() string {
	return [...]string{
		"EOF",
		"LineBreak",
		"Indent",
		"Component",
		"Property",
		"Text",
		"Number",
		"Bool",
		"Colon",
		"Quote",
	}[t]
}

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func New(input string) *Lexer {
	return &Lexer{
		input:  input,
		line:   1,
		column: 1,
	}
}

type Lexer struct {
	input  string
	pos    int
	line   int
	column int
	prev   Token
}

func (l *Lexer) next() Token {
	for {
		if l.isEOF() {
			return Token{Type: EOF}
		}

		if l.isIndent() {
			return l.lexIndent()
		}

		if l.prev.Type == Quote && l.lookAhead('"', '\n') != -1 {
			return l.lexString()
		}

		start := l.pos
		line := l.line
		column := l.column

		switch l.current() {
		case ' ':
			l.advance()
			continue
		case '\\':
			return l.lexComponent()
		case '"':
			l.advance()
			return Token{
				Type:   Quote,
				Value:  "\"",
				Line:   line,
				Column: column,
			}
		case ':':
			l.advance()
			return Token{
				Type:   Colon,
				Value:  ":",
				Line:   line,
				Column: column,
			}
		case '\n':
			return l.lexNewline()
		case '!':
			return l.lexBoolProperty()
		}

		if l.isProperty() {
			return l.lexProperty()
		}

		for l.pos < len(l.input) && l.current() != '\n' {
			l.advance()
		}
		return Token{
			Type:   Text,
			Value:  l.input[start:l.pos],
			Line:   line,
			Column: column,
		}
	}
}

func (l *Lexer) lookAhead(v, t byte) int {
	i := l.pos
	for i < len(l.input) {
		if l.input[i] == v {
			return i
		}
		if l.input[i] == t {
			return -1
		}
		i++
	}
	return -1
}

func (l *Lexer) Prev() Token {
	return l.prev
}

func (l *Lexer) Next() Token {
	l.prev = l.next()
	return l.prev
}

func (l *Lexer) Peek() Token {
	start := l.pos
	line := l.line
	column := l.column
	token := l.next()
	l.pos = start
	l.line = line
	l.column = column
	return token
}

func (l *Lexer) isEOF() bool {
	return l.pos >= len(l.input)
}

func (l *Lexer) lexComponent() Token {
	column := l.column
	l.advance() // skip backslash
	start := l.pos

	for l.pos < len(l.input) && isAlphaNumeric(l.current()) {
		l.advance()
	}

	return Token{
		Type:   Component,
		Value:  l.input[start:l.pos],
		Line:   l.line,
		Column: column,
	}
}

func (l *Lexer) lexBoolProperty() Token {
	l.advance() // skip !
	start := l.pos

	for l.pos < len(l.input) && isAlphaNumeric(l.current()) {
		l.advance()
	}

	return Token{
		Type:   Bool,
		Value:  l.input[start:l.pos],
		Line:   l.line,
		Column: l.column,
	}
}

func (l *Lexer) isIndent() bool {
	if l.pos == 0 {
		return false
	}
	if l.current() != ' ' && l.current() != '\t' {
		return false
	}
	return l.prev.Type == Indent || l.prev.Type == LineBreak
}

func (l *Lexer) isProperty() bool {
	end := l.lookAhead(':', '\n')
	if end == -1 {
		return false
	}
	for i := l.pos; i < end; i++ {
		if !isAlphaNumeric(l.input[i]) {
			return false
		}
	}
	return true
}

func (l *Lexer) lexProperty() Token {
	start := l.pos
	line := l.line
	column := l.column

	for l.pos < len(l.input) && l.current() != ':' {
		l.advance()
	}

	return Token{
		Type:   Property,
		Value:  l.input[start:l.pos],
		Line:   line,
		Column: column,
	}
}

func (l *Lexer) lexIndent() Token {
	line := l.line
	column := l.column
	token := Token{
		Type:   Indent,
		Value:  "\t",
		Line:   line,
		Column: column,
	}
	if l.current() == '\t' {
		l.advance()
		return token
	}
	size := 0
	for l.current() == ' ' {
		size++
		l.advance()
		if size == 4 {
			return token
		}
	}
	panic(fmt.Sprintf("invalid indent at line %d, column %d", line, column))
}

func (l *Lexer) lexNewline() Token {
	line := l.line
	column := l.column
	l.advance()
	return Token{
		Type:   LineBreak,
		Value:  "\n",
		Line:   line,
		Column: column - 1,
	}
}

func (l *Lexer) lexString() Token {
	start := l.pos
	line := l.line
	column := l.column

	var prev byte
	for l.pos < len(l.input) && l.current() != '"' && prev != '\\' {
		prev = l.current()
		l.advance()
	}

	return Token{
		Type:   Text,
		Value:  strings.Replace(l.input[start:l.pos], "\\\"", "\"", -1),
		Line:   line,
		Column: column,
	}
}

// Helper methods
func (l *Lexer) current() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) advance() {
	if l.pos >= len(l.input) {
		return
	}
	if l.input[l.pos] == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	l.pos++
}

func isAlphaNumeric(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}
