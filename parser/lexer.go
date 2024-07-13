package parser

import (
	"bytes"
	"io"
	"log"
)

const (
	bufferSize  = 4096
	invalidRune = '\uFFFD' // Unicode replacement character
)

type lexerBuffer struct {
	buffer chan *lexer
	new    func() *lexer
}

func newLexerBuffer(size int) *lexerBuffer {
	return &lexerBuffer{
		buffer: make(chan *lexer, size),
		new: func() *lexer {
			return &lexer{
				input: make([]byte, 0, bufferSize),
			}
		},
	}
}

func (lb *lexerBuffer) Get() *lexer {
	select {
	case l := <-lb.buffer:
		return l
	default:
		return lb.new()
	}
}

func (lb *lexerBuffer) Put(l *lexer) {
	select {
	case lb.buffer <- l:
		// Lexer added back to the buffer
	default:
		// Buffer is full, lexer is discarded
	}
}

var globalLexerBuffer = newLexerBuffer(10)

type lexer struct {
	input              []byte
	position           int
	readPosition       int
	ch                 byte
	line               int
	column             int
	lastToken          TokenType
	braceLevel         int
	bracketLevel       int
	squareBracketLevel int
	inAtRule           bool
	assigning          bool
	logger             *log.Logger
}

func Read(input io.Reader) *lexer {
	l := globalLexerBuffer.Get()
	l.reset(input)
	return l
}

func (l *lexer) reset(input io.Reader) {
	l.input = l.input[:0] // Clear the slice while keeping the capacity
	buf := bytes.NewBuffer(l.input)
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
	l.lastToken = TokenType("")
	l.braceLevel = 0
	l.bracketLevel = 0
	l.squareBracketLevel = 0
	l.inAtRule = false
	l.assigning = false
	l.readChar()
}

func (l *lexer) Release() {
	globalLexerBuffer.Put(l)
}

func (l *lexer) NextToken() Token {
	l.skipWhitespace()
	tok := Token{
		Line:   l.line,
		Column: l.column,
	}
	start := l.position

	if l.ch == 0 {
		tok.Type = EOF
		tok.Literal = []byte{}
		return tok
	}

	switch l.ch {
	case ';':
		tok.Type = SEMICOLON
	case ',':
		tok.Type = COMMA
	case '(':
		l.braceLevel++
		tok.Type = LPAREN
	case ')':
		l.braceLevel--
		tok.Type = RPAREN
	case '{':
		l.braceLevel++
		tok.Type = LBRACE
	case '}':
		l.braceLevel--
		tok.Type = RBRACE
	case '[':
		l.squareBracketLevel++
		tok.Type = LBRACKET
	case ']':
		l.squareBracketLevel--
		tok.Type = RBRACKET
	case '=':
		tok.Type = EQUALS
	case '+':
		tok.Type = PLUS
	case '>':
		tok.Type = GREATER
	case '~':
		tok.Type = TILDE
	case '|':
		tok.Type = PIPE
	case '^':
		tok.Type = CARET
	case '%':
		tok.Type = PERCENTAGE
	case '$':
		tok.Type = DOLLAR
	case '!':
		tok.Type = EXCLAMATION
	case '@':
		tok.Type = AT_RULE
		l.handleAt()
	case '*':
		tok.Type = SELECTOR
		l.handleUniversalSelector()
	case '/':
		tok.Type = l.handleSlash()
	case '"', '\'':
		tok.Type = STRING
		l.readString()
	case ':':
		tok.Type = l.handleColon()
	case '#':
		tok.Type = l.handleHash()
	case '.':
		tok.Type = l.handleDot()
	case '-':
		tok.Type = l.handleDash()
	case 0:
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			l.readIdentifier()
			if l.braceLevel == 0 && l.squareBracketLevel == 0 && l.input[start] != '-' && !l.inAtRule {
				tok.Type = SELECTOR
			} else {
				tok.Type = IDENT
			}
		} else if isDigit(l.ch) {
			tok.Type = NUMBER
			l.readNumber()
		} else {
			tok.Type = ILLEGAL
		}
	}

	if l.ch != 0 {
		l.readChar()
	}

	end := l.position
	tok.Literal = l.getLiteral(start, end)
	l.lastToken = tok.Type


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

func (l *lexer) handleAt() TokenType {
	for isIdentPart(l.peekChar()) {
		l.readChar()
	}
	l.inAtRule = true
	return AT_RULE
}

