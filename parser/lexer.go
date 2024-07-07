package parser

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Lexer struct {
	input              *bufio.Reader
	ch                 rune
	position           int
	readPosition       int
	line               int
	column             int
	lastToken          TokenType
	braceLevel         int
	bracketLevel       int
	squareBracketLevel int
	assigning          bool
	tokenBuffer        []*Token
	currentToken       int
}

func NewLexer(input io.Reader) *Lexer {
	l := &Lexer{
		input:       bufio.NewReader(input),
		line:        1,
		column:      0,
		tokenBuffer: make([]*Token, 100),
	}

	for i := range l.tokenBuffer {
		l.tokenBuffer[i] = &Token{}
	}

	l.readChar()
	return l
}

func (l *Lexer) getNextToken() *Token {
	if l.currentToken >= len(l.tokenBuffer) {
		l.currentToken = 0
	}
	tok := l.tokenBuffer[l.currentToken]
	l.currentToken++
	return tok
}

func (l *Lexer) readChar() {
	var err error
	l.ch, _, err = l.input.ReadRune()
	if err != nil {
		l.ch = 0
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

func (l *Lexer) peekChar() rune {
	ch, _, err := l.input.ReadRune()
	if err != nil {
		return 0
	}
	l.input.UnreadRune()
	return ch
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	tok := l.getNextToken()
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case ';':
		l.assigning = false
		l.newToken(tok, SEMICOLON, l.ch)
	case ',':
		l.newToken(tok, COMMA, l.ch)
	case '(':
		l.braceLevel++
		l.newToken(tok, LPAREN, l.ch)
	case ')':
		l.braceLevel--
		l.newToken(tok, RPAREN, l.ch)
	case '{':
		l.braceLevel++
		l.newToken(tok, LBRACE, l.ch)
	case '}':
		l.braceLevel--
		l.newToken(tok, RBRACE, l.ch)
	case '[':
		l.squareBracketLevel++
		l.newToken(tok, LBRACKET, l.ch)
	case ']':
		l.squareBracketLevel--
		l.newToken(tok, RBRACKET, l.ch)
	case '=':
		l.newToken(tok, EQUALS, l.ch)
	case '@':
		l.newToken(tok, AT, l.ch)
	case '*':
		l.newToken(tok, ASTERISK, l.ch)
	case '+':
		l.newToken(tok, PLUS, l.ch)
	case '>':
		l.newToken(tok, GREATER, l.ch)
	case '~':
		l.newToken(tok, TILDE, l.ch)
	case '|':
		l.newToken(tok, PIPE, l.ch)
	case '^':
		l.newToken(tok, CARET, l.ch)
	case '$':
		l.newToken(tok, DOLLAR, l.ch)
	case '%':
		tok.Type, tok.Literal = l.handlePercent()
	case '/':
		tok.Type, tok.Literal = l.handleSlash()
	case '"', '\'':
		tok.Type, tok.Literal = l.readString()
		l.lastToken = tok.Type
		return *tok
	case '!':
		tok.Type, tok.Literal = l.handleImportant()
		l.lastToken = tok.Type
	case ':':
		if l.lastToken == IDENT {
			l.assigning = true
		}

		tok.Type, tok.Literal = l.handleColon()
		l.lastToken = tok.Type
		return *tok
	case '#':
		tok.Type, tok.Literal = l.handleHash()
		l.lastToken = tok.Type
		return *tok
	case '.':
		tok.Type, tok.Literal = l.handleDot()
		l.lastToken = tok.Type
		return *tok
	case '-':
		tok.Type, tok.Literal = l.handleDash()
		l.lastToken = tok.Type
		return *tok
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Type, tok.Literal = l.readIdentifier()
			l.lastToken = tok.Type
			return *tok
		} else if isDigit(l.ch) {
			tok.Type, tok.Literal = l.readNumber()
			l.lastToken = tok.Type
			return *tok
		} else {
			l.newToken(tok, ILLEGAL, l.ch)
		}
	}

	l.readChar()
	l.lastToken = tok.Type
	return *tok
}

func (l *Lexer) handlePercent() (TokenType, string) {
	if l.lastToken == NUMBER {
		return UNIT, "%"
	}
	return ILLEGAL, "%"
}

func (l *Lexer) handleSlash() (TokenType, string) {
	if l.peekChar() == '*' {
		return l.readComment()
	}
	return DIVIDE, "/"
}

func (l *Lexer) handleImportant() (TokenType, string) {
	if l.peekString("important") {
		l.readChar() // consume '!'
		for range "important" {
			l.readChar()
		}
		return IMPORTANT, "!important"
	}
	return ILLEGAL, "!"
}

func (l *Lexer) handleColon() (TokenType, string) {
	l.readChar() // Consule the :
	if isIdentStart(l.ch) {
		var ident strings.Builder
		ident.WriteRune(':')
		for isIdentPart(l.ch) {
			ident.WriteRune(l.ch)
			l.readChar()
		}
		literal := ident.String()
		return SELECTOR, literal
	}

	return COLON, ":"
}

func (l *Lexer) handleHash() (TokenType, string) {
	l.readChar() // consume '#'
	if isIdentStart(l.ch) || isDigit(l.ch) {
		return l.readHashOrColor()
	}
	return ILLEGAL, "#"
}

func (l *Lexer) handleDot() (TokenType, string) {
	l.readChar() // consume '.'
	if isIdentStart(l.ch) {
		return l.readClassSelector()
	}
	if isDigit(l.ch) {
		return l.readNumber()
	}
	return DOT, "."
}

