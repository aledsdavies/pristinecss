package parser

import (
	"bufio"
	"io"
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
		tokenBuffer: make([]*Token, 1),
	}

	for i := range l.tokenBuffer {
		l.tokenBuffer[i] = NewToken()
	}

	l.readChar()
	return l
}

func (l *Lexer) getNextToken() *Token {
	if l.currentToken >= len(l.tokenBuffer) {
		l.currentToken = 0
	}
	tok := l.tokenBuffer[l.currentToken]
	tok.Reset()
	l.currentToken++
	return tok
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	tok := l.getNextToken()
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case ';':
		l.assigning = false
		l.updateToken(tok, SEMICOLON, l.ch)
	case ',':
		l.updateToken(tok, COMMA, l.ch)
	case '(':
		l.braceLevel++
		l.updateToken(tok, LPAREN, l.ch)
	case ')':
		l.braceLevel--
		l.updateToken(tok, RPAREN, l.ch)
	case '{':
		l.braceLevel++
		l.updateToken(tok, LBRACE, l.ch)
	case '}':
		l.braceLevel--
		l.updateToken(tok, RBRACE, l.ch)
	case '[':
		l.squareBracketLevel++
		l.updateToken(tok, LBRACKET, l.ch)
	case ']':
		l.squareBracketLevel--
		l.updateToken(tok, RBRACKET, l.ch)
	case '=':
		l.updateToken(tok, EQUALS, l.ch)
	case '@':
		l.updateToken(tok, AT, l.ch)
	case '*':
		l.updateToken(tok, ASTERISK, l.ch)
	case '+':
		l.updateToken(tok, PLUS, l.ch)
	case '>':
		l.updateToken(tok, GREATER, l.ch)
	case '~':
		l.updateToken(tok, TILDE, l.ch)
	case '|':
		l.updateToken(tok, PIPE, l.ch)
	case '^':
		l.updateToken(tok, CARET, l.ch)
	case '$':
		l.updateToken(tok, DOLLAR, l.ch)
	case '%':
		l.handlePercent(tok)
	case '/':
		l.handleSlash(tok)
	case '"', '\'':
		l.readString(tok)
	case '!':
		l.handleImportant(tok)
	case ':':
		if l.lastToken == IDENT {
			l.assigning = true
		}
		l.handleColon(tok)
	case '#':
		l.handleHash(tok)
	case '.':
		l.handleDot(tok)
	case '-':
		l.handleDash(tok)
	case 0:
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			l.readIdentifier(tok)
		} else if isDigit(l.ch) {
			l.readNumber(tok)
		} else {
			l.updateToken(tok, ILLEGAL, l.ch)
		}
	}

	if tok.Type != EOF {
		l.readChar()
	}
	l.lastToken = tok.Type
	return *tok
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

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) handlePercent(tok *Token) {
	tok.AppendLiteral('%')
	if l.lastToken == NUMBER {
		tok.Type = UNIT
	} else {
		tok.Type = ILLEGAL
	}
}

func (l *Lexer) handleSlash(tok *Token) {
	tok.AppendLiteral('/')
	if l.peekChar() == '*' {
		l.readComment(tok)
	} else {
		tok.Type = DIVIDE
	}
}

func (l *Lexer) handleImportant(tok *Token) {
	tok.AppendLiteral('!')
	if l.peekString("important") {
		tok.Type = IMPORTANT
		for range "important" {
			l.readChar()
			tok.AppendLiteral(l.ch)
		}
	} else {
		tok.Type = ILLEGAL
	}
}

func (l *Lexer) handleColon(tok *Token) {
	tok.AppendLiteral(l.ch)
	if isIdentStart(l.peekChar()) {
		l.readChar()
		l.readIdentifier(tok)
		tok.Type = SELECTOR
	} else {
		tok.Type = COLON
	}
}

func (l *Lexer) handleHash(tok *Token) {
    tok.AppendLiteral(l.ch)
    if isIdentStart(l.peekChar()) || isDigit(l.peekChar()) {
        l.readChar()
        l.readHashOrColor(tok)
    } else if l.peekChar() == '\\' {
        l.readChar()
        l.readEscapedChar(tok)
        l.readIdentifier(tok)
        tok.Type = SELECTOR
    } else {
        tok.Type = ILLEGAL
    }
}

func (l *Lexer) handleDot(tok *Token) {
	tok.AppendLiteral(l.ch)
	if isIdentStart(l.peekChar()) {
		l.readClassSelector(tok)
	} else if l.peekChar() == '\\' {
		l.readChar()
		l.readEscapedChar(tok)
        l.readIdentifier(tok)
	} else if isDigit(l.peekChar()) {
		l.readChar()
		l.readNumber(tok)
	} else {
		tok.Type = DOT
	}
}

func (l *Lexer) handleDash(tok *Token) {
	tok.AppendLiteral(l.ch)
	next := l.peekChar()
	if next == '-' {
		l.readChar()
		l.readCustomProperty(tok)
	} else if isDigit(next) {
		l.readChar()
		l.readNumber(tok)
	} else if isWhitespace(next) && l.lastToken == NUMBER {
		tok.Type = MINUS
	} else if isIdentStart(next) || next == '\\' {
		l.readChar()
		l.readIdentifier(tok)
	} else {
		tok.Type = MINUS
	}
}

func (l *Lexer) readString(tok *Token) {
	delimiter := l.ch
	tok.Type = STRING
	tok.AppendLiteral(delimiter)
	for l.peekChar() != delimiter && l.peekChar() != 0 && l.peekChar() != '\n' {
		l.readChar()
		if l.ch == '\\' && l.peekChar() == delimiter {
			tok.AppendLiteral(l.ch)
			l.readChar()
		}
		tok.AppendLiteral(l.ch)
	}
	if l.peekChar() == delimiter {
		l.readChar()
		tok.AppendLiteral(delimiter)
	}
}

