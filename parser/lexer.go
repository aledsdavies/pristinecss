package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Lexer struct {
	input        *bufio.Reader
	ch           rune
	peeked       bool
	peekedChar   rune
	lastToken    TokenType
	position     int
	readPosition int
	line         int
	column       int

	// Info to help with debuging
	lastLine   int
	lastColumn int
	lastChar   rune
	tokenCount int
	debugInfo  strings.Builder
}

func NewLexer(input io.Reader) *Lexer {
	l := &Lexer{input: bufio.NewReader(input), line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.peeked {
		l.ch = l.peekedChar
		l.peeked = false
	} else {
		ch, _, err := l.input.ReadRune()
		if err != nil {
			l.ch = 0 // ASCII code for 'NUL'
		} else {
			l.ch = ch
		}
	}

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}

	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peekChar() rune {
	if l.peeked {
		return l.peekedChar
	}

	ch, _, err := l.input.ReadRune()
	if err != nil {
		l.peekedChar = 0 // ASCII code for 'NUL'
	} else {
		l.peekedChar = ch
	}

	l.peeked = true
	return l.peekedChar
}

func (l *Lexer) peekString(s string) bool {
	pos := l.input.Buffered()
	buf, err := l.input.Peek(len(s))
	if err != nil {
		return false
	}
	if string(buf) == s {
		return true
	}
	l.input.Discard(pos)
	return false
}

func (l *Lexer) readComment() string {
	var comment strings.Builder
	l.readChar()
	l.readChar()

	for {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			break
		}
		if l.ch == 0 {
			break
		}
		comment.WriteRune(l.ch)
		l.readChar()
	}

	return comment.String()
}

func (l *Lexer) NextToken() Token {
	var tok Token
	l.skipWhitespace()
	line := l.line
	column := l.column

	// Safety check
	if line == l.lastLine && column == l.lastColumn && l.ch == l.lastChar {
		l.debugInfo.WriteString(fmt.Sprintf("FATAL: Lexer stuck at line %d, column %d, character '%c'\n", line, column, l.ch))
		l.debugInfo.WriteString(fmt.Sprintf("Total tokens processed: %d\n", l.tokenCount))
		l.debugInfo.WriteString(fmt.Sprintf("Current lexer state: %+v\n", l))
		panic(l.debugInfo.String())
	}

	l.lastLine = line
	l.lastColumn = column
	l.lastChar = l.ch
	l.tokenCount++

	switch l.ch {
	case ':':
		tok = newToken(COLON, l.ch, line, column)
	case ';':
		tok = newToken(SEMICOLON, l.ch, line, column)
	case ',':
		tok = newToken(COMMA, l.ch, line, column)
	case '(':
		tok = newToken(LPAREN, l.ch, line, column)
	case ')':
		tok = newToken(RPAREN, l.ch, line, column)
	case '{':
		tok = newToken(LBRACE, l.ch, line, column)
	case '}':
		tok = newToken(RBRACE, l.ch, line, column)
	case '.':
		tok = newToken(DOT, l.ch, line, column)
	case '#':
		if isHexDigit(l.peekChar()) {
			tok = l.readHexColor()
			return tok
		} else {
			tok = newToken(HASH, l.ch, line, column)
		}
	case '&':
		tok = newToken(AMPERSAND, l.ch, line, column)
	case '[':
		tok = newToken(LBRACKET, l.ch, line, column)
	case ']':
		tok = newToken(RBRACKET, l.ch, line, column)
	case '=':
		tok = newToken(EQUALS, l.ch, line, column)
	case '@':
		tok = newToken(AT, l.ch, line, column)
	case '*':
		tok = newToken(ASTERISK, l.ch, line, column)
	case '+':
		tok = newToken(PLUS, l.ch, line, column)
	case '-':
		if l.peekChar() == '-' {
			// This is a custom property
			tok.Type = IDENT
			tok.Literal = l.readCustomProperty()
			tok.Line = line
			tok.Column = column
			l.lastToken = tok.Type
			return tok
		} else if isDigit(l.peekChar()) {
			// If the next character is a digit, this is a negative number
			tok.Type = NUMBER
			tok.Literal = l.readNumber()
			tok.Line = line
			tok.Column = column
		} else if isLetter(l.peekChar()) || l.peekChar() == '_' {
			// This is an identifier starting with a hyphen
			tok.Type = IDENT
			tok.Literal = l.readIdentifier()
		} else {
			// Otherwise, it's the minus operator
			tok = newToken(MINUS, l.ch, line, column)
		}
	case '>':
		tok = newToken(GREATER, l.ch, line, column)
	case '~':
		tok = newToken(TILDE, l.ch, line, column)
	case '|':
		tok = newToken(PIPE, l.ch, line, column)
	case '^':
		tok = newToken(CARET, l.ch, line, column)
	case '$':
		tok = newToken(DOLLAR, l.ch, line, column)
	case '%':
		if l.lastToken == NUMBER {
			tok = newToken(UNIT, l.ch, line, column)
		} else {
			tok = newToken(ILLEGAL, l.ch, line, column)
		}
	case '/':
		if l.peekChar() == '*' {
			tok.Type = COMMENT
			tok.Literal = l.readComment()
			tok.Line = line
			tok.Column = column
			l.lastToken = tok.Type
			return tok
		} else {
			tok = newToken(DIVIDE, l.ch, line, column)
		}
	case '"', '\'':
		tok.Type = STRING
		tok.Literal = l.readString(l.ch)
		tok.Line = line
		tok.Column = column
		l.lastToken = tok.Type
		return tok
	case '!':
		if l.peekString("important") {
			tok.Type = IMPORTANT
			tok.Literal = "!important"
			tok.Line = line
			tok.Column = column
			for range "important" {
				l.readChar()
			}
		} else {
			tok = newToken(ILLEGAL, l.ch, line, column)
		}
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		tok.Line = line
		tok.Column = column
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			if l.lastToken == NUMBER && isUnit(tok.Literal) {
				tok.Type = UNIT
			} else {
				tok.Type = IDENT
			}
			tok.Line = line
			tok.Column = column
			l.lastToken = tok.Type
			return tok
		} else if isDigit(l.ch) {
			tok.Type = NUMBER
			tok.Literal = l.readNumber()
			tok.Line = line
			tok.Column = column
			l.lastToken = tok.Type
			return tok
		} else {
			tok = newToken(ILLEGAL, l.ch, line, column)
		}
	}
	l.readChar()
	l.lastToken = tok.Type
	return tok
}