func (l *Lexer) handleDash() (TokenType, string) {
    l.readChar() // consume the dash

    next := l.ch
    if next == '-' {
        // Custom property (e.g., --custom-property)
        return l.readCustomProperty()
    } else if isDigit(next) {
        // Negative number
        return l.readNumber()
    } else if isWhitespace(next) && l.lastToken == NUMBER {
        // Likely an arithmetic operation (e.g., 10 - 5)
        return MINUS, "-"
    } else if isIdentStart(next) || next == '\\' {
        // Identifier starting with a dash
        return l.readIdentifier()
    }

    // Single dash
    return MINUS, "-"
}

func (l *Lexer) readString() (TokenType, string) {
	delimiter := l.ch
	var str strings.Builder
	l.readChar() // consume opening quote

	for l.ch != delimiter && l.ch != 0 && l.ch != '\n' {
		if l.ch == '\\' && l.peekChar() == delimiter {
			l.readChar()
		}
		str.WriteRune(l.ch)
		l.readChar()
	}

	if l.ch == delimiter {
		l.readChar() // consume closing quote
	}

	return STRING, str.String()
}

func (l *Lexer) readNumber() (TokenType, string) {
	var num strings.Builder
	for isDigit(l.ch) {
		num.WriteRune(l.ch)
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		num.WriteRune(l.ch)
		l.readChar()
		for isDigit(l.ch) {
			num.WriteRune(l.ch)
			l.readChar()
		}
	}
	return NUMBER, num.String()
}

func (l *Lexer) readIdentifier() (TokenType, string) {
	var ident strings.Builder
	for isIdentPart(l.ch) || l.ch == '-' || (l.ch == ':' && l.braceLevel == 0 && l.bracketLevel == 0) {
		ident.WriteRune(l.ch)
		l.readChar()
	}
	literal := ident.String()
	if isUnit(literal) {
		return UNIT, literal
	}

	if l.braceLevel == 0 && l.squareBracketLevel == 0 {
		return SELECTOR, literal
	}

	return IDENT, literal
}

func (l *Lexer) readCustomProperty() (TokenType, string) {
	var prop strings.Builder
	prop.WriteString("--")
	l.readChar() // move to first character after '--'

	for isIdentPart(l.ch) || l.ch == '-' {
		prop.WriteRune(l.ch)
		l.readChar()
	}

	return IDENT, prop.String()
}

func (l *Lexer) readComment() (TokenType, string) {
	var comment strings.Builder
	l.readChar() // consume '/'
	l.readChar() // consume '*'

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

	return COMMENT, comment.String()
}

func (l *Lexer) readHashOrColor() (TokenType, string) {
	var value strings.Builder
	value.WriteRune('#')

	colorLength := 0
	for isHexDigit(l.ch) && colorLength < 6 {
		value.WriteRune(l.ch)
		l.readChar()
		colorLength++
	}

	// If it's a valid hex color length and the next character isn't a valid identifier part,
	// or we've reached the end of the input, it's a color
	if (colorLength == 3 || colorLength == 6) &&
		(!isIdentPart(l.ch) || l.ch == 0) {
		return COLOR, value.String()
	}

	// If it's not a valid color, treat it as a selector
	// Reset the lexer to just after the '#'
	l.position -= colorLength
	l.readPosition = l.position + 1

	// Read the selector
	for isIdentPart(l.ch) || l.ch == '-' || l.ch == ':' {
		if l.ch == '\\' {
			value.WriteString(l.handleEscapedChar())
		} else {
			value.WriteRune(l.ch)
			l.readChar()
		}
		value.WriteRune(l.ch)
		l.readChar()
	}

	// If we've only read the '#', it's invalid
	if len(value.String()) == 1 {
		return ILLEGAL, value.String()
	}

	return SELECTOR, value.String()
}

func (l *Lexer) readClassSelector() (TokenType, string) {
	var selector strings.Builder
	selector.WriteRune('.')

	for isIdentPart(l.ch) || l.ch == '-' || l.ch == ':' {
		if l.ch == '\\' {
			selector.WriteString(l.handleEscapedChar())
		} else {
			selector.WriteRune(l.ch)
			l.readChar()
		}
	}

	return SELECTOR, selector.String()
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekString(s string) bool {
	for _, r := range s {
		if l.peekChar() != r {
			return false
		}
		l.readChar()
	}
	return true
}

func (l *Lexer) newToken(tok *Token, tokenType TokenType, ch rune) {
	tok.Type = tokenType
	tok.Literal = string(ch)
}

func (l *Lexer) handleEscapedChar() string {
	l.readChar() // Consume the backslash

	if l.ch == 0 {
		return "\\" // Return backslash if it's at the end of input
	}

	if l.ch == '\n' {
		return "" // Ignore escaped newline
	}

	if isHexDigit(l.ch) {
		// Handle unicode escape
		var hexString strings.Builder
		for i := 0; i < 6 && isHexDigit(l.ch); i++ {
			hexString.WriteRune(l.ch)
			l.readChar()
		}

		// Convert hex to rune
		if code, err := strconv.ParseInt(hexString.String(), 16, 32); err == nil {
			// Consume one whitespace after hex digits if present
			if unicode.IsSpace(l.ch) {
				l.readChar()
			}
			return string(rune(code))
		}
		// If parsing fails, return the original sequence
		return "\\" + hexString.String()
	}

	// For any other character, just return it
	char := "\\" + string(l.ch)
	l.readChar()
	return char
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && ch != 0
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isIdentStart(ch rune) bool {
	return isLetter(ch) || ch == '_' || ch >= 0x80 || ch == '\\'
}

func isIdentPart(ch rune) bool {
	return isIdentStart(ch) || isDigit(ch) || ch == '-'
}

func isHexDigit(ch rune) bool {
	return isDigit(ch) || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

func isWhitespace(ch rune) bool {
    return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f'
}