func (l *lexer) handleSlash() TokenType {
	if l.peekChar() == '*' {
		l.readChar() // consume '*'
		l.readComment()
		return COMMENT
	}
	return DIVIDE
}

func (l *lexer) handleColon() TokenType {
	if l.lastToken == IDENT {
		l.assigning = true
		return COLON
	}
	if l.peekChar() == ':' {
		l.readChar() // consume second ':'
		if isIdentStart(l.peekChar()) {
			l.readChar()
			l.readIdentifier()
			return SELECTOR
		}
	} else if isIdentStart(l.peekChar()) {
		l.readChar()
		l.readIdentifier()
		return SELECTOR
	}
	return COLON
}

func (l *lexer) handleHash() TokenType {
	if isIdentStart(l.peekChar()) || isDigit(l.peekChar()) {
		return l.readHashOrColor()
	} else if l.peekChar() == '\\' {
		l.readChar() // consume '\'
		l.readEscapedChar()
		l.readIdentifier()
		return SELECTOR
	}
	return ILLEGAL
}

func (l *lexer) handleDot() TokenType {
	if isIdentStart(l.peekChar()) {
		return l.readClassSelector()
	} else if l.peekChar() == '\\' {
		l.readChar() // consume '\'
		l.readEscapedChar()
		l.readIdentifier()
		return SELECTOR
	} else if isDigit(l.peekChar()) {
		l.readNumber()
		return NUMBER
	}
	return DOT
}

func (l *lexer) handleDash() TokenType {
	if l.peekChar() == '-' {
		l.readChar() // consume second '-'
		return l.readCustomProperty()
	} else if isDigit(l.peekChar()) {
		l.readNumber()
		return NUMBER
	} else if isWhitespace(l.peekChar()) && l.lastToken == NUMBER {
		return MINUS
	} else if isIdentStart(l.peekChar()) || l.peekChar() == '\\' {
		l.readChar() // consume next char
		l.readIdentifier()
		return IDENT
	}
	return MINUS
}

func (l *lexer) handleUniversalSelector() TokenType {
    for isIdentPart(l.peekChar()) || l.peekChar() == '.' || l.peekChar() == '#' {
        l.readChar()
    }

    // Check for attribute selector after the identifier
    if l.peekChar() == '[' {
        l.readAttributeSelector()
    }
    return SELECTOR
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
    for isIdentPart(l.peekChar()) || l.peekChar() == '-' || l.peekChar() == '\\' {
        if l.peekChar() == '\\' {
            l.readChar() // consume '\'
            l.readEscapedChar()
        } else {
            l.readChar()
        }
    }
    // Check for attribute selector after the identifier
    if l.peekChar() == '[' {
        l.readAttributeSelector()
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

func (l *lexer) readAttributeSelector() {
    bracketDepth := 0
    for {
        if l.peekChar() == '[' {
            bracketDepth++
        } else if l.peekChar() == ']' {
            bracketDepth--
            if bracketDepth == 0 {
                l.readChar() // consume the closing ']'
                break
            }
        } else if l.peekChar() == 0 { // EOF
            break
        }
        l.readChar()
    }
}

func (l *lexer) readCustomProperty() TokenType {
    for isIdentPart(l.peekChar()) || l.peekChar() == '-' {
        l.readChar()
    }
    return IDENT
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

func (l *lexer) readHashOrColor() TokenType {
    colorLength := 1
    l.readChar() // consume first char after '#'
    for isHexDigit(l.peekChar()) && colorLength < 7 {
        l.readChar()
        colorLength++
    }

    if (colorLength == 3 || colorLength == 6) &&
        (!isIdentPart(l.peekChar()) || l.peekChar() == 0) {
        return COLOR
    } else {
        for isIdentPart(l.peekChar()) || l.peekChar() == '-' || l.peekChar() == ':' {
            if l.peekChar() == '\\' {
                l.readChar() // consume '\'
                l.readEscapedChar()
            } else {
                l.readChar()
            }
        }
        return SELECTOR
    }
}

func (l *lexer) readClassSelector() TokenType {
    for {
        if isIdentPart(l.peekChar()) || l.peekChar() == '-' || l.peekChar() == ':' {
            l.readChar()
        } else if l.peekChar() == '\\' {
            l.readChar() // consume '\'
            l.readEscapedChar()
        } else {
            break
        }
    }
    return SELECTOR
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
