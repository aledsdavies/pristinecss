package parser

import (
	"bufio"
	"io"
	"log"
)

const (
	important           = "important"
	readAheadBufferSize = 64 // Must be a power of 2
	bufferMask          = readAheadBufferSize - 1
	invalidRune         = '\uFFFD' // Unicode replacement character
)

type Lexer struct {
	input              *bufio.Reader
	buffer             *CircularBuffer
	ch                 rune
	position           int
	readPosition       int
	line               int
	column             int
	lastToken          TokenType
	braceLevel         int
	bracketLevel       int
	squareBracketLevel int
	inAtRule           bool
	assigning          bool
	tokenBuffer        []*Token
	currentToken       int
	logger             *log.Logger
}

type CircularBuffer struct {
	buf   []rune
	start int
	end   int
}

func NewCircularBuffer() *CircularBuffer {
	return &CircularBuffer{
		buf: make([]rune, readAheadBufferSize),
	}
}

func (cb *CircularBuffer) Push(r rune) {
	cb.buf[cb.end] = r
	cb.end = (cb.end + 1) & bufferMask
	if cb.end == cb.start {
		cb.start = (cb.start + 1) & bufferMask
	}
}

func (cb *CircularBuffer) Pop() rune {
	if cb.start == cb.end {
		return invalidRune
	}
	r := cb.buf[cb.start]
	cb.buf[cb.start] = invalidRune
	cb.start = (cb.start + 1) & bufferMask
	return r
}

func (cb *CircularBuffer) Peek(n int) rune {
	if n <= 0 || n > cb.Size() {
		return invalidRune
	}
	return cb.buf[(cb.start+n-1)&bufferMask]
}

func (cb *CircularBuffer) Size() int {
	return (cb.end - cb.start) & bufferMask
}

