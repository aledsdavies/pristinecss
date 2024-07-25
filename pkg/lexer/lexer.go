package lexer

import (
	"bytes"
	"io"
	"log"

	"github.com/aledsdavies/pristinecss/pkg/tokens"
	mempool "github.com/aledsdavies/pristinecss/pkg/utils"
)

const (
	EOF         = 0
	SEMICOLON   = ';'
	COMMA       = ','
	LPAREN      = '('
	RPAREN      = ')'
	LBRACE      = '{'
	RBRACE      = '}'
	LBRACKET    = '['
	RBRACKET    = ']'
	EQUALS      = '='
	PLUS        = '+'
	GREATER     = '>'
	TILDE       = '~'
	PIPE        = '|'
	CARET       = '^'
	AMPERSAND   = '&'
	PERCENTAGE  = '%'
	DOLLAR      = '$'
	EXCLAMATION = '!'
	AT          = '@'
	ASTERISK    = '*'
	COLON       = ':'
	DOT         = '.'
	HASH        = '#'
	DASH        = '-'
	BACKSLASH   = '\\'
	SLASH       = '/'
	DOUBLEQUOTE = '"'
	SINGLEQUOTE = '\''
)

var (
	isWhitespace [256]bool
	isLetter     [256]bool
	isDigit      [256]bool
	isIdentStart [256]bool
	isIdentPart  [256]bool
	isHexDigit   [256]bool
)

func init() {
	for i := 0; i < 256; i++ {
		ch := byte(i)
		isWhitespace[i] = ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f'
		isLetter[i] = ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_' || ch >= 0x80
		isDigit[i] = '0' <= ch && ch <= '9'
		isIdentStart[i] = isLetter[i] || ch == '_' || ch >= 0x80
		isIdentPart[i] = isIdentStart[i] || isDigit[i] || ch == '-'
		isHexDigit[i] = isDigit[i] || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
	}
}

var tokenPool *mempool.Pool[*tokens.Token]
var lexerPool *mempool.Pool[*lexer]
var bufferPool *mempool.Pool[*PoolByteBuffer]

func init() {
	tokenPool = mempool.NewPool(
		func() *tokens.Token { return &tokens.Token{} },
		mempool.WithCapacity(1),
	)
	lexerPool = mempool.NewPool(
		func() *lexer {
			l := &lexer{}
			l.Erase()
			return l
		},
		mempool.WithCapacity(10),
	)
	bufferPool = mempool.NewPool(
		func() *PoolByteBuffer { return &PoolByteBuffer{Buffer: bytes.Buffer{}} },
		mempool.WithCapacity(10),
	)
}

type PoolByteBuffer struct {
	bytes.Buffer
}

func (pbb *PoolByteBuffer) Erase() {
	pbb.Reset()
}

// estimateTokenCount estimates the number of tokens based on input size
func estimateTokenCount(inputSize int) int {
	// This is a rough estimate and may need tuning based on your specific CSS patterns
	return inputSize / 4
}

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
	l := lexerPool.Get()
	defer lexerPool.Put(l)

	buf := bufferPool.Get()
	defer bufferPool.Put(buf)

	_, err := io.Copy(&buf.Buffer, input)
	if err != nil {
		log.Fatalf("Fatal error reading input: %v", err)
	}
	l.input = buf.Bytes()
	l.readChar()

	return l.tokenize()
}

func (l *lexer) Erase() {
	l.input = l.input[:0]
	l.position = 0
	l.readPosition = 0
	l.ch = 0
	l.line = 1
	l.column = 0
}

func (l *lexer) tokenize() []tokens.Token {
	estimatedTokens := estimateTokenCount(len(l.input))
	result := make([]tokens.Token, 0, estimatedTokens)
	for {
		tok := l.nextToken()
		result = append(result, *tok)
		if tok.Type == tokens.EOF {
			tokenPool.Put(tok)
			break
		}
		tokenPool.Put(tok)
	}

	return result
}

