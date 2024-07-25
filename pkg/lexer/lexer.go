package lexer

import (
	"bytes"
	"io"
	"log"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
)

type lexer struct {
	input        []byte
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
	logger       *log.Logger
}

func Lex(input io.Reader) []tokens.Token {
	l := &lexer{}
	l.init(input)
	return l.tokenize()
}

func (l *lexer) init(input io.Reader) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, input)
	if err != nil {
		log.Fatalf("Fatal error reading input: %v", err)
	}
	l.input = buf.Bytes()
	l.position = 0
	l.readPosition = 0
	l.ch = 0
	l.line = 1
	l.column = 0
	l.readChar()
}

func (l *lexer) tokenize() []tokens.Token {
	var t []tokens.Token
	for {
		tok := l.nextToken()
		t = append(t, tok)
		if tok.Type == tokens.EOF {
			break
		}
	}
	return t
}

func (l *lexer) nextToken() tokens.Token {
	l.skipWhitespace()
	tok := tokens.Token{
		Line:   l.line,
		Column: l.column,
	}
	start := l.position

	if l.ch == 0 {
		tok.Type = tokens.EOF
		tok.Literal = []byte{}
		return tok
	}

	switch l.ch {
	case ';':
		tok.Type = tokens.SEMICOLON
	case ',':
		tok.Type = tokens.COMMA
	case '(':
		tok.Type = tokens.LPAREN
	case ')':
		tok.Type = tokens.RPAREN
	case '{':
		tok.Type = tokens.LBRACE
	case '}':
		tok.Type = tokens.RBRACE
	case '[':
		tok.Type = tokens.LBRACKET
	case ']':
		tok.Type = tokens.RBRACKET
	case '=':
		tok.Type = tokens.EQUALS
	case '+':
		tok.Type = tokens.PLUS
	case '>':
		tok.Type = tokens.GREATER
	case '~':
		tok.Type = tokens.TILDE
	case '|':
		tok.Type = tokens.PIPE
	case '^':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = tokens.STARTS_WITH
		} else {
			tok.Type = tokens.ILLEGAL
		}
	case '%':
		tok.Type = tokens.PERCENTAGE
	case '$':
		tok.Type = tokens.DOLLAR
	case '!':
		tok.Type = tokens.EXCLAMATION
	case '@':
		tok.Type = tokens.AT
	case '*':
		tok.Type = tokens.ASTERISK
	case ':':
		if l.peekChar() == ':' {
			l.readChar()
			tok.Type = tokens.DBLCOLON
		} else {
			tok.Type = tokens.COLON
		}
	case '.':
		tok.Type = tokens.DOT
	case 0:
		tok.Type = tokens.EOF
	case '#':
		tok.Type = l.readHashOrColor()
	case '-':
		tok.Type = l.handleDash()
	case '\\':
		tok.Type = tokens.IDENT
		l.readIdentifier()
	case '/':
		tok.Type = l.handleSlash()
	case '"', '\'':
		tok.Type = tokens.STRING
		l.readString()
	default:
		if isLetter(l.ch) {
			tok.Type = tokens.IDENT
			l.readIdentifier()
		} else if isDigit(l.ch) {
			tok.Type = tokens.NUMBER
			l.readNumber()
		} else {
			tok.Type = tokens.ILLEGAL
		}
	}

	if l.ch != 0 {
		l.readChar()
	}

	end := l.position
	tok.Literal = l.getLiteral(start, end)

	return tok
}

func (l *lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *lexer) getLiteral(start, end int) []byte {
	if start > end || start >= len(l.input) {
		return []byte{}
	}

	end = min(end, len(l.input))
	return l.input[start:end]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (l *lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *lexer) peekNextChar() byte {
	if l.readPosition+1 >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition+1]
}

func (l *lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *lexer) handleSlash() tokens.TokenType {
	if l.peekChar() == '*' {
		l.readChar() // consume '*'
		l.readComment()
		return tokens.COMMENT
	}
	return tokens.DIVIDE
}

func (l *lexer) handleDash() tokens.TokenType {
	if l.peekChar() == '-' {
		l.readChar() // consume second '-'
		return l.readCustomProperty()
	} else if isDigit(l.peekChar()) {
		l.readNumber()
		return tokens.NUMBER
	} else if isWhitespace(l.peekChar()) {
		return tokens.MINUS
	} else if isIdentStart(l.peekChar()) || l.peekChar() == '\\' {
		l.readChar() // consume next char
		l.readIdentifier()
		return tokens.IDENT
	}
	return tokens.MINUS
}

func (l *lexer) readString() {
	delimiter := l.ch
	for l.peekChar() != delimiter && l.peekChar() != 0 && l.peekChar() != '\n' {
		if l.peekChar() == '\\' {
			l.readChar() // consume '\'
			if l.peekChar() == delimiter {
				l.readChar() // consume escaped quote
			}
		}
		l.readChar()
	}
	if l.peekChar() == delimiter {
		l.readChar() // consume closing quote
	}
}

func (l *lexer) readNumber() {
	for isDigit(l.peekChar()) {
		l.readChar()
	}
	if l.peekChar() == '.' && isDigit(l.peekNextChar()) {
		l.readChar() // consume '.'
		for isDigit(l.peekChar()) {
			l.readChar()
		}
	}
}

func (l *lexer) readIdentifier() {
	if l.ch == '\\' {
		l.readEscapedChar()
	}
	for isIdentPart(l.peekChar()) || l.peekChar() == '-' || l.peekChar() == '\\' {
		if l.peekChar() == '\\' {
			l.readChar() // consume '\'
			l.readEscapedChar()
		} else {
			l.readChar()
		}
	}
}

func (l *lexer) readEscapedChar() {
	if isHexDigit(l.peekChar()) {
		hexChars := 0
		for isHexDigit(l.peekChar()) && hexChars < 6 {
			l.readChar()
			hexChars++
		}
		if l.peekChar() == ' ' {
			l.readChar()
		}
	} else if l.peekChar() != '\n' {
		l.readChar()
	}
}

func (l *lexer) readHashOrColor() tokens.TokenType {
	colorLength := 0
	start := l.position

	for isHexDigit(l.peekChar()) && colorLength < 6 {
		l.readChar()
		colorLength++
	}

	if (colorLength == 3 || colorLength == 6) &&
		(!isIdentPart(l.peekChar()) || l.peekChar() == 0) {
		return tokens.COLOR
	}

	// If it's not a valid color, treat it as a HASH
	l.position = start // Reset position to just after the '#'
	l.readPosition = start + 1
	l.ch = '#'
	return tokens.HASH
}

func (l *lexer) readCustomProperty() tokens.TokenType {
	for isIdentPart(l.peekChar()) || l.peekChar() == '-' {
		l.readChar()
	}
	return tokens.IDENT
}

func (l *lexer) readComment() {
	for {
		l.readChar()
		if l.ch == 0 { // EOF
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // consume '/'
			break
		}
	}
}

func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_' || ch >= 0x80
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isIdentStart(ch byte) bool {
	return isLetter(ch) || ch == '_' || ch >= 0x80
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch) || ch == '-'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f'
}