func (l *Lexer) readNumber(tok *Token) {
	tok.Type = NUMBER
	tok.AppendLiteral(l.ch)
	for isDigit(l.peekChar()) {
		l.readChar()
		tok.AppendLiteral(l.ch)
	}
	if l.peekChar() == '.' && isDigit(l.peekCharN(2)) {
		l.readChar()
		tok.AppendLiteral(l.ch)
		for isDigit(l.peekChar()) {
			l.readChar()
			tok.AppendLiteral(l.ch)
		}
	}
}

func (l *Lexer) readIdentifier(tok *Token) {
	tok.AppendLiteral(l.ch)
	for {
		next := l.peekChar()
		if isIdentPart(next) || next == '-' || (next == ':' && l.braceLevel == 0 && l.bracketLevel == 0) {
			l.readChar()
			tok.AppendLiteral(l.ch)
		} else if next == '\\' {
			l.readChar() // consume backslash
			l.readEscapedChar(tok)
            l.readIdentifier(tok)
		} else if next == '[' {
			l.readChar()
			tok.AppendLiteral(next)
			l.readAttributeSelector(tok)
		} else {
			break
		}
	}

	if isUnit(tok.Literal) {
		tok.Type = UNIT
	} else if l.braceLevel == 0 && l.squareBracketLevel == 0 && tok.Literal[0] != '-' {
		tok.Type = SELECTOR
	} else {
		tok.Type = IDENT
	}
}

func (l *Lexer) readEscapedChar(tok *Token) {
	tok.AppendLiteral('\\')
	if isHexDigit(l.peekChar()) {
		// Handle hexadecimal escape
		hexChars := 0
		for isHexDigit(l.peekChar()) && hexChars < 6 {
			l.readChar()
			tok.AppendLiteral(l.ch)
			hexChars++
		}

		// Lets us continue on white space
		if l.peekChar() == ' ' {
			l.readChar()
		}
	} else if l.peekChar() != '\n' {
		l.readChar()
		// For any other escaped character, just append it
		tok.AppendLiteral(l.ch)
	}
}

func (l *Lexer) readAttributeSelector(tok *Token) {
	bracketDepth := 1
	for bracketDepth > 0 {
		l.readChar()
		if l.ch == 0 { // EOF
			break
		}
		if l.ch == '\\' {
			// Handle escaped character
			tok.AppendLiteral(l.ch)
			l.readChar()
			tok.AppendLiteral(l.ch)
		} else {
			tok.AppendLiteral(l.ch)
			if l.ch == '[' {
				bracketDepth++
			} else if l.ch == ']' {
				bracketDepth--
			}
		}
	}
}

func (l *Lexer) readCustomProperty(tok *Token) {
	tok.AppendLiteral('-')
	for isIdentPart(l.peekChar()) || l.peekChar() == '-' {
		l.readChar()
		tok.AppendLiteral(l.ch)
	}
	tok.Type = IDENT
}

func (l *Lexer) readComment(tok *Token) {
    tok.Type = COMMENT
    for {
        l.readChar()
        if l.ch == 0 { // EOF
            break
        }
        tok.AppendLiteral(l.ch)
        if l.ch == '*' && l.peekChar() == '/' {
            l.readChar()
            tok.AppendLiteral(l.ch)
            break
        }
    }
}

func (l *Lexer) readHashOrColor(tok *Token) {
	tok.AppendLiteral(l.ch)
	colorLength := 1
	for isHexDigit(l.peekChar()) && colorLength < 7 {
		l.readChar()
		tok.AppendLiteral(l.ch)
		colorLength++
	}

	if (colorLength == 3 || colorLength == 6) &&
		(!isIdentPart(l.peekChar()) || l.peekChar() == 0) {
		tok.Type = COLOR
	} else {
		tok.Type = SELECTOR
		for isIdentPart(l.peekChar()) || l.peekChar() == '-' || l.peekChar() == ':' {
			l.readChar()
			if l.ch == '\\' {
				tok.AppendLiteral(l.ch)
				l.readChar()
			}
			tok.AppendLiteral(l.ch)
		}
	}
}

func (l *Lexer) readClassSelector(tok *Token) {
	tok.Type = SELECTOR
	for {
		next := l.peekChar()
		if isIdentPart(next) || next == '-' || next == ':' {
			l.readChar()
			tok.AppendLiteral(l.ch)
		} else if next == '\\' {
			l.readChar() // consume backslash
			l.readEscapedChar(tok)
		} else {
			break
		}
	}
}

func (l *Lexer) updateToken(tok *Token, tokenType TokenType, ch rune) {
	tok.Type = tokenType
	tok.SetLiteral([]rune{ch})
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

func (l *Lexer) peekCharN(n int) rune {
	if n < 1 {
		return l.ch
	}
	chars := make([]rune, n)
	for i := 0; i < n; i++ {
		r, _, err := l.input.ReadRune()
		if err != nil {
			break
		}
		chars[i] = r
	}
	// Unread all the runes we just read
	for i := len(chars) - 1; i >= 0; i-- {
		if chars[i] != 0 {
			l.input.UnreadRune()
		}
	}
	if n-1 < len(chars) {
		return chars[n-1]
	}
	return 0
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && ch != 0
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isIdentStart(ch rune) bool {
	return isLetter(ch) || ch == '_' || ch >= 0x80
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