func (l *lexer) nextToken() *tokens.Token {
	l.skipWhitespace()
	tok := tokenPool.Get()
	tok.Line = l.line
	tok.Column = l.column
	start := l.position

	if l.ch == EOF {
		tok.Type = tokens.EOF
		tok.Literal = []byte{}
		return tok
	}

	switch l.ch {
	case SEMICOLON:
		tok.Type = tokens.SEMICOLON
	case COMMA:
		tok.Type = tokens.COMMA
	case AMPERSAND:
		tok.Type = tokens.AMPERSAND
	case LPAREN:
		tok.Type = tokens.LPAREN
	case RPAREN:
		tok.Type = tokens.RPAREN
	case LBRACE:
		tok.Type = tokens.LBRACE
	case RBRACE:
		tok.Type = tokens.RBRACE
	case LBRACKET:
		tok.Type = tokens.LBRACKET
	case RBRACKET:
		tok.Type = tokens.RBRACKET
	case EQUALS:
		tok.Type = tokens.EQUALS
	case PLUS:
		tok.Type = tokens.PLUS
	case GREATER:
		tok.Type = tokens.GREATER
	case TILDE:
		tok.Type = tokens.TILDE
	case PIPE:
		tok.Type = tokens.PIPE
	case CARET:
		if l.peekChar() == EQUALS {
			l.readChar()
			tok.Type = tokens.STARTS_WITH
		} else {
			tok.Type = tokens.ILLEGAL
		}
	case PERCENTAGE:
		tok.Type = tokens.PERCENTAGE
	case DOLLAR:
		tok.Type = tokens.DOLLAR
	case EXCLAMATION:
		tok.Type = tokens.EXCLAMATION
	case AT:
		tok.Type = tokens.AT
	case ASTERISK:
		tok.Type = tokens.ASTERISK
	case COLON:
		if l.peekChar() == COLON {
			l.readChar()
			tok.Type = tokens.DBLCOLON
		} else {
			tok.Type = tokens.COLON
		}
	case DOT:
		tok.Type = tokens.DOT
	case HASH:
		tok.Type = l.readHashOrColor()
	case DASH:
		tok.Type = l.handleDash()
	case BACKSLASH:
		tok.Type = tokens.IDENT
		l.readIdentifier()
	case SLASH:
		tok.Type = l.handleSlash()
	case DOUBLEQUOTE, SINGLEQUOTE:
		tok.Type = tokens.STRING
		l.readString()
	default:
		if isLetter[l.ch] {
			tok.Type = tokens.IDENT
			l.readIdentifier()
		} else if isDigit[l.ch] {
			tok.Type = tokens.NUMBER
			l.readNumber()
		} else {
			tok.Type = tokens.ILLEGAL
		}
	}

	if l.ch != EOF {
		l.readChar()
	}

	end := l.position
	tok.Literal = l.getLiteral(start, end)

	return tok
}

func (l *lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = EOF
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
	} else if isDigit[l.peekChar()] {
		l.readNumber()
		return tokens.NUMBER
	} else if isWhitespace[l.peekChar()] {
		return tokens.MINUS
	} else if isIdentStart[l.peekChar()] || l.peekChar() == '\\' {
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
	for isDigit[l.peekChar()] {
		l.readChar()
	}
	if l.peekChar() == '.' && isDigit[l.peekNextChar()] {
		l.readChar() // consume '.'
		for isDigit[l.peekChar()] {
			l.readChar()
		}
	}
}

func (l *lexer) readIdentifier() {
	if l.ch == '\\' {
		l.readEscapedChar()
	}
	for isIdentPart[l.peekChar()] || l.peekChar() == '-' || l.peekChar() == '\\' {
		if l.peekChar() == '\\' {
			l.readChar() // consume '\'
			l.readEscapedChar()
		} else {
			l.readChar()
		}
	}
}

func (l *lexer) readEscapedChar() {
	if isHexDigit[l.peekChar()] {
		hexChars := 0
		for isHexDigit[l.peekChar()] && hexChars < 6 {
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

	for isHexDigit[l.peekChar()] && colorLength < 6 {
		l.readChar()
		colorLength++
	}

	if (colorLength == 3 || colorLength == 6) &&
		(!isIdentPart[l.peekChar()] || l.peekChar() == 0) {
		return tokens.COLOR
	}

	// If it's not a valid color, treat it as a HASH
	l.position = start // Reset position to just after the '#'
	l.readPosition = start + 1
	l.ch = '#'
	return tokens.HASH
}

func (l *lexer) readCustomProperty() tokens.TokenType {
	for isIdentPart[l.peekChar()] || l.peekChar() == '-' {
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

func (l *lexer) skipWhitespace() {
	for isWhitespace[l.ch] {
		l.readChar()
	}
}