func NewLexer(input io.Reader) *Lexer {
	l := &Lexer{
		input:       bufio.NewReader(input),
		buffer:      NewCircularBuffer(),
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

func (l *Lexer) readChar() {
	if l.buffer.Size() > 0 {
		l.ch = l.buffer.Pop()
	} else {
		var err error
		l.ch, _, err = l.input.ReadRune()
		if err != nil {
			l.ch = 0
		}
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
	if l.buffer.Size() > 0 {
		return l.buffer.Peek(1)
	}
	ch, _, _ := l.input.ReadRune() // Ignore error for speed
	if ch != 0 {
		l.buffer.Push(ch)
	}
	return ch
}

func (l *Lexer) peekCharN(n int) rune {
	currentSize := l.buffer.Size()
	if n <= currentSize {
		return l.buffer.Peek(n)
	}

	// Read additional characters into the buffer
	for i := 0; i < n-currentSize; i++ {
		ch, _, err := l.input.ReadRune()
		if err != nil {
			break
		}
		l.buffer.Push(ch)
	}

	if n <= l.buffer.Size() {
		return l.buffer.Peek(n)
	}
	return 0 // Not enough characters available
}

func (l *Lexer) peekString(s string, skipPrefixWhitespace bool) (int, bool) {
	index := 0

	if skipPrefixWhitespace {
		// Skip leading whitespace
		for {
			ch := l.peekCharN(index + 1)
			if ch == 0 || !isWhitespace(ch) {
				break
			}
			index++
		}
	}

	for _, expectedCh := range s {
		ch := l.peekCharN(index + 1)
		if ch != expectedCh {
			return index, false
		}
		index++
	}

	return index, true
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
		tok.Type = SEMICOLON
		tok.AppendLiteral(l.ch)
	case ',':
		tok.Type = COMMA
		tok.AppendLiteral(l.ch)
	case '(':
		l.braceLevel++
		l.inAtRule = false
		tok.Type = LPAREN
		tok.AppendLiteral(l.ch)
	case ')':
		l.braceLevel--
		tok.Type = RPAREN
		tok.AppendLiteral(l.ch)
	case '{':
		l.braceLevel++
		l.inAtRule = false
		tok.Type = LBRACE
		tok.AppendLiteral(l.ch)
	case '}':
		l.braceLevel--
		tok.Type = RBRACE
		tok.AppendLiteral(l.ch)
	case '[':
		l.squareBracketLevel++
		tok.Type = LBRACKET
		tok.AppendLiteral(l.ch)
	case ']':
		l.squareBracketLevel--
		tok.Type = RBRACKET
		tok.AppendLiteral(l.ch)
	case '=':
		tok.Type = EQUALS
		tok.AppendLiteral(l.ch)
	case '@':
		l.handleAt(tok)
	case '*':
		l.handleUniversalSelector(tok)
	case '+':
		tok.Type = PLUS
		tok.AppendLiteral(l.ch)
	case '>':
		tok.Type = GREATER
		tok.AppendLiteral(l.ch)
	case '~':
		tok.Type = TILDE
		tok.AppendLiteral(l.ch)
	case '|':
		tok.Type = PIPE
		tok.AppendLiteral(l.ch)
	case '^':
		tok.Type = CARET
		tok.AppendLiteral(l.ch)
	case '%':
		l.handlePercent(tok)
	case '$':
		tok.Type = DOLLAR
		tok.AppendLiteral(l.ch)
	case '/':
		l.handleSlash(tok)
	case '"', '\'':
		l.readString(tok)
	case '!':
		l.handleImportant(tok)
	case ':':
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
		if l.inAtRule && (isLetter(l.ch) || l.ch == '-') {
			l.readIdentifier(tok)
			tok.Type = IDENT
		} else if isLetter(l.ch) {
			l.readIdentifier(tok)
		} else if isDigit(l.ch) {
			l.readNumber(tok)
		} else {
			tok.Type = ILLEGAL
			tok.AppendLiteral(l.ch)
		}
	}

	// Ensure we always advance, except for EOF
	if l.ch != 0 {
		l.readChar()
	}
	l.lastToken = tok.Type

	return *tok
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

func (l *Lexer) handleAt(tok *Token) {
	tok.Type = AT_RULE
	tok.AppendLiteral('@')

	// Read the @-rule identifier
	for isIdentPart(l.peekChar()) {
		l.readChar()
		tok.AppendLiteral(l.ch)
	}

	// Set a flag to indicate we're in an @-rule context
	l.inAtRule = true
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
	peekedLength, match := l.peekString(important, true)
	if match {
		tok.Type = IMPORTANT
		for i := 0; i < peekedLength; i++ {
			l.readChar()
			tok.AppendLiteral(l.ch)
		}
	} else {
		tok.Type = ILLEGAL
	}
}

func (l *Lexer) handleColon(tok *Token) {
	tok.AppendLiteral(l.ch)

	if l.lastToken == IDENT {
		l.assigning = true
		tok.Type = COLON
		return
	}

	ch := l.peekChar()
	if ch == ':' {
        l.readChar()
        tok.AppendLiteral(l.ch)
        l.readChar()
		l.readIdentifier(tok)
		tok.Type = PSEUDO_ELEMENT
	} else if isIdentStart(ch) {
        l.readChar()
		l.readIdentifier(tok)
		tok.Type = PSEUDO_CLASS
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
	if l.peekChar() == '-' {
		l.readChar()
		l.readCustomProperty(tok)
	} else if isDigit(l.peekChar()) {
		l.readChar()
		l.readNumber(tok)
	} else if isWhitespace(l.peekChar()) && l.lastToken == NUMBER {
		tok.Type = MINUS
	} else if isIdentStart(l.peekChar()) || l.peekChar() == '\\' {
		l.readChar()
		l.readIdentifier(tok)
	} else {
		tok.Type = MINUS
	}
}

func (l *Lexer) handleUniversalSelector(tok *Token) {
	tok.Type = SELECTOR
	tok.AppendLiteral('*')

	// Check if it's part of a more complex selector
	for isIdentPart(l.peekChar()) || l.peekChar() == '.' || l.peekChar() == '#' || l.peekChar() == '[' {
		l.readChar()
		tok.AppendLiteral(l.ch)

		if l.ch == '[' {
			// Handle attribute selector
			l.readAttributeSelector(tok)
		}
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
		if isIdentPart(next) || next == '-' {
			l.readChar()
			tok.AppendLiteral(l.ch)
		} else if next == '\\' {
			l.readChar() // consume backslash
			l.readEscapedChar(tok)
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

func (l *Lexer) readFunctionalPseudoClass(tok *Token) {
	parenthesesCount := 1
	l.readChar() // consume opening parenthesis
	tok.AppendLiteral(l.ch)
	for parenthesesCount > 0 {
		l.readChar()
		if l.ch == 0 { // EOF
			break
		}
		tok.AppendLiteral(l.ch)
		switch l.ch {
		case '(':
			parenthesesCount++
		case ')':
			parenthesesCount--
		}
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