func newToken(tokenType TokenType, ch rune, line int, column int) Token {
	return Token{Type: tokenType, Literal: string(ch), Line: line, Column: column}
}

func (l *Lexer) readString(delimiter rune) string {
	var str strings.Builder
	l.readChar()

	for l.ch != delimiter {
		if l.ch == 0 || l.ch == '\n' {
			break
		}

		if l.ch == '\\' && l.peekChar() == delimiter {
			l.readChar()
			str.WriteRune(l.ch)
			l.readChar()
			continue
		}

		str.WriteRune(l.ch)
		l.readChar()
	}

	l.readChar()
	return str.String()
}

func (l *Lexer) readIdentifier() string {
    var ident strings.Builder
    ident.WriteRune(l.ch) // Write the first character (which could be '-')
    l.readChar() // Move to the next character

    // Read subsequent characters
    for isIdentPart(l.ch) || l.ch == '\\' {
        if l.ch == '\\' {
            l.readEscapedOrRune(&ident)
        } else {
            ident.WriteRune(l.ch)
            l.readChar()
        }
    }

    return ident.String()
}

func (l *Lexer) readEscapedOrRune(builder *strings.Builder) {
	if l.ch == '\\' {
		l.readChar() // Consume backslash
		if isHexDigit(l.ch) {
			// Unicode escape sequence
			hexValue := l.readHexEscape()
			if r, err := strconv.ParseInt(hexValue, 16, 32); err == nil {
				builder.WriteRune(rune(r))
			} else {
				// Invalid hex escape, write as is
				builder.WriteRune('\\')
				builder.WriteString(hexValue)
			}
		} else if l.ch == '\n' {
			// Escaped newline, ignore
			l.readChar()
		} else {
			// Other escaped character
			builder.WriteRune(l.ch)
			l.readChar()
		}
	} else {
		builder.WriteRune(l.ch)
		l.readChar()
	}
}

func (l *Lexer) readHexEscape() string {
	var hexValue strings.Builder
	for i := 0; i < 6 && isHexDigit(l.ch); i++ {
		hexValue.WriteRune(l.ch)
		l.readChar()
	}
	// Consume one whitespace after hex digits if present
	if isWhitespace(l.ch) {
		l.readChar()
	}
	return hexValue.String()
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f'
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch >= 0x80
}

func isIdentPart(ch rune) bool {
	return isIdentStart(ch) || unicode.IsDigit(ch) || ch == '-'
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '-' ||
		(ch >= '\u0080' && ch != '\uFFFD' && !unicode.IsDigit(ch))
}

func (l *Lexer) readNumber() string {
	var num strings.Builder
	dotCount := 0

	for isDigit(l.ch) || l.ch == '.' {
		if l.ch == '.' {
			dotCount++
			if dotCount > 1 {
				break // More than one dot is not allowed in a number
			}
		}
		num.WriteRune(l.ch)
		l.readChar()
	}

	return num.String()
}

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

func (l *Lexer) readHexColor() Token {
	var color strings.Builder
	startLine, startColumn := l.line, l.column
	color.WriteRune('#')
	l.readChar() // consume the '#'

	digitCount := 0
	for isHexDigit(l.ch) {
		color.WriteRune(l.ch)
		l.readChar()
		digitCount++
		if digitCount == 8 {
			break
		}
	}

	if digitCount == 3 || digitCount == 6 || digitCount == 4 || digitCount == 8 {
		return Token{Type: COLOR, Literal: color.String(), Line: startLine, Column: startColumn}
	} else {
		return Token{Type: ILLEGAL, Literal: color.String(), Line: startLine, Column: startColumn}
	}
}

func isHexDigit(ch rune) bool {
	return unicode.IsDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func (l *Lexer) readCustomProperty() string {
	var prop strings.Builder
	// We're already on the first '-', so start by writing it
	prop.WriteRune(l.ch)
	l.readChar() // Move to the second '-'

	// Write the second '-'
	if l.ch == '-' {
		prop.WriteRune(l.ch)
		l.readChar() // Move to the first character after '--'
	} else {
		// If there's no second '-', this isn't a custom property
		return prop.String()
	}

	// Read the rest of the custom property name
	for isIdentPart(l.ch) || l.ch == '-' {
		prop.WriteRune(l.ch)
		l.readChar()
	}

	return prop.String()
}
